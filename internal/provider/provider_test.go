// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"strings"
	"testing"

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

	ctx := context.Background()
	endpoints := make(map[string]interface{})
	for _, serviceKey := range names.Aliases() {
		endpoints[serviceKey] = ""
	}
	endpoints["sts"] = "https://sts.fake.test"

	results, err := expandEndpoints(ctx, []interface{}{endpoints})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
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

		results, err := expandEndpoints(ctx, []interface{}{endpoints})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
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
	}{
		{
			endpoints: map[string]string{},
			envvars: map[string]string{
				"TF_AWS_STS_ENDPOINT": "https://sts.fake.test",
			},
			expectedService:  names.STS,
			expectedEndpoint: "https://sts.fake.test",
		},
		{
			endpoints: map[string]string{},
			envvars: map[string]string{
				"AWS_STS_ENDPOINT": "https://sts-deprecated.fake.test",
			},
			expectedService:  names.STS,
			expectedEndpoint: "https://sts-deprecated.fake.test",
		},
		{
			endpoints: map[string]string{},
			envvars: map[string]string{
				"TF_AWS_STS_ENDPOINT": "https://sts.fake.test",
				"AWS_STS_ENDPOINT":    "https://sts-deprecated.fake.test",
			},
			expectedService:  names.STS,
			expectedEndpoint: "https://sts.fake.test",
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

		results, err := expandEndpoints(ctx, []interface{}{endpoints})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		if a, e := len(results), 1; a != e {
			t.Errorf("Expected 1 endpoint, got %d", len(results))
		}

		if v := results[testcase.expectedService]; v != testcase.expectedEndpoint {
			t.Errorf("Expected endpoint[%s] to be %q, got %v", testcase.expectedService, testcase.expectedEndpoint, results)
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
		p := strings.SplitN(e, "=", 2)
		k, v := p[0], ""
		if len(p) > 1 {
			v = p[1]
		}
		os.Setenv(k, v)
	}
}
