// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicediscovery_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicediscovery "github.com/hashicorp/terraform-provider-aws/internal/service/servicediscovery"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestServiceCustomHealthConfig(t *testing.T) {
	t.Parallel()

	for name, testCase := range map[string]struct {
		config []any
		want   string
	}{
		"absent": {},
		"empty block": {
			config: []any{nil},
			want:   `{}`,
		},
		"empty object": {
			config: []any{map[string]any{}},
			want:   `{}`,
		},
		"null threshold": {
			config: []any{map[string]any{"failure_threshold": nil}},
			want:   `{}`,
		},
		"explicit threshold": {
			config: []any{map[string]any{"failure_threshold": 1}},
			want:   `{"FailureThreshold":1}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			requests := make(chan string, 1)
			service := map[string]any{
				"Id":          "srv-test",
				"Arn":         "arn:aws:servicediscovery:us-east-1:123456789012:service/srv-test",
				"Name":        "test",
				"NamespaceId": "ns-test",
				"Type":        "HTTP",
			}
			if testCase.want != "" {
				service["HealthCheckCustomConfig"] = map[string]any{"FailureThreshold": 1}
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.Header.Get("X-Amz-Target"), ".CreateService") {
					var input map[string]json.RawMessage
					if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
						t.Errorf("decoding CreateService request: %s", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					requests <- string(input["HealthCheckCustomConfig"])
				} else if !strings.HasSuffix(r.Header.Get("X-Amz-Target"), ".GetService") {
					t.Errorf("unexpected operation: %s", r.Header.Get("X-Amz-Target"))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/x-amz-json-1.1")
				if err := json.NewEncoder(w).Encode(map[string]any{"Service": service}); err != nil {
					t.Errorf("encoding service response: %s", err)
				}
			}))
			t.Cleanup(server.Close)

			client := &conns.AWSClient{}
			client.SetServicePackages(ctx, map[string]conns.ServicePackage{
				names.ServiceDiscovery: tfservicediscovery.ServicePackage(ctx),
			})
			configDir := t.TempDir()
			providerConfig := conns.Config{
				AccessKey:               "test",
				SecretKey:               "test",
				Region:                  "us-east-1",
				Endpoints:               map[string]string{names.ServiceDiscovery: server.URL},
				SkipCredsValidation:     true,
				SkipRequestingAccountId: true,
				SharedConfigFiles:       []string{filepath.Join(configDir, "config")},
				SharedCredentialsFiles:  []string{filepath.Join(configDir, "credentials")},
			}
			client, diags := providerConfig.ConfigureProvider(ctx, client)
			if diags.HasError() {
				t.Fatalf("configuring provider: %v", diags)
			}

			r := tfservicediscovery.ResourceService()
			rawConfig := map[string]any{
				names.AttrName: "test",
				"namespace_id": "ns-test",
				names.AttrType: "HTTP",
			}
			if testCase.config != nil {
				rawConfig["health_check_custom_config"] = testCase.config
			}
			diff, err := r.Diff(ctx, nil, terraform.NewResourceConfigRaw(rawConfig), client)
			if err != nil {
				t.Fatalf("planning service: %s", err)
			}
			state, diags := r.Apply(ctx, nil, diff, client)
			if diags.HasError() {
				t.Fatalf("creating service: %v", diags)
			}
			if got := <-requests; got != testCase.want {
				t.Errorf("HealthCheckCustomConfig = %q, want %q", got, testCase.want)
			}
			state, diags = r.RefreshWithoutUpgrade(ctx, state, client)
			if diags.HasError() {
				t.Fatalf("refreshing service: %v", diags)
			}
			if testCase.want != "" {
				if got := state.Attributes["health_check_custom_config.0.failure_threshold"]; got != "1" {
					t.Errorf("refreshed failure_threshold = %q, want 1", got)
				}
			}
			diff, err = r.Diff(ctx, state, terraform.NewResourceConfigRaw(rawConfig), client)
			if err != nil {
				t.Fatalf("planning after refresh: %s", err)
			}
			// Provider-level interceptors manage tags_all; this component test checks custom health.
			if diff != nil {
				for key, attr := range diff.Attributes {
					if strings.HasPrefix(key, "health_check_custom_config.") {
						t.Errorf("unexpected custom health diff after refresh: %s = %#v", key, attr)
					}
				}
			}
			if testCase.want != "" {
				rawConfig["health_check_custom_config"] = []any{map[string]any{}}
				diff, err = r.Diff(ctx, state, terraform.NewResourceConfigRaw(rawConfig), client)
				if err != nil {
					t.Fatalf("planning empty custom health config: %s", err)
				}
				if diff != nil {
					for key, attr := range diff.Attributes {
						if strings.HasPrefix(key, "health_check_custom_config.") {
							t.Errorf("unexpected custom health diff when omitting threshold: %s = %#v", key, attr)
						}
					}
				}
				delete(rawConfig, "health_check_custom_config")
			} else {
				rawConfig["health_check_custom_config"] = []any{map[string]any{}}
			}
			diff, err = r.Diff(ctx, state, terraform.NewResourceConfigRaw(rawConfig), client)
			if err != nil {
				t.Fatalf("planning custom health configuration change: %s", err)
			}
			if diff == nil || !diff.RequiresNew() {
				t.Errorf("adding or removing custom health config must replace the service: %#v", diff)
			}
		})
	}
}
