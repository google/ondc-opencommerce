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

// Package errorcode is responsible for looking up appropriate ONDC error code.
//
// All error codes are define in [ONDC Error Codes]
//
// [ONDC Error Codes]: https://github.com/ONDC-Official/ONDC-Protocol-Specs/blob/master/protocol-specifications/docs/draft/Error%20Codes.md
package errorcode

// ErrType is an ONDC error.
type ErrType string

// Role is a role in ONDC network.
type Role string

type lookupKey struct {
	role Role
	err  ErrType
}

// Roles in ONDC network
const (
	RoleGateway   Role = "Gateway"
	RoleBuyerApp  Role = "Buyer App"
	RoleSellerApp Role = "Seller App"
	RoleLogistics Role = "Logistics"
)

// Error defined in ONDC specification
const (
	ErrInvalidSignature ErrType = "Invalid Signature"
	ErrInvalidRequest   ErrType = "Invalid Request"
)

// This table does not contain all of ONDC error code
// but it is enough for our use cases.
var lookupTable = map[lookupKey]int{
	{role: RoleGateway, err: ErrInvalidRequest}:   10000,
	{role: RoleGateway, err: ErrInvalidSignature}: 10001,

	{role: RoleBuyerApp, err: ErrInvalidSignature}: 20001,

	{role: RoleSellerApp, err: ErrInvalidRequest}:   30000,
	{role: RoleSellerApp, err: ErrInvalidSignature}: 30016,

	{role: RoleLogistics, err: ErrInvalidSignature}: 60005,
	{role: RoleLogistics, err: ErrInvalidRequest}:   60006,
}

// Lookup lookups for corresponding error code with given role and error.
func Lookup(role Role, err ErrType) (int, bool) {
	code, ok := lookupTable[lookupKey{role: role, err: err}]
	return code, ok
}
