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

// Package ga4gh provides primitives for dealing with identities as described
// by the Global Alliance for Genomics and Healthcare's Data Use and Researcher
// Identity workstream.
package ga4gh

// StringValue represents a string value and claim source.
type StringValue struct {
	Value  string `json:"value"`
	Source string `json:"source"`
}

// BoolValue represents a boolean value and claim source.
type BoolValue struct {
	Value  bool   `json:"value"`
	Source string `json:"source"`
}

// Identity is a GA4GH identity as described by the Data Use and Researcher
// Identity stream.
type Identity struct {
	Subject string `json:"sub,omitempty"`
	Issuer  string `json:"iss,omitempty"`

	OriginOrganization              []StringValue `json:"ga4gh.IdentityOriginOrganization"`
	AcademicInstitutionAffiliations []StringValue `json:"ga4gh.AcademicInstitutionAffiliations"`
	Role                            []StringValue `json:"ga4gh.Role"`
	HasAcknowledgedEthicsTerms      []StringValue `json:"ga4gh.HasAcknowledgedEthicsTerms"`
	BonaFide                        []BoolValue   `json:"ga4gh.ResearcherStatus.BonaFide"`
}
