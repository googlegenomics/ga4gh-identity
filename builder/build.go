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

// Package builder provides a way to construct a ga4gh.Evaluator from a
// protocol buffer description of it.  This is useful for applications which
// have a stored configuration.
package builder

import (
	"context"
	"fmt"

	ga4gh "github.com/googlegenomics/ga4gh-identity"
	"github.com/googlegenomics/ga4gh-identity/shim/elixir"
	"github.com/googlegenomics/ga4gh-identity/validator"
)

// Build constructs a ga4gh.Evaluator from the protocol buffer definition of an
// Evaluator.
func Build(ctx context.Context, e *Evaluator) (*ga4gh.Evaluator, error) {
	parser, err := buildParser(ctx, e.Parser)
	if err != nil {
		return nil, fmt.Errorf("building parser: %v", err)
	}

	validator, err := buildValidator(ctx, e.Validator)
	if err != nil {
		return nil, fmt.Errorf("building validator: %v", err)
	}

	return &ga4gh.Evaluator{
		Parser:    parser,
		Validator: validator,
	}, nil
}

func buildParser(ctx context.Context, p *Parser) (*ga4gh.Parser, error) {
	var shims []ga4gh.Shim
	for _, shim := range p.Shims {
		gs, err := buildShim(ctx, shim)
		if err != nil {
			return nil, fmt.Errorf("building shim: %v", err)
		}
		shims = append(shims, gs)
	}
	return ga4gh.NewParser(ctx, shims, p.Issuers)
}

func buildShim(ctx context.Context, s *Shim) (ga4gh.Shim, error) {
	switch s := s.Shim.(type) {
	case *Shim_Elixir_:
		gs, err := elixir.NewShim(ctx, s.Elixir.ClientId)
		if err != nil {
			return nil, fmt.Errorf("building Elixir shim: %v", err)
		}
		return gs, nil

	default:
		return nil, fmt.Errorf("unsupported %T shim", s)
	}
}

func buildValidator(ctx context.Context, v *Validator) (ga4gh.Validator, error) {
	if v == nil {
		return &validator.Constant{OK: false}, nil
	}

	switch v := v.Validator.(type) {
	case *Validator_And_:
		vs, err := buildValidators(ctx, v.And.Validators)
		if err != nil {
			return nil, fmt.Errorf("building 'And' validator: %v", err)
		}
		return validator.And(vs), nil

	case *Validator_Or_:
		vs, err := buildValidators(ctx, v.Or.Validators)
		if err != nil {
			return nil, fmt.Errorf("building 'Or' validator: %v", err)
		}
		return validator.Or(vs), nil

	case *Validator_Simple_:
		gv := make(validator.Simple)
		for claim, value := range v.Simple.Claims {
			gv[claim] = value
		}
		return gv, nil

	case *Validator_Constant_:
		return &validator.Constant{OK: v.Constant.Value}, nil

	default:
		return nil, fmt.Errorf("unsupported %T validator", v)
	}
}

func buildValidators(ctx context.Context, vs []*Validator) ([]ga4gh.Validator, error) {
	var built []ga4gh.Validator
	for _, v := range vs {
		gv, err := buildValidator(ctx, v)
		if err != nil {
			return nil, fmt.Errorf("building validator %q: %v", v, err)
		}
		built = append(built, gv)
	}
	return built, nil
}
