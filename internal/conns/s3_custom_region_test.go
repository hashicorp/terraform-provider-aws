// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package conns_test

import (
	"context"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2"
)

func TestConfigureProvider_s3CustomRegion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		region                 string
		customS3Endpoint       string
		expectSubstitution     bool
		expectedOriginalRegion string
	}{
		"standard region no custom endpoint": {
			region:                 "us-east-1",
			customS3Endpoint:       "",
			expectSubstitution:     false,
			expectedOriginalRegion: "",
		},
		"standard region with custom endpoint": {
			region:                 "us-west-2",
			customS3Endpoint:       "https://s3.example.com",
			expectSubstitution:     false,
			expectedOriginalRegion: "",
		},
		"non-standard region with colon": {
			region:                 "ceph-objectstore:region-premium",
			customS3Endpoint:       "https://ceph.example.com",
			expectSubstitution:     true,
			expectedOriginalRegion: "ceph-objectstore:region-premium",
		},
		"non-standard region with uppercase": {
			region:                 "CUSTOM-REGION",
			customS3Endpoint:       "https://minio.example.com",
			expectSubstitution:     true,
			expectedOriginalRegion: "CUSTOM-REGION",
		},
		"non-standard region with underscore": {
			region:                 "custom_region_1",
			customS3Endpoint:       "https://s3-compat.example.com",
			expectSubstitution:     true,
			expectedOriginalRegion: "custom_region_1",
		},
		"non-standard region without custom endpoint": {
			region:                 "ceph:region",
			customS3Endpoint:       "",
			expectSubstitution:     false,
			expectedOriginalRegion: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := map[string]any{
				"access_key":                  "StaticAccessKey",
				"secret_key":                  servicemocks.MockStaticSecretKey,
				"region":                      tc.region,
				"skip_credentials_validation": true,
				"skip_requesting_account_id":  true,
				"skip_region_validation":      true,
			}

			if tc.customS3Endpoint != "" {
				config["endpoints"] = []any{
					map[string]any{
						"s3": tc.customS3Endpoint,
					},
				}
			}

			p, err := sdkv2.NewProvider(ctx)
			if err != nil {
				t.Fatal(err)
			}

			p.TerraformVersion = "1.0.0"

			diags := p.Configure(ctx, terraformsdk.NewResourceConfigRaw(config))
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			meta := p.Meta().(*conns.AWSClient)

			// Check if region substitution occurred
			actualRegion := meta.Region(ctx)
			if tc.expectSubstitution {
				if actualRegion != "us-east-1" {
					t.Errorf("expected region to be substituted to 'us-east-1', got %q", actualRegion)
				}
			} else {
				if actualRegion != tc.region {
					t.Errorf("expected region to remain %q, got %q", tc.region, actualRegion)
				}
			}

			// Verify S3 client can be created without panic
			// This is the key test - it would panic before the fix
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("S3Client() panicked: %v", r)
					}
				}()
				_ = meta.S3Client(ctx)
			}()
		})
	}
}
