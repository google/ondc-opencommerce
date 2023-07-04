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

// Package transactionclient provide a client for storing ONDC transaction on Cloud Spanner.
package transactionclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/api/option"
)

var (
	transactionTypeMap = map[string]int{
		"REQUEST-ACTION":  1,
		"CALLBACK-ACTION": 2,
	}

	transactionAPIMap = map[string]int{
		"search":     1,
		"on_search":  2,
		"select":     3,
		"on_select":  4,
		"init":       5,
		"on_init":    6,
		"confirm":    7,
		"on_confirm": 8,
		"status":     9,
		"on_status":  10,
		"track":      11,
		"on_track":   12,
		"cancel":     13,
		"on_cancel":  14,
		"update":     15,
		"on_update":  16,
		"rating":     17,
		"on_rating":  18,
		"support":    19,
		"on_support": 20,
	}
)

// Client is a wrapper of Spanner Client for storing ONDC transaction logs.
type Client struct {
	spannerClient *spanner.Client
}

// TransactionData represent data of transaction logs.
type TransactionData struct {
	ID              string
	Type            string
	API             string
	MessageID       string
	Payload         any
	ProviderID      string
	MessageStatus   string
	ErrorType       string
	ErrorCode       string
	ErrorPath       string
	ErrorMessage    string
	ReqReceivedTime time.Time
}

// New creates a new transaction client.
func New(ctx context.Context, projectID, instanceID, databaseID string, opts ...option.ClientOption) (*Client, error) {
	database := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)
	spannerClient, err := spanner.NewClient(ctx, database, opts...)
	if err != nil {
		return nil, err
	}
	client := &Client{spannerClient: spannerClient}
	return client, nil
}

// StoreTransaction inserts the ONDC transaction details in the Spanner table.
func (c *Client) StoreTransaction(ctx context.Context, transaction TransactionData) error {
	typeCode, ok := transactionTypeMap[transaction.Type]
	if !ok {
		return fmt.Errorf("store transaction: invalid type %q", transaction.Type)
	}

	apiCode, ok := transactionAPIMap[transaction.API]
	if !ok {
		return fmt.Errorf("store transaction: invalid API %q", transaction.API)
	}

	_, err := c.spannerClient.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL: `
			INSERT INTO Transaction (
				TransactionID,
				TransactionType,
				TransactionAPI,
				MessageID,
				RequestID,
				Payload,
				ProviderID,
				MessageStatus,
				ErrorType,
				ErrorCode,
				ErrorPath,
				ErrorMessage,
				ReqReceivedTime
			)
			VALUES (
				@transactionID,
				@transactionType,
				@transactionAPI,
				@messageID,
                @requestID,
				@payload,
				@providerID,
				@messageStatus,
				@errorType,
				@errorCode,
				@errorPath,
				@errorMessage,
				@reqReceivedTime
			)`,
			Params: map[string]any{
				"transactionID":   transaction.ID,
				"transactionType": typeCode,
				"transactionAPI":  apiCode,
				"messageID":       transaction.MessageID,
				"requestID":       uuid.New().String(),
				"payload":         spanner.NullJSON{Value: transaction.Payload, Valid: true},
				"providerID":      transaction.ProviderID,
				"messageStatus":   transaction.MessageStatus,
				"errorType":       transaction.ErrorType,
				"errorCode":       transaction.ErrorCode,
				"errorPath":       transaction.ErrorPath,
				"errorMessage":    transaction.ErrorMessage,
				"reqReceivedTime": transaction.ReqReceivedTime,
			},
		}
		_, err := txn.Update(ctx, stmt)
		return err
	})
	if err != nil {
		var ae *apierror.APIError
		if errors.As(err, &ae) {
			return fmt.Errorf("storing transaction failed: %s, %s", ae.Error(), ae.Details())
		}
	}
	return err
}
