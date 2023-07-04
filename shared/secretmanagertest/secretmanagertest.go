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

// Package secretmanagertest provides helper functions for creating fake Secret Manager server.
package secretmanagertest

import (
	"context"
	"fmt"
	"net"
	"testing"

	rpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	smgrpc "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	smpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/bazelbuild/remote-apis-sdks/go/pkg/portpicker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/local"
)

// Service mocks the secret manager service.
type Service struct {
	smgrpc.UnimplementedSecretManagerServiceServer

	AccessSecretVersionResp *smpb.AccessSecretVersionResponse
	AccessSecretVersionErr  error
	SecretVersion           *rpb.SecretVersion
	SecretVersionErr        error
}

func (s *Service) AccessSecretVersion(_ context.Context, req *smpb.AccessSecretVersionRequest) (*smpb.AccessSecretVersionResponse, error) {
	return s.AccessSecretVersionResp, s.AccessSecretVersionErr
}

func (s *Service) AddSecretVersion(_ context.Context, req *smpb.AddSecretVersionRequest) (*rpb.SecretVersion, error) {
	return s.SecretVersion, s.SecretVersionErr
}

func startServerT(tb testing.TB, server *grpc.Server) string {
	tb.Helper()
	port, err := portpicker.PickUnusedPort()
	if err != nil {
		tb.Fatalf("Picking unused port: %v", err)
	}
	// Best effort to reuse the port.
	tb.Cleanup(func() { portpicker.RecycleUnusedPort(port) })

	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		tb.Fatalf("Creating TCP listener: %v", err)
	}
	// server.Stop() should close this, but do it anyway to avoid resource leak.
	tb.Cleanup(func() { lis.Close() })

	go server.Serve(lis)
	tb.Cleanup(server.Stop)
	// Listen on INADDR_ANY, but return localhost address to make
	// sure both IPv4 and IPv6 listeners are possible.
	// Using ip6-localhost breaks IPv4 tests.
	return fmt.Sprintf("localhost:%d", port)
}

// NewFakeSecretManagerGRPCCon creates a fake secret manager server.
func NewFakeSecretManagerGRPCCon(t *testing.T, fakeService *Service) *grpc.ClientConn {
	t.Helper()
	server := grpc.NewServer()
	smgrpc.RegisterSecretManagerServiceServer(server, fakeService)
	addr := startServerT(t, server)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(local.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	shutdown := func() {
		conn.Close()
		server.GracefulStop()
	}
	t.Cleanup(shutdown)
	return conn
}
