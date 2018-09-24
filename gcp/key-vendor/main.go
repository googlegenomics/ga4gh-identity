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

// The key-vendor daemon returns Google Cloud Platform service account keys for
// external GA4GH identities.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	ga4gh "github.com/googlegenomics/ga4gh-identity"
	"github.com/googlegenomics/ga4gh-identity/gcp/internal/appengine"
)

func main() {
	ctx := context.Background()
	ev := appengine.MustBuildEvaluator(ctx)
	wh := appengine.MustBuildAccountWarehouse(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/GetAccountKey", func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		id, ok := ga4gh.IdentityFromContext(ctx)
		if !ok {
			log.Printf("Context missing identity")
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}
		key, err := wh.GetAccountKey(ctx, id.Subject)
		if err != nil {
			log.Printf("Error getting account key: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(key); err != nil {
			log.Printf("Error writing response: %v", err)
			return
		}
	})
	handler := ga4gh.Handler{
		Evaluator: ev,
		Handler:   mux,
	}
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), &handler))
}
