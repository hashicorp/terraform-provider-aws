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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// go test -bench=BenchmarkSDKProviderInitialization -benchmem -run=Bench -v ./internal/provider
func BenchmarkSDKProviderInitialization(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := New(context.Background())
		if err != nil {
			b.Fatal(err)
		}
	}
}

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
	testcases := map[string]struct {
		endpoints        map[string]string
		envvars          map[string]string
		expectedService  string
		expectedEndpoint string
		expectedDiags    diag.Diagnostics
	}{
		"AWS_ENDPOINT_URL_STS envvar": {
			endpoints: map[string]string{},
			envvars: map[string]string{
				"AWS_ENDPOINT_URL_STS": "https://sts.fake.test",
			},
			expectedService:  names.STS,
			expectedEndpoint: "https://sts.fake.test",
		},
		"TF_AWS_STS_ENDPOINT envvar": {
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
		"AWS_STS_ENDPOINT envvar": {
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
		"STS envvar precedence": {
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
		"STS config precedence": {
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

	for name, testcase := range testcases { //nolint:paralleltest
		t.Run(name, func(t *testing.T) {
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
		})
	}
}

func TestExpandDefaultTags(t *testing.T) { //nolint:paralleltest
	ctx := context.Background()
	testcases := map[string]struct {
		tags                  map[string]interface{}
		envvars               map[string]string
		expectedDefaultConfig *tftags.DefaultConfig
	}{
		"nil": {
			tags:                  nil,
			envvars:               map[string]string{},
			expectedDefaultConfig: nil,
		},
		"envvar": {
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
		"config": {
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
		"envvar and config": {
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

	for name, testcase := range testcases { //nolint:paralleltest
		t.Run(name, func(t *testing.T) {
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
		})
	}
}

func TestExpandIgnoreTags(t *testing.T) { //nolint:paralleltest
	ctx := context.Background()
	testcases := map[string]struct {
		keys                 []interface{}
		keyPrefixes          []interface{}
		envvars              map[string]string
		expectedIgnoreConfig *tftags.IgnoreConfig
	}{
		"nil": {
			keys:                 nil,
			keyPrefixes:          nil,
			envvars:              map[string]string{},
			expectedIgnoreConfig: nil,
		},
		"envvar single": {
			envvars: map[string]string{
				tftags.IgnoreTagsKeysEnvVar:        "env1",
				tftags.IgnoreTagsKeyPrefixesEnvVar: "env2",
			},
			expectedIgnoreConfig: &tftags.IgnoreConfig{
				Keys:        tftags.New(ctx, []interface{}{"env1"}),
				KeyPrefixes: tftags.New(ctx, []interface{}{"env2"}),
			},
		},
		"envvar multiple": {
			envvars: map[string]string{
				tftags.IgnoreTagsKeysEnvVar:        "env1,env2",
				tftags.IgnoreTagsKeyPrefixesEnvVar: "env3,env4",
			},
			expectedIgnoreConfig: &tftags.IgnoreConfig{
				Keys:        tftags.New(ctx, []interface{}{"env1", "env2"}),
				KeyPrefixes: tftags.New(ctx, []interface{}{"env3", "env4"}),
			},
		},
		"envvar keys": {
			envvars: map[string]string{
				tftags.IgnoreTagsKeysEnvVar: "env1,env1",
			},
			expectedIgnoreConfig: &tftags.IgnoreConfig{
				Keys: tftags.New(ctx, []interface{}{"env1"}),
			},
		},
		"envvar key_prefixes": {
			envvars: map[string]string{
				tftags.IgnoreTagsKeyPrefixesEnvVar: "env1,env1",
			},
			expectedIgnoreConfig: &tftags.IgnoreConfig{
				KeyPrefixes: tftags.New(ctx, []interface{}{"env1"}),
			},
		},
		"config": {
			keys:        []interface{}{"config1", "config2"},
			keyPrefixes: []interface{}{"config3", "config4"},
			envvars:     map[string]string{},
			expectedIgnoreConfig: &tftags.IgnoreConfig{
				Keys:        tftags.New(ctx, []interface{}{"config1", "config2"}),
				KeyPrefixes: tftags.New(ctx, []interface{}{"config3", "config4"}),
			},
		},
		"envvar and config": {
			keys:        []interface{}{"config1", "config2"},
			keyPrefixes: []interface{}{"config3", "config4"},
			envvars: map[string]string{
				tftags.IgnoreTagsKeysEnvVar:        "env1,env2",
				tftags.IgnoreTagsKeyPrefixesEnvVar: "env3,env4",
			},
			expectedIgnoreConfig: &tftags.IgnoreConfig{
				Keys:        tftags.New(ctx, []interface{}{"env1", "env2", "config1", "config2"}),
				KeyPrefixes: tftags.New(ctx, []interface{}{"env3", "env4", "config3", "config4"}),
			},
		},
		"envvar and config keys duplicates": {
			keys: []interface{}{"example1", "example2"},
			envvars: map[string]string{
				tftags.IgnoreTagsKeysEnvVar: "example1,example3",
			},
			expectedIgnoreConfig: &tftags.IgnoreConfig{
				Keys: tftags.New(ctx, []interface{}{"example1", "example2", "example3"}),
			},
		},
		"envvar and config key_prefixes duplicates": {
			keyPrefixes: []interface{}{"example1", "example2"},
			envvars: map[string]string{
				tftags.IgnoreTagsKeyPrefixesEnvVar: "example1,example3",
			},
			expectedIgnoreConfig: &tftags.IgnoreConfig{
				KeyPrefixes: tftags.New(ctx, []interface{}{"example1", "example2", "example3"}),
			},
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest
		t.Run(name, func(t *testing.T) {
			oldEnv := stashEnv()
			defer popEnv(oldEnv)
			for k, v := range testcase.envvars {
				os.Setenv(k, v)
			}

			results := expandIgnoreTags(ctx, map[string]interface{}{
				"keys":         schema.NewSet(schema.HashString, testcase.keys),
				"key_prefixes": schema.NewSet(schema.HashString, testcase.keyPrefixes),
			})

			if results == nil && testcase.expectedIgnoreConfig != nil {
				t.Errorf("Expected ignore tags config to be %v, got nil", testcase.expectedIgnoreConfig)
			}

			if diff := cmp.Diff(testcase.expectedIgnoreConfig, results); diff != "" {
				t.Errorf("Unexpected ignore_tags diff: %s", diff)
			}
		})
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
