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

import (
	"encoding/json"
)

// GenericRequest is a generic form of all BPP request.
type GenericRequest struct {
	Context *Context         `json:"context" validate:"required"`
	Message *json.RawMessage `json:"message" validate:"required"`
}

// SearchRequest contains search intent for products and services.
type SearchRequest struct {
	Context *Context       `json:"context" validate:"required"`
	Message *SearchMessage `json:"message" validate:"required"`
}

// SearchMessage is an inner message of SearchRequest.
type SearchMessage struct {
	Intent *Intent `json:"intent"`
}

// SelectRequest contains selected items from the catalog for building the order.
type SelectRequest struct {
	Context *Context       `json:"context" validate:"required"`
	Message *SelectMessage `json:"message" validate:"required"`
}

// SelectMessage is an inner message of SelectRequest.
type SelectMessage struct {
	Order *Order `json:"order" validate:"required"`
}

// InitRequest contains needed information for initializing an order
// such as billing or shipping details.
type InitRequest struct {
	Context *Context     `json:"context" validate:"required"`
	Message *InitMessage `json:"message" validate:"required"`
}

// InitMessage is an inner message of InnitRequest.
type InitMessage struct {
	Order *Order `json:"order" validate:"required"`
}

// ConfirmRequest contains needed information for confirming an order.
type ConfirmRequest struct {
	Context *Context        `json:"context" validate:"required"`
	Message *ConfirmMessage `json:"message" validate:"required"`
}

// ConfirmMessage is an inner message of ConfirmRequest.
type ConfirmMessage struct {
	Order *Order `json:"order" validate:"required"`
}

// StatusRequest contains an order ID for fetching its latest status.
type StatusRequest struct {
	Context *Context       `json:"context" validate:"required"`
	Message *StatusMessage `json:"message" validate:"required"`
}

// StatusMessage is an inner message of StatusRequest.
type StatusMessage struct {
	OrderID *string `json:"order_id" validate:"required"`
}

// TrackRequest contains needed information for tracking active order.
type TrackRequest struct {
	Context *Context      `json:"context" validate:"required"`
	Message *TrackMessage `json:"message" validate:"required"`
}

// TrackMessage is an inner message of TrackRequest.
type TrackMessage struct {
	OrderID     *string `json:"order_id" validate:"required"`
	CallbackURL string  `json:"callback_url,omitempty"`
}

// CancelRequest contains needed information for cancelling an order.
type CancelRequest struct {
	Context *Context       `json:"context" validate:"required"`
	Message *CancelMessage `json:"message" validate:"required"`
}

// CancelMessage is an inner message of CancelRequest.
type CancelMessage struct {
	OrderID              *string     `json:"order_id" validate:"required"`
	CancellationReasonID string      `json:"cancellation_reason_id"`
	Descriptor           *Descriptor `json:"descriptor"`
}

// UpdateRequest contains needed information for updating an order.
type UpdateRequest struct {
	Context *Context       `json:"context" validate:"required"`
	Message *UpdateMessage `json:"message" validate:"required"`
}

// UpdateMessage is an inner message of UpdateRequest.
type UpdateMessage struct {
	// Comma separated values of order objects being updated.
	// For example: "update_target":"item,billing,fulfillment"
	UpdateTarget *string `json:"update_target" validate:"required"`

	Order *Order `json:"order" validate:"required"`
}

// RatingRequest contains rating feedback for a service.
type RatingRequest struct {
	Context *Context `json:"context" validate:"required"`
	Message *Rating  `json:"message" validate:"required"`
}

// SupportRequest contains needed information for contacting a support.
type SupportRequest struct {
	Context *Context        `json:"context" validate:"required"`
	Message *SupportMessage `json:"message" validate:"required"`
}

// SupportMessage is an inner message of SupportRequest.
type SupportMessage struct {
	RefID string `json:"ref_id,omitempty"`
}

// BPPRequest represent all request schemas of BAP API.
type BPPRequest interface {
	SearchRequest | SelectRequest | InitRequest | ConfirmRequest | StatusRequest | TrackRequest | CancelRequest | UpdateRequest | RatingRequest | SupportRequest
	GetContext() Context
}

// GetContext returns the BAP URI in the context.
func (r SearchRequest) GetContext() Context { return *r.Context }

// GetContext returns the BAP URI in the context.
func (r SelectRequest) GetContext() Context { return *r.Context }

// GetContext returns the BAP URI in the context.
func (r InitRequest) GetContext() Context { return *r.Context }

// GetContext returns the BAP URI in the context.
func (r ConfirmRequest) GetContext() Context { return *r.Context }

// GetContext returns the BAP URI in the context.
func (r StatusRequest) GetContext() Context { return *r.Context }

// GetContext returns the BAP URI in the context.
func (r TrackRequest) GetContext() Context { return *r.Context }

// GetContext returns the BAP URI in the context.
func (r CancelRequest) GetContext() Context { return *r.Context }

// GetContext returns the BAP URI in the context.
func (r UpdateRequest) GetContext() Context { return *r.Context }

// GetContext returns the BAP URI in the context.
func (r RatingRequest) GetContext() Context { return *r.Context }

// GetContext returns the BAP URI in the context.
func (r SupportRequest) GetContext() Context { return *r.Context }
