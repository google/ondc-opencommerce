// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Server send messages to ONDC Buyer.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/benbjohnson/clock"
	log "github.com/golang/glog"
	"golang.org/x/sync/errgroup"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/keyclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/transactionclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/signing-authentication/authentication"
)

type server struct {
	pubsubClient      *pubsub.Client
	httpClient        *http.Client
	keyClient         keyClient
	transactionClient *transactionclient.Client
	config            config.CallbackActionConfig
	clk               clock.Clock

	subs []*pubsub.Subscription
}

type keyClient interface {
	ServiceSigningPrivateKeyset(context.Context) ([]byte, error)
}

func main() {
	flag.Set("alsologtostderr", "true")
	ctx := context.Background()

	configPath, ok := os.LookupEnv("CONFIG")
	if !ok {
		log.Exit("CONFIG env is not set")
	}

	conf, err := config.Read[config.CallbackActionConfig](configPath)
	if err != nil {
		log.Exit(err)
	}

	keyClient, err := keyclient.New(ctx, conf.ProjectID, conf.SecretID)
	if err != nil {
		log.Exit(err)
	}

	pubsubClient, err := pubsub.NewClient(ctx, conf.ProjectID)
	if err != nil {
		log.Exit(err)
	}

	transactionClient, err := transactionclient.New(ctx, conf.ProjectID, conf.InstanceID, conf.DatabaseID)
	if err != nil {
		log.Exit(err)
	}

	srv, err := initServer(ctx, http.DefaultClient, pubsubClient, keyClient, transactionClient, conf, clock.New())
	if err != nil {
		log.Exit(err)
	}
	defer srv.close()
	log.Info("Server initialization successs")

	if err := srv.serve(ctx); err != nil {
		log.Exitf("Serving failed: %v", err)
	}
}

func initServer(ctx context.Context, httpClient *http.Client, pubsubClient *pubsub.Client, keyClient keyClient, transactionClient *transactionclient.Client, conf config.CallbackActionConfig, clk clock.Clock) (*server, error) {
	// validate clients.
	if httpClient == nil {
		return nil, errors.New("init server: HTTP client is nil")
	}
	if pubsubClient == nil {
		return nil, errors.New("init server: Pub/Sub client is nil")
	}
	if keyClient == nil {
		return nil, errors.New("init server: key client is nil")
	}
	if transactionClient == nil {
		return nil, errors.New("init server: transaction client is nil")
	}

	// validate the callback topic
	callbackTopic := pubsubClient.Topic(conf.TopicID)
	ok, err := callbackTopic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("init server: topic %q does not exist", callbackTopic.ID())
	}

	// validate the subscriptions
	subs := make([]*pubsub.Subscription, 0, len(conf.SubscriptionID))
	for _, subID := range conf.SubscriptionID {
		sub := pubsubClient.Subscription(subID)

		ok, err := sub.Exists(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("init server: subscription %q does not exist", sub.ID())
		}

		subs = append(subs, sub)
	}

	server := &server{
		pubsubClient:      pubsubClient,
		httpClient:        httpClient,
		keyClient:         keyClient,
		transactionClient: transactionClient,
		config:            conf,
		clk:               clk,
		subs:              subs,
	}
	return server, nil
}

// close closed underlying connections.
func (s *server) close() {
	s.pubsubClient.Close()
}

// serve handles multiple Pub/Sub subscriptions in parallel.
func (s *server) serve(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, sub := range s.subs {
		// create a subscription as a local variable
		// so that it can be passed to handleSubscription safely.
		sub := sub
		g.Go(func() error {
			return s.handleSubscription(ctx, sub)
		})
	}

	log.Info("Ready to receive messages")
	return g.Wait()
}

// handleSubscription receives and handles messages from the Pub/Sub subscription.
func (s *server) handleSubscription(ctx context.Context, sub *pubsub.Subscription) error {
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		defer func() {
			// Ack the msg irrespective of whether the message was successfully processed or not
			// since we do not want the msg to be retried.
			msg.Ack()
			log.Infof("Handling of message %q ends", msg.ID)
		}()

		log.Infof("Receiving a message from %q, message ID: %q", sub.ID(), msg.ID)

		// example actions: `on_search`, `on_init`
		action, ok := msg.Attributes["action"]
		if !ok {
			log.Error(`"action" attribute is not present in the message`)
			return
		}

		var originalReq model.GenericCallbackRequest
		if err := json.Unmarshal(msg.Data, &originalReq); err != nil {
			log.Errorf("Unmarshal request failed: %v", err)
			return
		}

		// Determine the request endpoint
		// For API v1.2.0 on_search is trasmitted directly to buyer app
		url := *originalReq.Context.BapURI

		// Replace BPP data so that the callback is sended to our BPP API Service
		originalReq.Context.BppID = s.config.SubscriberID
		originalReq.Context.BppURI = s.config.SubscriberURL
		adjustedReqJSON, err := json.Marshal(originalReq)
		if err != nil {
			log.Errorf("Marshal adjusted request failed: %v", err)
			return
		}

		request, err := s.createONDCRequest(ctx, action, url, adjustedReqJSON)
		if err != nil {
			log.Errorf("Creating request failed: %v", err)
			return
		}

		response, err := s.httpClient.Do(request)
		if err != nil {
			log.Errorf("Sending request to ONDC network failed: %v", err)
			return
		}
		defer response.Body.Close()

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Errorf("Reading response body failed: %v", err)
			return
		}

		if err := s.storeTransaction(ctx, action, adjustedReqJSON, responseBody); err != nil {
			log.Errorf("Storing transaction failed: %v", err)
			return
		}

		if response.StatusCode != http.StatusOK {
			log.Infof("Sending request to ONDC network got an error: status code %d, body %s", response.StatusCode, responseBody)
			return
		}

		log.Info("Handle the message successfully")
		msg.Ack()
	})

	return err
}

// createONDCRequest create a HTTP request for ONDC network with a Authorization header.
func (s *server) createONDCRequest(ctx context.Context, action, url string, body []byte) (*http.Request, error) {
	keyset, err := s.keyClient.ServiceSigningPrivateKeyset(ctx)
	if err != nil {
		return nil, err
	}

	currentTime := s.clk.Now()
	// Use outer bound of request ttl which is 30 seconds.
	expiredTime := currentTime.Add(30 * time.Second)
	signature, err := authentication.CreateAuthSignature(body, keyset, currentTime.Unix(), expiredTime.Unix(), s.config.SubscriberID, s.config.KeyID)
	if err != nil {
		return nil, err
	}

	fullURL := url + "/" + action
	request, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", signature)
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}

func (s *server) storeTransaction(ctx context.Context, action string, requestBody, responseBody []byte) error {
	switch action {
	case "on_search":
		return storeTransaction[model.OnSearchRequest](ctx, s, action, requestBody, responseBody)
	case "on_select":
		return storeTransaction[model.OnSelectRequest](ctx, s, action, requestBody, responseBody)
	case "on_init":
		return storeTransaction[model.OnInitRequest](ctx, s, action, requestBody, responseBody)
	case "on_confirm":
		return storeTransaction[model.OnConfirmRequest](ctx, s, action, requestBody, responseBody)
	case "on_track":
		return storeTransaction[model.OnTrackRequest](ctx, s, action, requestBody, responseBody)
	case "on_cancel":
		return storeTransaction[model.OnCancelRequest](ctx, s, action, requestBody, responseBody)
	case "on_update":
		return storeTransaction[model.OnUpdateRequest](ctx, s, action, requestBody, responseBody)
	case "on_status":
		return storeTransaction[model.OnStatusRequest](ctx, s, action, requestBody, responseBody)
	case "on_rating":
		return storeTransaction[model.OnRatingRequest](ctx, s, action, requestBody, responseBody)
	case "on_support":
		return storeTransaction[model.OnSupportRequest](ctx, s, action, requestBody, responseBody)
	}
	return nil
}

func storeTransaction[R model.BAPRequest](ctx context.Context, s *server, action string, requestBody []byte, responseBody []byte) error {
	var request R
	if err := json.Unmarshal(requestBody, &request); err != nil {
		return err
	}
	msgContext := request.GetContext()

	var response model.AckResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return err
	}

	data := transactionclient.TransactionData{
		ID:              *msgContext.TransactionID,
		Type:            "CALLBACK-ACTION",
		API:             action,
		MessageID:       *msgContext.MessageID,
		Payload:         request,
		ProviderID:      msgContext.BppID,
		MessageStatus:   response.Message.Ack.Status,
		ReqReceivedTime: s.clk.Now(),
	}

	if response.Error != nil {
		data.ErrorType = response.Error.Type
		data.ErrorCode = *response.Error.Code
		data.ErrorMessage = response.Error.Message
		data.ErrorPath = response.Error.Path
	}

	return s.transactionClient.StoreTransaction(ctx, data)
}
