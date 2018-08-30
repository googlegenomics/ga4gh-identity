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
)

// Evaluator combines both parsing and validation of authorization tokens.
type Evaluator struct {
	Parser    *Parser
	Validator Validator
}

// Evaluate attempts to parse auth using ev.Parser, and then validate it using
// ev.Validator.  It will only return a non-error result if the identity both
// parses and validates.
func (ev *Evaluator) Evaluate(ctx context.Context, auth string) (*Identity, error) {
	if auth == "" {
		return nil, errors.New("empty authorization")
	}

	id, err := ev.Parser.Parse(ctx, auth)
	if err != nil {
		return nil, fmt.Errorf("parsing authorization: %v", err)
	}

	ok, err := ev.Validator.Validate(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("validating identity: %v", err)
	}
	if !ok {
		return nil, errors.New("validation failed")
	}

	return id, nil
}
