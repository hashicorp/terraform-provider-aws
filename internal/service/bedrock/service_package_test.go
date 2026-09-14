// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestServicePackageNewClient_ignoresAWSBearerToken(t *testing.T) {
	t.Setenv("AWS_BEARER_TOKEN_BEDROCK", "test-token")

	authorization := make(chan string, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization <- r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"modelSummaries":[]}`)); err != nil {
			t.Errorf("writing response: %s", err)
		}
	}))
	t.Cleanup(server.Close)

	cfg := aws.Config{
		Region:      "us-west-2",
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("access-key", "secret-key", "")),
		HTTPClient:  server.Client(),
	}
	client, err := (&servicePackage{}).NewClient(t.Context(), map[string]any{
		"aws_sdkv2_config": &cfg,
		names.AttrEndpoint: server.URL,
		names.AttrRegion:   cfg.Region,
	})
	if err != nil {
		t.Fatalf("creating Bedrock client: %s", err)
	}

	if _, err := client.ListFoundationModels(t.Context(), &bedrock.ListFoundationModelsInput{}); err != nil {
		t.Fatalf("listing Bedrock foundation models: %s", err)
	}

	if got := <-authorization; !strings.HasPrefix(got, "AWS4-HMAC-SHA256 ") {
		t.Errorf("Authorization header = %q, want SigV4 authentication", got)
	}
}
