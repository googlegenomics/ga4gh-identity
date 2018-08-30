// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"context"
	"fmt"
	"reflect"

	ga4gh "github.com/googlegenomics/ga4gh-identity"
)

// valueType is the set of types which are treated like claims that have a
// value and source as sociated with them.
var valueType = map[reflect.Type]bool{
	reflect.TypeOf([]ga4gh.StringValue{}): true,
	reflect.TypeOf([]ga4gh.BoolValue{}):   true,
}

// Simple is a ga4gh.Validator that compares the values in the incoming
// identity to match the keys and values it contains.  For example, the Simple
// validator: &Simple{"Role": "human"} would validate all identities containing
// at least one Role with the value "human".  The Simple Validator does not
// inspect the source of claims.
type Simple map[string]interface{}

// Validate returns true iff there is a corresponding field in the input that
// contains at least one value that matches each of the key and values stored
// in the Simple.
func (s Simple) Validate(ctx context.Context, identity *ga4gh.Identity) (bool, error) {
	v := reflect.ValueOf(*identity)
	for name, expected := range s {
		field := v.FieldByName(name)
		if !field.IsValid() {
			return false, fmt.Errorf("no field named %q on ga4gh.Identity", name)
		}
		if valueType[field.Type()] {
			var matched bool
			for i := 0; i < field.Len(); i++ {
				if reflect.DeepEqual(field.Index(i).FieldByName("Value").Interface(), expected) {
					matched = true
				}
			}
			if !matched {
				return false, nil
			}
		} else {
			if !reflect.DeepEqual(field.Interface(), expected) {
				return false, nil
			}
		}
	}
	return true, nil
}
