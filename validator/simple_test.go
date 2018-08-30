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
	"testing"

	"github.com/googlegenomics/ga4gh-identity"
)

func TestSimple(t *testing.T) {
	tests := []struct {
		name      string
		id        *ga4gh.Identity
		validator Simple
		ok        bool
		err       bool
	}{
		{
			name:      "empty identity",
			id:        &ga4gh.Identity{},
			validator: Simple{},
			ok:        true,
		},
		{
			name: "simple success",
			id: &ga4gh.Identity{
				Role: []ga4gh.StringValue{{Value: "person"}},
			},
			validator: Simple{"Role": "person"},
			ok:        true,
		},
		{
			name:      "simple failure",
			id:        &ga4gh.Identity{},
			validator: Simple{"Role": "researcher"},
			ok:        false,
		},
		{
			name:      "scalar field",
			id:        &ga4gh.Identity{Issuer: "https://very-real.idp"},
			validator: Simple{"Issuer": "https://very-real.idp"},
			ok:        true,
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ok, err := test.validator.Validate(ctx, test.id)
			if test.err != (err != nil) {
				t.Fatalf("Unexpected error during validation: %v", err)
			}
			if test.ok != ok {
				t.Fatalf("Unexpected validation result, got = %v, wanted = %v", ok, test.ok)
			}
		})
	}
}

func ExampleSimple() {
	id := &ga4gh.Identity{
		Issuer: "https://very-real.idp",
		Role:   []ga4gh.StringValue{{Value: "human"}},
	}
	v := Simple{
		"Issuer": "https://very-real.idp",
		"Role":   "human",
	}
	ok, err := v.Validate(context.Background(), id)
	if err == nil && ok {
		fmt.Println("validated!")
	}
	// Output: validated!
}
