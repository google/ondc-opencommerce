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

// Package transactiontest provides helper functions for creating test transaction database.
package transactiontest

import (
	"context"
	"fmt"
	"os"
	"testing"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	databasepb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	instancepb "cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// NewDatabase create a new fake Cloud Spanner database for storing transaction data.
func NewDatabase(ctx context.Context, t *testing.T, projectID, instanceID, databaseID string) []option.ClientOption {
	t.Helper()

	addr, ok := os.LookupEnv("SPANNER_EMULATOR_ADDRESS")
	if !ok {
		addr = "localhost:9010" // default: gRPC default
	}

	opts := []option.ClientOption{
		option.WithEndpoint(addr), // gRPC Endpoint
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithInsecure()),
	}

	instanceClient, err := instance.NewInstanceAdminClient(ctx, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	createInstanceOp, err := instanceClient.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     fmt.Sprintf("projects/%s", projectID),
		InstanceId: instanceID,
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if _, err := createInstanceOp.Wait(ctx); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	dbClient, err := database.NewDatabaseAdminClient(ctx, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	createDBOp, err := dbClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID),
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", databaseID),
		ExtraStatements: []string{
			`
            CREATE TABLE Transaction(
              TransactionID STRING(36) NOT NULL,
              TransactionType INT64 NOT NULL,
              TransactionAPI INT64 NOT NULL,
              MessageID STRING(36) NOT NULL,
              RequestID STRING(36),
              Payload JSON NOT NULL,
              ProviderID STRING(255) NOT NULL,
              MessageStatus STRING(5),
              ErrorType STRING(36),
              ErrorCode STRING(255),
              ErrorPath STRING(MAX),
              ErrorMessage STRING(MAX),
              ReqReceivedTime TIMESTAMP,
              AdditionalData JSON,)
              PRIMARY KEY(TransactionID, TransactionType, MessageID, RequestID)`,
		},
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if _, err := createDBOp.Wait(ctx); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	return opts
}
