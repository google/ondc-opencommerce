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

// Server serves HTTP request seller mockup services.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
	"time"

	log "github.com/golang/glog"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"

	_ "embed"
)

var (
	//go:embed payload-mock/on_search_request.json
	onSearchPayload string
	//go:embed payload-mock/on_select_request.json
	onSelectPayload string
	//go:embed payload-mock/on_init_request.json
	onInitPayload string
	//go:embed payload-mock/on_confirm_request.json
	onConfirmPayload string
	//go:embed payload-mock/on_status_request.json
	onStatusPayload string
	//go:embed payload-mock/on_track_request.json
	onTrackPayload string
	//go:embed payload-mock/on_cancel_request.json
	onCancelPayload string
	//go:embed payload-mock/on_update_request.json
	onUpdatePayload string
	//go:embed payload-mock/on_rating_request.json
	onRatingPayload string
	//go:embed payload-mock/on_support_request.json
	onSupportPayload string
)

var validate = model.Validator()

type server struct {
	conf config.MockSellerSystemConfig
	mux  *http.ServeMux
}

func main() {
	flag.Set("alsologtostderr", "true")

	configPath, ok := os.LookupEnv("CONFIG")
	if !ok {
		log.Exit("CONFIG env is not set")
	}

	conf, err := config.Read[config.MockSellerSystemConfig](configPath)
	if err != nil {
		log.Exit(err)
	}

	srv, err := initServer(conf)
	if err != nil {
		log.Exit(err)
	}
	log.Info("Server initialization successs")

	err = srv.serve()
	if errors.Is(err, http.ErrServerClosed) {
		log.Info("Server is closed")
	} else if err != nil {
		log.Exitf("Serving failed: %v", err)
	}
}

func initServer(conf config.MockSellerSystemConfig) (*server, error) {
	srv := &server{conf: conf}

	mux := http.NewServeMux()
	for _, e := range []struct {
		path     string
		response string
	}{
		{"/search", onSearchPayload},
		{"/select", onSelectPayload},
		{"/init", onInitPayload},
		{"/confirm", onConfirmPayload},
		{"/status", onStatusPayload},
		{"/track", onTrackPayload},
		{"/cancel", onCancelPayload},
		{"/update", onUpdatePayload},
		{"/rating", onRatingPayload},
		{"/support", onSupportPayload},
	} {
		if !json.Valid([]byte(e.response)) {
			return nil, fmt.Errorf("init server: response body of %q is not a valid JSON", e.path)
		}

		template, err := template.New(e.path).Parse(e.response)
		if err != nil {
			return nil, fmt.Errorf("init server: %s", err)
		}

		mux.Handle(e.path, mockHandler(template))
	}
	srv.mux = mux

	return srv, nil
}

func (s *server) serve() error {
	addr := fmt.Sprintf(":%d", s.conf.Port)
	log.Info("Server is serving")
	return http.ListenAndServe(addr, s.mux)
}

func mockHandler(resTemplate *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Reading request body failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var ondcCtx struct {
			Context *model.Context `json:"context" validate:"required"`
		}
		if err := json.Unmarshal(body, &ondcCtx); err != nil {
			log.Errorf("Unmarshal request body failed: %s", err)
			log.Errorf("Request body:\n%s", body)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := validate.Struct(&ondcCtx); err != nil {
			log.Errorf("Invalid request context: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		templateVal := map[string]string{
			"bap_id":         *ondcCtx.Context.BapID,
			"bap_uri":        *ondcCtx.Context.BapURI,
			"transaction_id": *ondcCtx.Context.TransactionID,
			"message_id":     *ondcCtx.Context.MessageID,
			"timestamp":      time.Now().Format(time.RFC3339),
		}
		if err := resTemplate.Execute(w, templateVal); err != nil {
			log.Errorf("Response failed: %s", err)
		}
	})
}
