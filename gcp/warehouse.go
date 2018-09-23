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

// Package gcp abstracts interacting with certain aspects of Google Cloud
// Platform, such as creating service account keys and access tokens.
package gcp

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"path"
	"strings"

	"golang.org/x/crypto/sha3"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	iam "google.golang.org/api/iam/v1"
	iamcredentials "google.golang.org/api/iamcredentials/v1"
)

// AccountWarehouseOptions is used with NewWarehouse to configure a new warehouse.
type AccountWarehouseOptions struct {
	Project     string
	DefaultRole string
	Scopes      []string
}

// AccountWarehouse is used to create Google Cloud Platform Service Account
// keys and access tokens associated with a specific identity.
type AccountWarehouse struct {
	opts  AccountWarehouseOptions
	iam   *iam.Service
	creds *iamcredentials.Service
	crm   *cloudresourcemanager.Service
}

// NewAccountWarehouse creates a new AccountWarehouse using the provided client
// and options.
func NewAccountWarehouse(client *http.Client, opts *AccountWarehouseOptions) (*AccountWarehouse, error) {
	iamSvc, err := iam.New(client)
	if err != nil {
		return nil, fmt.Errorf("creating IAM client: %v", err)
	}

	creds, err := iamcredentials.New(client)
	if err != nil {
		return nil, fmt.Errorf("creating IAM credentials client: %v", err)
	}

	crm, err := cloudresourcemanager.New(client)
	if err != nil {
		return nil, fmt.Errorf("creating cloud resource manager client: %v", err)
	}

	return &AccountWarehouse{
		opts:  *opts,
		iam:   iamSvc,
		creds: creds,
		crm:   crm,
	}, nil
}

// GetAccountKey returns a service account key associated with id.
func (wh *AccountWarehouse) GetAccountKey(ctx context.Context, id string) ([]byte, error) {
	account, err := wh.getBackingAccount(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting backing account: %v", err)
	}

	// TODO(#2): Remove old keys from service account before creating new keys.
	keys := wh.iam.Projects.ServiceAccounts.Keys
	result, err := keys.Create(accountID("-", account), &iam.CreateServiceAccountKeyRequest{
		PrivateKeyType: "TYPE_GOOGLE_CREDENTIALS_FILE",
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("creating key: %v", err)
	}

	out, err := base64.StdEncoding.DecodeString(result.PrivateKeyData)
	if err != nil {
		return nil, fmt.Errorf("decoding key: %v", err)
	}

	return out, nil
}

// GetAccessToken returns an access token for the service account uniquely
// associated with id.
func (wh *AccountWarehouse) GetAccessToken(ctx context.Context, id string) (string, error) {
	account, err := wh.getBackingAccount(ctx, id)
	if err != nil {
		return "", fmt.Errorf("getting backing account: %v", err)
	}

	response, err := wh.creds.Projects.ServiceAccounts.GenerateAccessToken(accountID("-", account), &iamcredentials.GenerateAccessTokenRequest{
		Scope: wh.opts.Scopes,
	}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("generating access token: %v", err)
	}

	return response.AccessToken, nil
}

func (wh *AccountWarehouse) getBackingAccount(ctx context.Context, id string) (string, error) {
	service := wh.iam.Projects.ServiceAccounts

	hid := hashID(id)
	name := accountID(wh.opts.Project, fmt.Sprintf("%s@%s.iam.gserviceaccount.com", hid, wh.opts.Project))
	account, err := service.Get(name).Context(ctx).Do()
	if err == nil {
		if err := wh.configureRole(ctx, account.Email); err != nil {
			return "", fmt.Errorf("configuring role for existing account: %v", err)
		}
		return account.Email, nil
	}
	if err, ok := err.(*googleapi.Error); !ok || err.Code != http.StatusNotFound {
		return "", fmt.Errorf("getting account: %v", err)
	}

	account, err = service.Create(projectID(wh.opts.Project), &iam.CreateServiceAccountRequest{
		AccountId: hid,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: id,
		},
	}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("creating backing account: %v", err)
	}
	if err := wh.configureRole(ctx, account.Email); err != nil {
		return "", fmt.Errorf("configuring role for new account: %v", err)
	}
	return account.Email, nil
}

func (wh *AccountWarehouse) configureRole(ctx context.Context, email string) error {
	role := wh.opts.DefaultRole

	project := wh.opts.Project
	if parts := strings.Split(role, "/"); len(parts) == 4 && parts[0] == "projects" {
		project = parts[1]
	}

	projects := wh.crm.Projects
	policy, err := projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy for project %q: %v", project, err)
	}

	var binding *cloudresourcemanager.Binding
	for _, b := range policy.Bindings {
		if b.Role == role {
			binding = b
			break
		}
	}
	if binding == nil {
		return fmt.Errorf("no bindings for %q in policy", role)
	}

	qualifiedName := "serviceAccount:" + email
	for _, member := range binding.Members {
		if member == qualifiedName {
			return nil
		}
	}

	binding.Members = append(binding.Members, qualifiedName)
	_, err = projects.SetIamPolicy(project, &cloudresourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy for project %q: %v", project, err)
	}
	return nil
}

func hashID(id string) string {
	hash := sha3.Sum224([]byte(id))
	return "i" + hex.EncodeToString(hash[:])[:29]
}

func keyID(project, account, key string) string {
	return path.Join(accountID(project, account), "keys", key)
}

func accountID(project, account string) string {
	return path.Join(projectID(project), "serviceAccounts", account)
}

func projectID(project string) string {
	return path.Join("projects", project)
}
