// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns_test

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"testing"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

type proxyCase struct {
	url           string
	expectedProxy string
}

func TestProxyConfig(t *testing.T) {
	cases := map[string]struct {
		config               map[string]any
		environmentVariables map[string]string
		expectedDiags        diag.Diagnostics
		urls                 []proxyCase
	}{
		"no config": {
			config: map[string]any{},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"http_proxy empty string": {
			config: map[string]any{
				"http_proxy": "",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"http_proxy config": {
			config: map[string]any{
				"http_proxy": "http://http-proxy.test:1234",
			},
			expectedDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Missing HTTPS Proxy",
					Detail: fmt.Sprintf(
						"An HTTP proxy was set but no HTTPS proxy was. Using HTTP proxy %q for HTTPS requests. This behavior may change in future versions.\n\n"+
							"To specify no proxy for HTTPS, set the HTTPS to an empty string.",
						"http://http-proxy.test:1234"),
				},
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
			},
		},

		"https_proxy config": {
			config: map[string]any{
				"https_proxy": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"http_proxy config https_proxy config": {
			config: map[string]any{
				"http_proxy":  "http://http-proxy.test:1234",
				"https_proxy": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"http_proxy config https_proxy config empty string": {
			config: map[string]any{
				"http_proxy":  "http://http-proxy.test:1234",
				"https_proxy": "",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"https_proxy config http_proxy config empty string": {
			config: map[string]any{
				"http_proxy":  "",
				"https_proxy": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"http_proxy config https_proxy config no_proxy config": {
			config: map[string]any{
				"http_proxy":  "http://http-proxy.test:1234",
				"https_proxy": "http://https-proxy.test:1234",
				"no_proxy":    "dont-proxy.test",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "http://dont-proxy.test",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
				{
					url:           "https://dont-proxy.test",
					expectedProxy: "",
				},
			},
		},

		"HTTP_PROXY envvar": {
			config: map[string]any{},
			environmentVariables: map[string]string{
				"HTTP_PROXY": "http://http-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"http_proxy envvar": {
			config: map[string]any{},
			environmentVariables: map[string]string{
				"http_proxy": "http://http-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"HTTPS_PROXY envvar": {
			config: map[string]any{},
			environmentVariables: map[string]string{
				"HTTPS_PROXY": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"https_proxy envvar": {
			config: map[string]any{},
			environmentVariables: map[string]string{
				"https_proxy": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"http_proxy config HTTPS_PROXY envvar": {
			config: map[string]any{
				"http_proxy": "http://http-proxy.test:1234",
			},
			environmentVariables: map[string]string{
				"HTTPS_PROXY": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"http_proxy config https_proxy envvar": {
			config: map[string]any{
				"http_proxy": "http://http-proxy.test:1234",
			},
			environmentVariables: map[string]string{
				"https_proxy": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"http_proxy config NO_PROXY envvar": {
			config: map[string]any{
				"http_proxy": "http://http-proxy.test:1234",
			},
			environmentVariables: map[string]string{
				"NO_PROXY": "dont-proxy.test",
			},
			expectedDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Missing HTTPS Proxy",
					Detail: fmt.Sprintf(
						"An HTTP proxy was set but no HTTPS proxy was. Using HTTP proxy %q for HTTPS requests. This behavior may change in future versions.\n\n"+
							"To specify no proxy for HTTPS, set the HTTPS to an empty string.",
						"http://http-proxy.test:1234"),
				},
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "http://dont-proxy.test",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://dont-proxy.test",
					expectedProxy: "",
				},
			},
		},

		"http_proxy config no_proxy envvar": {
			config: map[string]any{
				"http_proxy": "http://http-proxy.test:1234",
			},
			environmentVariables: map[string]string{
				"no_proxy": "dont-proxy.test",
			},
			expectedDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Missing HTTPS Proxy",
					Detail: fmt.Sprintf(
						"An HTTP proxy was set but no HTTPS proxy was. Using HTTP proxy %q for HTTPS requests. This behavior may change in future versions.\n\n"+
							"To specify no proxy for HTTPS, set the HTTPS to an empty string.",
						"http://http-proxy.test:1234"),
				},
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "http://dont-proxy.test",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://dont-proxy.test",
					expectedProxy: "",
				},
			},
		},

		"HTTP_PROXY envvar HTTPS_PROXY envvar NO_PROXY envvar": {
			config: map[string]any{},
			environmentVariables: map[string]string{
				"HTTP_PROXY":  "http://http-proxy.test:1234",
				"HTTPS_PROXY": "http://https-proxy.test:1234",
				"NO_PROXY":    "dont-proxy.test",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "http://dont-proxy.test",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
				{
					url:           "https://dont-proxy.test",
					expectedProxy: "",
				},
			},
		},

		"http_proxy config overrides HTTP_PROXY envvar": {
			config: map[string]any{
				"http_proxy": "http://config-proxy.test:1234",
			},
			environmentVariables: map[string]string{
				"HTTP_PROXY": "http://envvar-proxy.test:1234",
			},
			expectedDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Missing HTTPS Proxy",
					Detail: fmt.Sprintf(
						"An HTTP proxy was set but no HTTPS proxy was. Using HTTP proxy %q for HTTPS requests. This behavior may change in future versions.\n\n"+
							"To specify no proxy for HTTPS, set the HTTPS to an empty string.",
						"http://config-proxy.test:1234"),
				},
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://config-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://config-proxy.test:1234",
				},
			},
		},

		"https_proxy config overrides HTTPS_PROXY envvar": {
			config: map[string]any{
				"https_proxy": "http://config-proxy.test:1234",
			},
			environmentVariables: map[string]string{
				"HTTPS_PROXY": "http://envvar-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://config-proxy.test:1234",
				},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			config := map[string]any{
				"access_key":                  "StaticAccessKey",
				"secret_key":                  servicemocks.MockStaticSecretKey,
				"region":                      "us-west-2",
				"skip_credentials_validation": true,
				"skip_requesting_account_id":  true,
			}

			for k, v := range tc.environmentVariables {
				t.Setenv(k, v)
			}

			maps.Copy(config, tc.config)

			p, err := provider.New(ctx)
			if err != nil {
				t.Fatal(err)
			}

			expectedDiags := tc.expectedDiags
			expectedDiags = append(
				expectedDiags,
				errs.NewWarningDiagnostic(
					"AWS account ID not found for provider",
					"See https://registry.terraform.io/providers/hashicorp/aws/latest/docs#skip_requesting_account_id for implications.",
				),
			)

			diags := p.Configure(ctx, terraformsdk.NewResourceConfigRaw(config))

			if diff := cmp.Diff(diags, expectedDiags, cmp.Comparer(sdkdiag.Comparer)); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			meta := p.Meta().(*conns.AWSClient)

			client := meta.AwsConfig(ctx).HTTPClient
			bClient, ok := client.(*awshttp.BuildableClient)
			if !ok {
				t.Fatalf("expected awshttp.BuildableClient, got %T", client)
			}
			transport := bClient.GetTransport()
			proxyF := transport.Proxy

			for _, url := range tc.urls {
				req, _ := http.NewRequest(http.MethodGet, url.url, nil)
				pUrl, err := proxyF(req)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if url.expectedProxy != "" {
					if pUrl == nil {
						t.Errorf("expected proxy for %q, got none", url.url)
					} else if pUrl.String() != url.expectedProxy {
						t.Errorf("expected proxy %q for %q, got %q", url.expectedProxy, url.url, pUrl.String())
					}
				} else {
					if pUrl != nil {
						t.Errorf("expected no proxy for %q, got %q", url.url, pUrl.String())
					}
				}
			}
		})
	}
}
