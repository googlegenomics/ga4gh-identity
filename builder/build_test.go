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

package builder

import (
	"context"
	"testing"

	ga4gh "github.com/googlegenomics/ga4gh-identity"
)

var id = &ga4gh.Identity{
	OriginOrganization: []ga4gh.StringValue{
		{Value: "Earth"},
		{Value: "Mars"},
	},
	Role: []ga4gh.StringValue{
		{Value: "human"},
		{Value: "person"},
	},
}

func TestBuildValidator(t *testing.T) {
	tests := []struct {
		name string
		in   *Validator
		ok   bool
	}{
		{
			name: "simple",
			in: &Validator{
				Validator: &Validator_Simple_{
					Simple: &Validator_Simple{
						Claims: map[string]string{
							"OriginOrganization": "Mars",
							"Role":               "person",
						},
					},
				},
			},
			ok: true,
		},
		{
			name: "constant true",
			in: &Validator{
				Validator: &Validator_Constant_{
					Constant: &Validator_Constant{
						Value: true,
					},
				},
			},
			ok: true,
		},
		{
			name: "constant false",
			in: &Validator{
				Validator: &Validator_Constant_{
					Constant: &Validator_Constant{
						Value: false,
					},
				},
			},
			ok: false,
		},
		{
			name: "boolean and",
			in: &Validator{
				Validator: &Validator_And_{
					And: &Validator_And{
						Validators: []*Validator{
							{
								Validator: &Validator_Simple_{
									Simple: &Validator_Simple{
										Claims: map[string]string{"Role": "human"},
									},
								},
							},
							{
								Validator: &Validator_Simple_{
									Simple: &Validator_Simple{
										Claims: map[string]string{"OriginOrganization": "Earth"},
									},
								},
							},
						},
					},
				},
			},
			ok: true,
		},
		{
			name: "boolean or",
			in: &Validator{
				Validator: &Validator_Or_{
					Or: &Validator_Or{
						Validators: []*Validator{
							{
								Validator: &Validator_Simple_{
									Simple: &Validator_Simple{
										Claims: map[string]string{"Role": "robot"},
									},
								},
							},
							{
								Validator: &Validator_Simple_{
									Simple: &Validator_Simple{
										Claims: map[string]string{"Role": "toaster"},
									},
								},
							},
						},
					},
				},
			},
			ok: false,
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator, err := buildValidator(ctx, test.in)
			if err != nil {
				t.Fatalf("Error building validator: %v", err)
			}
			ok, err := validator.Validate(ctx, id)
			if err != nil {
				t.Fatalf("Error validating identity: %v", err)
			}
			if ok != test.ok {
				t.Fatalf("Validate() = %v, want = %v", ok, test.ok)
			}
		})
	}
}
