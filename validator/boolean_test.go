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
	"errors"
	"fmt"
	"testing"

	"github.com/googlegenomics/ga4gh-identity"
)

func TestBoolean(t *testing.T) {
	tests := []struct {
		name string
		in   ga4gh.Validator
		ok   bool
		err  bool
	}{
		{
			name: "true and true",
			in:   And{&Constant{OK: true}, &Constant{OK: true}},
			ok:   true,
		},
		{
			name: "true and false",
			in:   And{&Constant{OK: true}, &Constant{OK: false}},
			ok:   false,
		},
		{
			name: "true or true",
			in:   Or{&Constant{OK: true}, &Constant{OK: true}},
			ok:   true,
		},
		{
			name: "false or true",
			in:   Or{&Constant{OK: false}, &Constant{OK: true}},
			ok:   true,
		},
		{
			name: "false or false",
			in:   Or{&Constant{OK: false}, &Constant{OK: false}},
			ok:   false,
		},
		{
			name: "single-input and, true",
			in:   And{&Constant{OK: true}},
			ok:   true,
		},
		{
			name: "single-input and, false",
			in:   And{&Constant{OK: false}},
			ok:   false,
		},
		{
			name: "true and error",
			in:   And{&Constant{OK: true}, &Constant{Err: errors.New("failure")}},
			err:  true,
		},
		{
			name: "false or error",
			in:   Or{&Constant{OK: false}, &Constant{Err: errors.New("failure")}},
			err:  true,
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ok, err := test.in.Validate(ctx, nil)
			if ok != test.ok {
				t.Fatalf("Unexpected validation result, got = %v, want = %v", ok, test.ok)
			}
			if (err != nil) != test.err {
				t.Fatalf("Unexpected validation error: %v", err)
			}
		})
	}
}

func ExampleAnd() {
	id := &ga4gh.Identity{
		Role:               []ga4gh.StringValue{{Value: "human"}},
		OriginOrganization: []ga4gh.StringValue{{Value: "Earth"}},
	}
	v := And{
		&Simple{"Role": "human"},
		&Simple{"OriginOrganization": "Earth"},
	}
	if ok, err := v.Validate(context.Background(), id); ok && err == nil {
		fmt.Println("validated!")
	}
	// Output: validated!
}

func ExampleOr() {
	id := &ga4gh.Identity{
		Role: []ga4gh.StringValue{{Value: "human"}},
	}
	v := Or{
		&Simple{"Role": "human"},
		&Simple{"Role": "robot"},
	}
	if ok, err := v.Validate(context.Background(), id); ok && err == nil {
		fmt.Println("validated!")
	}
	// Output: validated!
}
