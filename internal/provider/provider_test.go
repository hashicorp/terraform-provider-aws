// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestProvider(t *testing.T) {
	t.Parallel()

	p, err := New(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	err = p.InternalValidate()

	if err != nil {
		t.Fatal(err)
	}
}

func TestExpandEndpoints(t *testing.T) { //nolint:paralleltest
	oldEnv := stashEnv()
	defer popEnv(oldEnv)

	var expectedDiags diag.Diagnostics

	ctx := context.Background()
	endpoints := make(map[string]interface{})
	for _, serviceKey := range names.Aliases() {
		endpoints[serviceKey] = ""
	}
	endpoints["sts"] = "https://sts.fake.test"

	results, diags := expandEndpoints(ctx, []interface{}{endpoints})
	if diff := cmp.Diff(diags, expectedDiags, cmp.Comparer(sdkdiag.Comparer)); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(results))
	}

	if v := results["sts"]; v != "https://sts.fake.test" {
		t.Errorf("Expected endpoint %q, got %v", "https://sts.fake.test", results)
	}
}

func TestEndpointMultipleKeys(t *testing.T) { //nolint:paralleltest
	ctx := context.Background()
	testcases := []struct {
		endpoints        map[string]string
		expectedService  string
		expectedEndpoint string
		expectedDiags    diag.Diagnostics
	}{
		{
			endpoints: map[string]string{
				"transcribe": "https://transcribe.fake.test",
			},
			expectedService:  names.Transcribe,
			expectedEndpoint: "https://transcribe.fake.test",
		},
		{
			endpoints: map[string]string{
				"transcribeservice": "https://transcribe.fake.test",
			},
			expectedService:  names.Transcribe,
			expectedEndpoint: "https://transcribe.fake.test",
		},
		{
			endpoints: map[string]string{
				"transcribe":        "https://transcribe.fake.test",
				"transcribeservice": "https://transcribeservice.fake.test",
			},
			expectedService:  names.Transcribe,
			expectedEndpoint: "https://transcribe.fake.test",
			expectedDiags: diag.Diagnostics{ConflictingEndpointsWarningDiag(
				cty.GetAttrPath("endpoints").IndexInt(0),
				"transcribe",
				"transcribeservice",
			)},
		},
	}

	for _, testcase := range testcases {
		oldEnv := stashEnv()
		defer popEnv(oldEnv)

		endpoints := make(map[string]interface{})
		for _, serviceKey := range names.Aliases() {
			endpoints[serviceKey] = ""
		}
		for k, v := range testcase.endpoints {
			endpoints[k] = v
		}

		results, diags := expandEndpoints(ctx, []interface{}{endpoints})
		if diff := cmp.Diff(diags, testcase.expectedDiags, cmp.Comparer(sdkdiag.Comparer)); diff != "" {
			t.Errorf("unexpected diagnostics difference: %s", diff)
		}

		if a, e := len(results), 1; a != e {
			t.Errorf("Expected 1 endpoint, got %d", len(results))
		}

		if v := results[testcase.expectedService]; v != testcase.expectedEndpoint {
			t.Errorf("Expected endpoint[%s] to be %q, got %v", testcase.expectedService, testcase.expectedEndpoint, results)
		}
	}
}

func TestEndpointEnvVarPrecedence(t *testing.T) { //nolint:paralleltest
	ctx := context.Background()
	testcases := []struct {
		endpoints        map[string]string
		envvars          map[string]string
		expectedService  string
		expectedEndpoint string
		expectedDiags    diag.Diagnostics
	}{
		{
			endpoints: map[string]string{},
			envvars: map[string]string{
				"AWS_ENDPOINT_URL_STS": "https://sts.fake.test",
			},
			expectedService:  names.STS,
			expectedEndpoint: "https://sts.fake.test",
		},
		{
			endpoints: map[string]string{},
			envvars: map[string]string{
				"TF_AWS_STS_ENDPOINT": "https://sts.fake.test",
			},
			expectedService:  names.STS,
			expectedEndpoint: "https://sts.fake.test",
			expectedDiags: diag.Diagnostics{
				DeprecatedEnvVarDiag("TF_AWS_STS_ENDPOINT", "AWS_ENDPOINT_URL_STS"),
			},
		},
		{
			endpoints: map[string]string{},
			envvars: map[string]string{
				"AWS_STS_ENDPOINT": "https://sts-deprecated.fake.test",
			},
			expectedService:  names.STS,
			expectedEndpoint: "https://sts-deprecated.fake.test",
			expectedDiags: diag.Diagnostics{
				DeprecatedEnvVarDiag("AWS_STS_ENDPOINT", "AWS_ENDPOINT_URL_STS"),
			},
		},
		{
			endpoints: map[string]string{},
			envvars: map[string]string{
				"TF_AWS_STS_ENDPOINT": "https://sts.fake.test",
				"AWS_STS_ENDPOINT":    "https://sts-deprecated.fake.test",
			},
			expectedService:  names.STS,
			expectedEndpoint: "https://sts.fake.test",
			expectedDiags: diag.Diagnostics{
				DeprecatedEnvVarDiag("TF_AWS_STS_ENDPOINT", "AWS_ENDPOINT_URL_STS"),
			},
		},
		{
			endpoints: map[string]string{
				"sts": "https://sts-config.fake.test",
			},
			envvars: map[string]string{
				"TF_AWS_STS_ENDPOINT": "https://sts-env.fake.test",
			},
			expectedService:  names.STS,
			expectedEndpoint: "https://sts-config.fake.test",
		},
	}

	for _, testcase := range testcases {
		oldEnv := stashEnv()
		defer popEnv(oldEnv)

		for k, v := range testcase.envvars {
			os.Setenv(k, v)
		}

		endpoints := make(map[string]interface{})
		for _, serviceKey := range names.Aliases() {
			endpoints[serviceKey] = ""
		}
		for k, v := range testcase.endpoints {
			endpoints[k] = v
		}

		results, diags := expandEndpoints(ctx, []interface{}{endpoints})
		if diff := cmp.Diff(diags, testcase.expectedDiags, cmp.Comparer(sdkdiag.Comparer)); diff != "" {
			t.Errorf("unexpected diagnostics difference: %s", diff)
		}

		if a, e := len(results), 1; a != e {
			t.Errorf("Expected 1 endpoint, got %d", len(results))
		}

		if v := results[testcase.expectedService]; v != testcase.expectedEndpoint {
			t.Errorf("Expected endpoint[%s] to be %q, got %v", testcase.expectedService, testcase.expectedEndpoint, results)
		}
	}
}

func TestExpandDefaultTags(t *testing.T) { //nolint:paralleltest
	ctx := context.Background()
	testcases := []struct {
		tags                  map[string]interface{}
		envvars               map[string]string
		expectedDefaultConfig *tftags.DefaultConfig
	}{
		{
			tags:                  nil,
			envvars:               map[string]string{},
			expectedDefaultConfig: nil,
		},
		{
			tags: nil,
			envvars: map[string]string{
				tftags.DefaultTagsEnvVarPrefix + "Owner": "my-team",
			},
			expectedDefaultConfig: &tftags.DefaultConfig{
				Tags: tftags.New(ctx, map[string]string{
					"Owner": "my-team",
				}),
			},
		},
		{
			tags: map[string]interface{}{
				"Owner": "my-team",
			},
			envvars: map[string]string{},
			expectedDefaultConfig: &tftags.DefaultConfig{
				Tags: tftags.New(ctx, map[string]string{
					"Owner": "my-team",
				}),
			},
		},
		{
			tags: map[string]interface{}{
				"Application": "foobar",
			},
			envvars: map[string]string{
				tftags.DefaultTagsEnvVarPrefix + "Application": "my-app",
				tftags.DefaultTagsEnvVarPrefix + "Owner":       "my-team",
			},
			expectedDefaultConfig: &tftags.DefaultConfig{
				Tags: tftags.New(ctx, map[string]string{
					"Application": "foobar",
					"Owner":       "my-team",
				}),
			},
		},
	}

	for _, testcase := range testcases {
		oldEnv := stashEnv()
		defer popEnv(oldEnv)
		for k, v := range testcase.envvars {
			os.Setenv(k, v)
		}

		results := expandDefaultTags(ctx, map[string]interface{}{
			"tags": testcase.tags,
		})

		if results == nil {
			if testcase.expectedDefaultConfig == nil {
				return
			} else {
				t.Errorf("Expected default tags config to be %v, got nil", testcase.expectedDefaultConfig)
			}
		} else if !testcase.expectedDefaultConfig.TagsEqual(results.Tags) {
			t.Errorf("Expected default tags config to be %v, got %v", testcase.expectedDefaultConfig, results)
		}
	}
}

func stashEnv() []string {
	env := os.Environ()
	os.Clearenv()
	return env
}

func popEnv(env []string) {
	os.Clearenv()

	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		os.Setenv(k, v)
	}
}
