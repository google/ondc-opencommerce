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
	"regexp"

	validator "github.com/go-playground/validator/v10"
)

// Validator creates a new validator that support custom validation tags.
func Validator() *validator.Validate {
	validate := validator.New()
	validate.RegisterValidation("custom_decimal_value", isRegex(decimalValueRegex))
	validate.RegisterValidation("custom_gps", isRegex(gpsRegex))
	validate.RegisterValidation("custom_name", isRegex(nameRegex))
	return validate
}

// custom regex patterns from the ONDC API Specification
var (
	decimalValueRegex = regexp.MustCompile(`[+-]?([0-9]*[.])?[0-9]+`)
	gpsRegex          = regexp.MustCompile(`^[-+]?([1-8]?\d(\.\d+)?|90(\.0+)?),\s*[-+]?(180(\.0+)?|((1[0-7]\d)|([1-9]?\d))(\.\d+)?)$`)
	nameRegex         = regexp.MustCompile(`^\./[^/]+/[^/]*/[^/]*/[^/]*/[^/]*/[^/]*$`)
)

func isRegex(regex *regexp.Regexp) validator.Func {
	return func(fl validator.FieldLevel) bool {
		return regex.MatchString(fl.Field().String())
	}
}
