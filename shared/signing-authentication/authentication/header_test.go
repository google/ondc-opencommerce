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
	"testing"
)

func TestExtractInfoFromHeaderSuccess(t *testing.T) {
	const exampleHeader = `Signature keyId="example-bg.com|bg3456|ed25519",algorithm="ed25519",created="1641287875",expires="1641291475",headers="(created) (expires) digest",signature="cjbhP0PFyrlSCNszJM1F/YmHDVAWsZqJUPzojnE/7TJU3fJ/rmIlgaUHEr5E0/2PIyf0tpSnWtT6cyNNlpmoAQ=="`
	got, err := ExtractInfoFromHeader(exampleHeader)
	if err != nil {
		t.Fatalf(err.Error())
	}

	want := Info{
		SubscriberID:   "example-bg.com",
		UniqueKeyID:    "bg3456",
		KeyIDAlgorithm: "ed25519",
		Algorithm:      "ed25519",
		Created:        1641287875,
		Expired:        1641291475,
		Signature:      "cjbhP0PFyrlSCNszJM1F/YmHDVAWsZqJUPzojnE/7TJU3fJ/rmIlgaUHEr5E0/2PIyf0tpSnWtT6cyNNlpmoAQ==",
	}
	if got != want {
		t.Errorf("ExtractInfoFromHeader() = %v, want = %v", got, want)
	}
}

func TestExtractInfoFromHeaderFailed(t *testing.T) {
	headers := []string{
		`Signature algorithm="ed25519",created="1641287875",expires="1641291475",headers="(created) (expires) digest",signature="cjbhP0PFyrlSCNszJM1F/YmHDVAWsZqJUPzojnE/7TJU3fJ/rmIlgaUHEr5E0/2PIyf0tpSnWtT6cyNNlpmoAQ=="`,
		`Signature keyId="example-bg.com|bg3456",algorithm="ed25519",created="1641287875",expires="1641291475",headers="(created) (expires) digest",signature="cjbhP0PFyrlSCNszJM1F/YmHDVAWsZqJUPzojnE/7TJU3fJ/rmIlgaUHEr5E0/2PIyf0tpSnWtT6cyNNlpmoAQ=="`,
		`Signature keyId="example-bg.com|bg3456|ed25519",algorithm="ed25519",created="invalid",expires="1641291475",headers="(created) (expires) digest",signature="cjbhP0PFyrlSCNszJM1F/YmHDVAWsZqJUPzojnE/7TJU3fJ/rmIlgaUHEr5E0/2PIyf0tpSnWtT6cyNNlpmoAQ=="`,
		`Signature keyId="example-bg.com|bg3456|ed25519",algorithm="ed25519",created="1641287875",expires="invalid",headers="(created) (expires) digest",signature="cjbhP0PFyrlSCNszJM1F/YmHDVAWsZqJUPzojnE/7TJU3fJ/rmIlgaUHEr5E0/2PIyf0tpSnWtT6cyNNlpmoAQ=="`,
	}
	for _, header := range headers {
		if _, err := ExtractInfoFromHeader(header); err == nil { // If NO error
			t.Errorf("ExtractInfoFromHeader() succeed unexpectedly")
		}
	}
}
