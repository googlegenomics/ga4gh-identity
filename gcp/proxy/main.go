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
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	ga4gh "github.com/googlegenomics/ga4gh-identity"
	"github.com/googlegenomics/ga4gh-identity/builder"
	"github.com/googlegenomics/ga4gh-identity/gcp"
	"golang.org/x/oauth2/google"
)

func main() {
	target := mustGetenv("TARGET")
	t, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Error parsing TARGET=%q: %v", target, err)
	}

	ev, err := buildEvaluator(mustGetenv("CONFIG"))
	if err != nil {
		log.Fatalf("Error building evaluator: %v", err)
	}

	client, err := google.DefaultClient(context.Background(), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		log.Fatalf("Error creating HTTP client: %v", err)
	}

	wh, err := gcp.NewAccountWarehouse(client, &gcp.AccountWarehouseOptions{
		Project:     mustGetenv("PROJECT"),
		DefaultRole: mustGetenv("ROLE"),
		Scopes:      strings.Split(mustGetenv("SCOPES"), ","),
	})
	if err != nil {
		log.Fatalf("Error creating account warehouse: %v", err)
	}

	log.Fatal(http.ListenAndServe(":"+mustGetenv("PORT"), newProxy(t, ev, wh)))
}

func mustGetenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Environment variable %q must be set, see app.yaml for more information.", key)
	}
	return v
}

func buildEvaluator(encoded string) (*ga4gh.Evaluator, error) {
	var e builder.Evaluator
	if err := proto.UnmarshalText(encoded, &e); err != nil {
		return nil, fmt.Errorf("unmarshaling evaluator: %v", err)
	}
	return builder.Build(context.Background(), &e)
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
