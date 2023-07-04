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

// Server handle buyer's messages and adapt Seller System to ONDC specification.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	log "github.com/golang/glog"
	"golang.org/x/sync/errgroup"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
)

type server struct {
	pubsubClient *pubsub.Client
	httpClient   *http.Client
	config       config.SellerAdapterConfig

	subs          []*pubsub.Subscription
	callbackTopic *pubsub.Topic
}

func main() {
	flag.Set("alsologtostderr", "true")
	ctx := context.Background()

	configPath, ok := os.LookupEnv("CONFIG")
	if !ok {
		log.Exit("CONFIG env is not set")
	}

	conf, err := config.Read[config.SellerAdapterConfig](configPath)
	if err != nil {
		log.Exit(err)
	}

	pubsubClient, err := pubsub.NewClient(ctx, conf.ProjectID)
	if err != nil {
		log.Exit(err)
	}

	srv, err := initializeServer(ctx, http.DefaultClient, pubsubClient, conf)
	if err != nil {
		log.Exit(err)
	}
	defer srv.close()
	log.Info("Server initialization successs")

	if err := srv.serve(ctx); err != nil {
		log.Exitf("Serving failed: %v", err)
	}
}

func initializeServer(ctx context.Context, httpClient *http.Client, pubsubClient *pubsub.Client, conf config.SellerAdapterConfig) (*server, error) {
	// validate the HTTP client.
	if httpClient == nil {
		return nil, fmt.Errorf("HTTP client is nil")
	}

	// validate the callback topic
	callbackTopic := pubsubClient.Topic(conf.CallbackTopicID)
	ok, err := callbackTopic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("topic %q does not exist", callbackTopic.ID())
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
			return nil, fmt.Errorf("subscription %q does not exist", sub.ID())
		}

		subs = append(subs, sub)
	}

	server := &server{
		pubsubClient:  pubsubClient,
		httpClient:    httpClient,
		config:        conf,
		subs:          subs,
		callbackTopic: callbackTopic,
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
		action, ok := msg.Attributes["action"]
		if !ok {
			log.Error(`"action" attribute is not present in the message`)
			return
		}

		sellerEndpoint := s.config.SellerSystemURL + "/" + action
		response, err := s.httpClient.Post(sellerEndpoint, "application/json", bytes.NewReader(msg.Data))
		if err != nil {
			log.Errorf("Sending request to %s failed: %v", sellerEndpoint, err)
			return
		}
		defer response.Body.Close()

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Errorf("Reading response body failed: %v", err)
			return
		}

		if response.StatusCode != http.StatusOK {
			log.Infof("Sending request to %s got an error: status code %d, body %s", sellerEndpoint, response.StatusCode, responseBody)
			return
		}

		publishResult := s.callbackTopic.Publish(ctx, &pubsub.Message{
			Attributes: map[string]string{
				"action": fmt.Sprintf("on_%s", action),
			},
			Data: responseBody,
		})
		if _, err := publishResult.Get(ctx); err != nil {
			log.Errorf("Publishing message failed: %v", err)
			return
		}

		log.Info("Handle the message successfully")
		msg.Ack()
	})

	return err
}
