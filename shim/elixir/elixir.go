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

// Package elixir provides a ga4gh.Shim implementation for translating ELIXIR
// identities into GA4GH identities.
package elixir

import (
	"context"
	"fmt"

	oidc "github.com/coreos/go-oidc"
	ga4gh "github.com/googlegenomics/ga4gh-identity"
)

const (
	issuer = "https://login.elixir-czech.org/oidc/"
)

// Shim is a ga4gh.Shim that converts ELIXIR identities into GA4GH identities.
type Shim struct {
	verifier *oidc.IDTokenVerifier
}

// NewShim creates a new Shim with the provided OIDC client ID.  If the tokens
// passed to this shim do not have an audience claim with a value equal to the
// clientID value then they will be rejected.
func NewShim(ctx context.Context, clientID string) (*Shim, error) {
	return newShim(ctx, &oidc.Config{
		ClientID: clientID,
	})
}

func newShim(ctx context.Context, config *oidc.Config) (*Shim, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("creating provider: %v", err)
	}

	return &Shim{
		verifier: provider.Verifier(config),
	}, nil
}

// Shim implements the ga4gh.Shim interface.
func (s *Shim) Shim(ctx context.Context, auth string) (*ga4gh.Identity, error) {
	token, err := s.verifier.Verify(ctx, auth)
	if err != nil {
		return nil, fmt.Errorf("verifying token: %v", err)
	}
	var claims struct {
		BonaFide *string `json:"bona_fide_status"`
	}
	if err := token.Claims(&claims); err != nil {
		return nil, fmt.Errorf("getting claims: %v", err)
	}

	id := ga4gh.Identity{
		Issuer:  token.Issuer,
		Subject: token.Subject,
	}
	if claims.BonaFide != nil {
		id.BonaFide = []ga4gh.BoolValue{
			{
				Value:  true,
				Source: issuer,
			},
		}
	}

	return &id, nil
}
