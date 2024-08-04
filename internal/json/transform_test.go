// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestKeyFirstLower(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		input    string
		want     string
	}{
		{
			testName: "empty JSON",
			input:    "{}",
			want:     "{}",
		},
		{
			testName: "single field, lowercase",
			input:    `{ "key": 42 }`,
			want:     `{"key":42}`,
		},
		{
			testName: "single field, uppercase",
			input:    `{ "Key": 42 }`,
			want:     `{"key":42}`,
		},
		{
			testName: "multiple fields",
			input: `
[
  {
    "Name": "FIRST",
    "Image": "alpine",
    "Cpu": 10,
    "Memory": 512,
    "Essential": true,
    "PortMappings": [
      {
        "ContainerPort": 80,
        "HostPort": 80
      }
    ]
  }
]
			`,
			want: `[{"name":"FIRST","image":"alpine","cpu":10,"memory":512,"essential":true,"portMappings":[{"containerPort":80,"hostPort":80}]}]`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			if got, want := string(json.KeyFirstLower([]byte(testCase.input))), testCase.want; got != want {
				t.Errorf("KeyFirstLower(%q) = %q, want %q", testCase.input, got, want)
			}
		})
	}
}
