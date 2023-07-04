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

package model

import "encoding/json"

// GenericCallbackRequest is a generic form of all BPP callback request.
type GenericCallbackRequest struct {
	Context *Context         `json:"context" validate:"required"`
	Message *json.RawMessage `json:"message" validate:"required"`
	Error   *json.RawMessage `json:"error,omitempty"`
}

// OnSearchRequest contains on_search Catalog for products and services.
type OnSearchRequest struct {
	Context *Context         `json:"context" validate:"required"`
	Message *OnSearchMessage `json:"message,omitempty"`
	Error   *Error           `json:"error,omitempty"`
}

// OnSearchMessage is an inner message of OnSearchRequest.
type OnSearchMessage struct {
	Catalog *Catalog `json:"catalog" validate:"required"`
}

// OnSelectRequest contains an on_select Catalog for products and services.
type OnSelectRequest struct {
	Context *Context         `json:"context" validate:"required"`
	Message *OnSelectMessage `json:"message,omitempty"`
	Error   *Error           `json:"error,omitempty"`
}

// OnSelectMessage is an inner message of OnSelectRequest.
type OnSelectMessage struct {
	Order *Order `json:"order" validate:"required"`
}

// OnInitRequest contains an on_init Catalog for products and services.
type OnInitRequest struct {
	Context *Context       `json:"context" validate:"required"`
	Message *OnInitMessage `json:"message,omitempty"`
	Error   *Error         `json:"error,omitempty"`
}

// OnInitMessage is an inner message of OnInitRequest.
type OnInitMessage struct {
	Order *Order `json:"order" validate:"required"`
}

// OnConfirmRequest contains an on_confirm Catalog for products and services.
type OnConfirmRequest struct {
	Context *Context          `json:"context" validate:"required"`
	Message *OnConfirmMessage `json:"message,omitempty"`
	Error   *Error            `json:"error,omitempty"`
}

// OnConfirmMessage is an inner message of OnConfirmRequest.
type OnConfirmMessage struct {
	Order *Order `json:"order" validate:"required"`
}

// OnTrackRequest contains an on_track Catalog for products and services.
type OnTrackRequest struct {
	Context *Context        `json:"context" validate:"required"`
	Message *OnTrackMessage `json:"message,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// OnTrackMessage is an inner message of OnTrackRequest.
type OnTrackMessage struct {
	Tracking *Tracking `json:"tracking" validate:"required"`
}

// OnCancelRequest contains an on_cancel Catalog for products and services.
type OnCancelRequest struct {
	Context *Context         `json:"context" validate:"required"`
	Message *OnCancelMessage `json:"message,omitempty"`
	Error   *Error           `json:"error,omitempty"`
}

// OnCancelMessage is an inner message of OnCancelRequest.
type OnCancelMessage struct {
	Order *Order `json:"order" validate:"required"`
}

// OnUpdateRequest contains an on_update Catalog for products and services.
type OnUpdateRequest struct {
	Context *Context         `json:"context" validate:"required"`
	Message *OnUpdateMessage `json:"message,omitempty"`
	Error   *Error           `json:"error,omitempty"`
}

// OnUpdateMessage is an inner message of OnUpdateRequest.
type OnUpdateMessage struct {
	Order *Order `json:"order" validate:"required"`
}

// OnStatusRequest contains an on_status Catalog for products and services.
type OnStatusRequest struct {
	Context *Context         `json:"context" validate:"required"`
	Message *OnStatusMessage `json:"message,omitempty"`
	Error   *Error           `json:"error,omitempty"`
}

// OnStatusMessage is an inner message of OnStatusRequest.
type OnStatusMessage struct {
	Order *Order `json:"order" validate:"required"`
}

// OnRatingRequest contains an on_rating Catalog for products and services.
type OnRatingRequest struct {
	Context *Context   `json:"context" validate:"required"`
	Message *ratingAck `json:"message,omitempty"`
	Error   *Error     `json:"error,omitempty"`
}

// OnSupportRequest contains an on_support Catalog for products and services.
type OnSupportRequest struct {
	Context *Context          `json:"context" validate:"required"`
	Message *OnSupportMessage `json:"message,omitempty"`
	Error   *Error            `json:"error,omitempty"`
}

// OnSupportMessage is an inner message of OnSupportRequest.
type OnSupportMessage struct {
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// BAPRequest represent all request schemas of BAP API.
type BAPRequest interface {
	OnSearchRequest | OnSelectRequest | OnInitRequest | OnConfirmRequest | OnStatusRequest | OnTrackRequest | OnCancelRequest | OnUpdateRequest | OnRatingRequest | OnSupportRequest
	GetContext() Context
}

// GetContext returns the ONDC context of the request.
func (r OnSearchRequest) GetContext() Context { return *r.Context }

// GetContext returns the ONDC context of the request.
func (r OnSelectRequest) GetContext() Context { return *r.Context }

// GetContext returns the ONDC context of the request.
func (r OnInitRequest) GetContext() Context { return *r.Context }

// GetContext returns the ONDC context of the request.
func (r OnConfirmRequest) GetContext() Context { return *r.Context }

// GetContext returns the ONDC context of the request.
func (r OnStatusRequest) GetContext() Context { return *r.Context }

// GetContext returns the ONDC context of the request.
func (r OnTrackRequest) GetContext() Context { return *r.Context }

// GetContext returns the ONDC context of the request.
func (r OnCancelRequest) GetContext() Context { return *r.Context }

// GetContext returns the ONDC context of the request.
func (r OnUpdateRequest) GetContext() Context { return *r.Context }

// GetContext returns the ONDC context of the request.
func (r OnRatingRequest) GetContext() Context { return *r.Context }

// GetContext returns the ONDC context of the request.
func (r OnSupportRequest) GetContext() Context { return *r.Context }
