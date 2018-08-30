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
	"net/http"
	"strings"
)

// Handler implements an http.Handler that parses an incoming identity,
// validates it, and then passes it to an underlying http.Handler.  The
// http.Request passed to the underlying handler has an identity associated
// with it via NewIdentityContext.
type Handler struct {
	// Evaluator is used to provide the parsing and validation logic.
	Evaluator *Evaluator

	// Handler is invoked only if the incoming identity could be parsed and
	// validated.  The http.Request will have a ga4gh.Identity associated with it
	// via NewIdentityContext.
	Handler http.Handler
}

// ServeHTTP implements the http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	parts := strings.SplitN(req.Header.Get("authorization"), " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		http.Error(w, "authorization requires a bearer token", http.StatusUnauthorized)
		return
	}

	ctx := req.Context()
	id, err := h.Evaluator.Evaluate(ctx, parts[1])
	if err != nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	h.Handler.ServeHTTP(w, req.WithContext(NewIdentityContext(ctx, id)))
}
