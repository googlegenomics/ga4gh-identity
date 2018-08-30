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

// Package shim provides implementations of the ga4gh.Shim interface for
// shimming between different identity providers and GA4GH identities.
package shim

import (
	"context"

	ga4gh "github.com/googlegenomics/ga4gh-identity"
)

// Static is a ga4gh.Shim that returns a single static Identity.
type Static struct {
	Identity *ga4gh.Identity
}

// Shim implements the ga4gh.Shim interface.
func (s *Static) Shim(context.Context, string) (*ga4gh.Identity, error) {
	return s.Identity, nil
}
