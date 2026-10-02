// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestConsentPortalFlattenExpand(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	out := bedrockagentcorecontrol.GetConsentPortalOutput{
		ConsentPortalId:  aws.String("example-1234567890"),
		ConsentPortalArn: aws.String("arn:aws:bedrock-agentcore:us-west-2:123456789012:consent-portal/example-1234567890"), //lintignore:AWSAT003,AWSAT005
		Name:             aws.String("example"),
		ExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/example"), //lintignore:AWSAT005
		PortalUrl:        aws.String("https://example.com"),
		IdpConfig: &awstypes.ConsentPortalIdpConfig{
			CredentialProviderArn: aws.String("arn:aws:bedrock-agentcore:us-west-2:123456789012:token-vault/default/oauth2credentialprovider/example"), //lintignore:AWSAT003,AWSAT005
			Scopes:                []string{"openid", names.AttrEmail},
			Audience:              aws.String("example-audience"),
		},
		Sources: []awstypes.ConsentPortalSource{{Identifier: aws.String("gateway-1234567890"), Type: awstypes.ConsentPortalSourceTypeAgentcoreGateway}},
	}
	var model consentPortalResourceModel
	if diags := fwflex.Flatten(ctx, &out, &model); diags.HasError() {
		t.Fatalf("flattening: %v", diags)
	}
	if model.ConsentPortalID.ValueString() != aws.ToString(out.ConsentPortalId) || model.PortalURL.ValueString() != aws.ToString(out.PortalUrl) {
		t.Fatalf("incorrect computed values: %#v", model)
	}
	var input bedrockagentcorecontrol.CreateConsentPortalInput
	if diags := fwflex.Expand(ctx, model, &input); diags.HasError() {
		t.Fatalf("expanding: %v", diags)
	}
	if input.IdpConfig == nil || aws.ToString(input.IdpConfig.CredentialProviderArn) != aws.ToString(out.IdpConfig.CredentialProviderArn) || len(input.IdpConfig.Scopes) != 2 || aws.ToString(input.IdpConfig.Audience) != "example-audience" {
		t.Fatalf("incorrect IdP config: %#v", input.IdpConfig)
	}
	if len(input.Sources) != 1 || aws.ToString(input.Sources[0].Identifier) != "gateway-1234567890" || input.Sources[0].Type != awstypes.ConsentPortalSourceTypeAgentcoreGateway {
		t.Fatalf("incorrect sources: %#v", input.Sources)
	}
	out.Description = aws.String("")
	out.IdpConfig.Audience = aws.String("")
	if diags := fwflex.Flatten(ctx, &out, &model); diags.HasError() {
		t.Fatalf("flattening cleared fields: %v", diags)
	}
	if !model.Description.IsNull() {
		t.Fatalf("expected null description, got %v", model.Description)
	}
	idp, diags := model.IDPConfig.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading IdP config: %v", diags)
	}
	if !idp.Audience.IsNull() {
		t.Fatalf("expected null audience, got %v", idp.Audience)
	}
	var data consentPortalDataSourceModel
	if diags := fwflex.Flatten(ctx, &out, &data); diags.HasError() {
		t.Fatalf("flattening data source: %v", diags)
	}
	if !data.IDPConfig.Equal(model.IDPConfig) || !data.Sources.Equal(model.Sources) {
		t.Fatal("resource and data source nested values differ")
	}
}

func TestConsentPortalStatus(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		code       int
		body       map[string]any
		wantStatus string
		identifier string
		wantError  bool
	}{
		{
			name:       names.AttrARN,
			identifier: "arn:aws:bedrock-agentcore:us-west-2:123456789012:consent-portal/example", //lintignore:AWSAT003,AWSAT005
			code:       http.StatusOK,
			body:       map[string]any{names.AttrStatus: "ACTIVE", "consentPortalId": "example"},
			wantStatus: "ACTIVE",
		},
		{name: "active", code: http.StatusOK, body: map[string]any{names.AttrStatus: "ACTIVE", "consentPortalId": "example"}, wantStatus: "ACTIVE"},
		{name: "missing", code: http.StatusNotFound, body: map[string]any{"__type": "ResourceNotFoundException", names.AttrMessage: "missing"}},
		{name: "access denied", code: http.StatusForbidden, body: map[string]any{"__type": "AccessDeniedException", names.AttrMessage: "denied"}, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/identities/GetConsentPortal" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var input map[string]string
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				if input["consentPortalIdentifier"] != "example" {
					t.Errorf("unexpected identifier: %v", input)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.code)
				if err := json.NewEncoder(w).Encode(tc.body); err != nil {
					t.Errorf("encoding response: %v", err)
				}
			}))
			defer server.Close()
			conn := bedrockagentcorecontrol.New(bedrockagentcorecontrol.Options{Region: "us-west-2", BaseEndpoint: aws.String(server.URL), Credentials: aws.AnonymousCredentials{}, RetryMaxAttempts: 1}) //lintignore:AWSAT003,AWSAT005
			id := tc.identifier
			if id == "" {
				id = "example"
			}
			out, status, err := statusConsentPortal(conn, id)(t.Context())
			if (err != nil) != tc.wantError {
				t.Fatalf("unexpected error: %v", err)
			}
			if status != tc.wantStatus {
				t.Errorf("status = %q, want %q", status, tc.wantStatus)
			}
			if tc.name == "missing" && out != nil {
				t.Errorf("expected missing result, got %v", out)
			}
		})
	}
}

func TestConsentPortalWaitFailure(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{names.AttrStatus: "FAILED", "statusReason": "invalid identity provider"}); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()
	conn := bedrockagentcorecontrol.New(bedrockagentcorecontrol.Options{Region: "us-west-2", BaseEndpoint: aws.String(server.URL), Credentials: aws.AnonymousCredentials{}, RetryMaxAttempts: 1}) //lintignore:AWSAT003,AWSAT005
	_, err := waitConsentPortalCreated(t.Context(), conn, "example", time.Second)
	if err == nil || !strings.Contains(err.Error(), "invalid identity provider") {
		t.Fatalf("expected AWS failure reason, got %v", err)
	}
}
