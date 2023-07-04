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
	"testing"
)

var validate = Validator()

/*
This demonstrate how we should create structs and `validate` struct tag for validating JSON payloads
against the given OpenAPI Specification with `playgroundvalidator` v8 .

Test schemas
schemas:
	Item:
		type: object
		properties:
			name:
				type: string
			description:
				type: string
			price:
				type: integer
			review:
				type: string
				enum:
					- good
					- bad
			available:
				type: string
				enum:
					- good
					- bad
		required:
			- name
			- price
			- available
	Location:
		type: object
		properties:
			lat:
				type: integer
			long:
				type: integer
	Request:
		type: object
		properties:
			location:
				$ref: '#/schemas/Location
			required_item:
				$ref: '#/schemas/Item'
			optional_item:
				$ref: '#/schemas/Item'
		required:
			- location
			- required_item
*/

type item struct {
	Name        *string `json:"name" validate:"required"`                   // required
	Description string  `json:"description"`                                // optional
	Price       *int    `json:"price" validate:"required"`                  // required
	Discount    int     `json:"discount"`                                   // optional
	Available   string  `json:"available" validate:"oneof=yes no"`          // required
	Review      *string `json:"review" validate:"omitempty,oneof=good bad"` // optional
}

type location struct {
	Lat  int `json:"lat"`  // optional
	Long int `json:"long"` // optional
}

type request struct {
	Location     *location `json:"location" validate:"required"`      // required
	RequiredItem *item     `json:"required_item" validate:"required"` // required
	OptionalItem *item     `json:"optional_item"`                     // optional
}

func TestValidateValidItem(t *testing.T) {
	tests := []string{
		`{
			"name": "apple",
			"description": "Good apple",
			"price": 10,
			"discount": 0,
			"available": "yes",
			"review": "good"
		}`,
		`{
			"name": "apple",
			"description": "Good apple",
			"price": 10,
			"available": "yes"
		}`,
		`{
			"name": "apple",
			"price": 10,
			"available": "yes"
		}`,
		`{
			"name": "",
			"description": "",
			"price": 0,
			"available": "yes",
			"review": "good"
		}`,
	}

	for _, test := range tests {
		var payload item
		if err := json.Unmarshal([]byte(test), &payload); err != nil {
			t.Fatalf("json.Unmarshal(%s) failed: %v", test, err)
		}
		if err := validate.Struct(payload); err != nil {
			t.Errorf("validate.Struct(%s) failed: %v", test, err)
			t.Logf("Payload struct: %#v", payload)
		}
	}
}

func TestValidateItemFail(t *testing.T) {
	tests := []string{
		`{}`,
		`{
			"name": null
		}`,
		`{
			"price": null
		}`,
		`{
			"available": null
		}`,
		`{
			"name": null,
			"price": null,
			"available": null
		}`,
		`{
			"name": "apple"
		}`,
		`{
			"name": "apple",
			"price": 10
		}`,
		`{
			"name": "apple",
			"available": "yes"
		}`,
		`{
			"price": 10,
			"available": "yes"
		}`,
		`{
			"name": "apple",
			"price": 10,
			"available": ""
		}`,
		`{
			"name": "apple",
			"price": 10,
			"available": "yes",
			"review": ""
		}`,
	}

	for _, test := range tests {
		var payload item
		if err := json.Unmarshal([]byte(test), &payload); err != nil {
			t.Fatalf("json.Unmarshal(%s) failed: %v", test, err)
		}
		if err := validate.Struct(payload); err == nil { // If NO error
			t.Errorf("validate.Struct(%s) succeed unexpectedly", test)
			t.Logf("Payload struct: %#v", payload)
		}
	}
}

func TestValidateRequestSucess(t *testing.T) {
	tests := []string{
		`{
			"location": {},
			"required_item": {
				"name": "apple",
				"description": "Good apple",
				"price": 10,
				"discount": 0,
				"available": "yes",
				"review": "good"
			},
			"optional_item": {
				"name": "apple",
				"description": "Good apple",
				"price": 10,
				"discount": 0,
				"available": "yes",
				"review": "good"
			}
		}`,
		`{
			"location": {},
			"required_item": {
				"name": "apple",
				"description": "Good apple",
				"price": 10,
				"discount": 0,
				"available": "yes",
				"review": "good"
			}
		}`,
		`{
			"location": {},
			"required_item": {
				"name": "apple",
				"description": "Good apple",
				"price": 10,
				"discount": 0,
				"available": "yes",
				"review": "good"
			},
			"optional_item": null	
		}`,
	}

	for _, test := range tests {
		var payload request
		if err := json.Unmarshal([]byte(test), &payload); err != nil {
			t.Fatalf("json.Unmarshal(%s) failed: %v", test, err)
		}
		if err := validate.Struct(payload); err != nil {
			t.Errorf("validate.Struct(%s) failed: %v", test, err)
			t.Logf("Payload struct: %#v", payload)
		}
	}
}

func TestValidateRequestFail(t *testing.T) {
	tests := []string{
		`{}`,
		`{"location": {}}`,
		`{
			"required_item": {
				"name": "apple",
				"description": "Good apple",
				"price": 10,
				"discount": 0,
				"available": "yes",
				"review": "good"
			}
		}`,
		`{
			"location": {},
			"required_item": {}
		}`,
		`{
			"location": {},
			"required_item": {
				"price": 10,
				"available": "yes"
			}
		}`,
		`{
			"location": {},
			"required_item": {
				"name": "apple",
				"description": "Good apple",
				"price": 10,
				"discount": 0,
				"available": "yes",
				"review": "good"
			},
			"optional_item": {}	
		}`,
		`{
			"location": {},
			"required_item": {
				"name": "apple",
				"description": "Good apple",
				"price": 10,
				"discount": 0,
				"available": "yes",
				"review": "good"
			},
			"optional_item": {
				"price": 10,
				"available": "yes"
				}	
		}`,
	}

	for _, test := range tests {
		var payload request
		if err := json.Unmarshal([]byte(test), &payload); err != nil {
			t.Fatalf("json.Unmarshal(%s) failed: %v", test, err)
		}
		if err := validate.Struct(payload); err == nil { // If NO error
			t.Errorf("validate.Struct(%s) succeed unexpectedly", test)
			t.Logf("Payload struct: %#v", payload)
		}
	}
}

type testData struct {
	field1 string `validate:"oneof=yes no"`
	field2 int    `validate:"oneof=10 -10"`
	field3 uint   `validate:"oneof=10 20"`
}

func TestIsOneOfSuccess(t *testing.T) {
	datas := []testData{
		{
			field1: "yes",
			field2: 10,
			field3: 20,
		},
		{
			field1: "no",
			field2: -10,
			field3: 10,
		},
	}
	for _, data := range datas {
		if err := validate.Struct(data); err != nil {
			t.Errorf("validate.Struct() failed: %v", err)
			t.Logf("Payload struct: %#v", data)
		}
	}
}
