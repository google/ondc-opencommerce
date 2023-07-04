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

package authentication

import (
	"fmt"
	"strconv"
	"strings"
)

// Info represents authentication information in ONDC.
type Info struct {
	SubscriberID   string
	UniqueKeyID    string
	KeyIDAlgorithm string
	Algorithm      string
	Created        int64
	Expired        int64
	Signature      string
}

// ExtractInfoFromHeader extracts authentication information from the given header value.
func ExtractInfoFromHeader(header string) (Info, error) {
	var info Info
	values := make(map[string]string)

	trimmedHeader := strings.TrimPrefix(header, "Signature ")
	for _, element := range strings.Split(trimmedHeader, ",") {
		key, val, found := strings.Cut(element, "=")
		if found {
			key = strings.TrimSpace(key)
			val = strings.TrimSpace(val)
			val = strings.Trim(val, "\"")
			values[key] = val
		}
	}

	// check if needed keys are present.
	for _, k := range [6]string{"keyId", "algorithm", "created", "expires", "headers", "signature"} {
		if _, ok := values[k]; !ok {
			return info, fmt.Errorf("Missing %s", k)
		}
	}

	// extract info from key ID
	keyInfo := strings.Split(values["keyId"], "|")
	if len(keyInfo) != 3 {
		return info, fmt.Errorf("invalid Key ID: %s", values["keyId"])
	}
	subscriberID := keyInfo[0]
	uniqueKeyID := keyInfo[1]
	keyIDAlgorithm := keyInfo[2]

	created, err := strconv.Atoi(values["created"])
	if err != nil {
		return info, err
	}

	expires, err := strconv.Atoi(values["expires"])
	if err != nil {
		return info, err
	}

	info = Info{
		SubscriberID:   subscriberID,
		UniqueKeyID:    uniqueKeyID,
		KeyIDAlgorithm: keyIDAlgorithm,
		Algorithm:      values["algorithm"],
		Created:        int64(created),
		Expired:        int64(expires),
		Signature:      values["signature"],
	}
	return info, nil
}
