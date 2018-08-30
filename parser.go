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

package ga4gh

import (
	"context"
	"errors"
	"fmt"

	oidc "github.com/coreos/go-oidc"
	"gopkg.in/square/go-jose.v2/jwt"
)

// Shim is used to convert an HTTP bearer authorization string that is _not_ in
// the normal Identity format into an Identity.  This is useful when
// interoperating with systems that do not yet provide a GA4GH identity.
type Shim interface {
	Shim(ctx context.Context, auth string) (*Identity, error)
}

// Parser parses OIDC bearer tokens into Identity structs.
type Parser struct {
	shims   []Shim
	issuers map[string]*oidc.IDTokenVerifier
}

// NewParser constructs a new Parser using shims for translating external
// identities and issuers as a map of OAuth 2.0 base URLs to client IDs.  When
// parsing an identity token it first tries to use each of the shims in order
// to perform the conversion.  If none of the shims succeed it then checks if
// the token was issued by any of the OAuth 2.0 providers in issuers, and
// directly accepting the claims present if it is.
func NewParser(ctx context.Context, shims []Shim, issuers map[string]string) (*Parser, error) {
	iss := make(map[string]*oidc.IDTokenVerifier)
	for issuer, clientID := range issuers {
		provider, err := oidc.NewProvider(ctx, issuer)
		if err != nil {
			return nil, fmt.Errorf("creating provider for %q: %v", issuer, err)
		}
		iss[issuer] = provider.Verifier(&oidc.Config{ClientID: clientID})
	}
	return &Parser{
		shims:   shims,
		issuers: iss,
	}, nil
}

// Parse takes an authorization string (usually an HTTP authorization bearer
// token) and converts it into an Identity.
func (p *Parser) Parse(ctx context.Context, auth string) (*Identity, error) {
	for _, shim := range p.shims {
		id, err := shim.Shim(ctx, auth)
		if err == nil {
			return id, nil
		}
	}

	parsed, err := jwt.ParseSigned(auth)
	if err != nil {
		return nil, fmt.Errorf("parsing JWT: %v", err)
	}

	// Unwrap the claims in the incoming JWT so that we can determine which
	// IDTokenVerifier is responsible for this token.  Ignoring verification
	// failures is safe here because the per-issuer verifiers will fully verify
	// the token.
	var claims jwt.Claims
	if err := parsed.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, fmt.Errorf("extracting base claims: %v", err)
	}

	verifier, ok := p.issuers[claims.Issuer]
	if !ok {
		return nil, errors.New("invalid issuer")
	}

	token, err := verifier.Verify(ctx, auth)
	if err != nil {
		return nil, fmt.Errorf("verifying token: %v", err)
	}

	var id Identity
	if err := token.Claims(&id); err != nil {
		return nil, fmt.Errorf("extracting claims: %v", err)
	}

	return &id, nil
}
