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

package errorcode

import (
	"testing"
)

func TestLookupSucess(t *testing.T) {
	tests := []struct {
		role Role
		err  ErrType
		want int
	}{
		{
			role: RoleGateway,
			err:  ErrInvalidSignature,
			want: 10001,
		},
		{
			role: RoleSellerApp,
			err:  ErrInvalidRequest,
			want: 30000,
		},
		{
			role: RoleLogistics,
			err:  ErrInvalidRequest,
			want: 60006,
		},
	}

	for _, test := range tests {
		got, ok := Lookup(test.role, test.err)

		if !ok {
			t.Fatalf("Lookup(%q, %q) do not found the result", test.role, test.err)
		}
		if got != test.want {
			t.Errorf("Lookup(%q, %q) = %d, want %d", test.role, test.err, got, test.want)
		}
	}
}

func TestLookupNotFound(t *testing.T) {
	tests := []struct {
		role Role
		err  ErrType
	}{
		{
			role: RoleBuyerApp,
			err:  ErrInvalidRequest,
		},
		{
			role: "Unknown Role",
			err:  ErrInvalidRequest,
		},
		{
			role: RoleLogistics,
			err:  "Unknown Error",
		},
	}

	for _, test := range tests {
		if _, ok := Lookup(test.role, test.err); ok {
			t.Errorf("Lookup(%q, %q) unexpectedly found the result", test.role, test.err)
		}
	}
}
