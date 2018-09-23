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

// This package provides a single-host reverse proxy that rewrites bearer
// tokens in Authorization headers to be Google Cloud Platform access tokens.
// For configuration information see app.yaml.
package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	ga4gh "github.com/googlegenomics/ga4gh-identity"
	"github.com/googlegenomics/ga4gh-identity/gcp"
	"github.com/googlegenomics/ga4gh-identity/gcp/internal/appengine"
)

func main() {
	target := os.Getenv("TARGET")
	t, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Error parsing TARGET=%q: %v", target, err)
	}

	ctx := context.Background()
	ev := appengine.MustBuildEvaluator(ctx)
	wh := appengine.MustBuildAccountWarehouse(ctx)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), newProxy(t, ev, wh)))
}

type proxy struct {
	*httputil.ReverseProxy
	target    *url.URL
	evaluator *ga4gh.Evaluator
	warehouse *gcp.AccountWarehouse
}

func newProxy(target *url.URL, evaluator *ga4gh.Evaluator, warehouse *gcp.AccountWarehouse) *proxy {
	p := &proxy{
		target:    target,
		evaluator: evaluator,
		warehouse: warehouse,
	}
	p.ReverseProxy = &httputil.ReverseProxy{
		Director: p.director,
	}
	return p
}

func (p *proxy) director(req *http.Request) {
	auth := strings.Fields(req.Header.Get("Authorization"))
	if len(auth) == 2 && strings.ToLower(auth[0]) == "bearer" {
		p.swapAuthHeader(req, auth[1])
	}

	req.Host = ""
	req.URL.Scheme = p.target.Scheme
	req.URL.Host = p.target.Host
}

func (p *proxy) swapAuthHeader(req *http.Request, auth string) {
	ctx := req.Context()
	id, err := p.evaluator.Evaluate(ctx, auth)
	if err != nil {
		log.Printf("Error during evaluation: %v", err)
		return
	}

	token, err := p.warehouse.GetAccessToken(ctx, id.Subject)
	if err != nil {
		log.Printf("Error getting access token: %v", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
}
