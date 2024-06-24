// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"testing"

	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
)

func TestEquivalentNodePropertiesJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ApiJson           string
		ConfigurationJson string
		ExpectEquivalent  bool
		ExpectError       bool
	}{
		"empty": {
			ApiJson:           ``,
			ConfigurationJson: ``,
			ExpectEquivalent:  true,
		},
		"Single Node with empty environment variable": {
			ApiJson: `
{
	"mainNode": 1,
	"nodeRangeProperties": [
		{
			"container":
			{
				"command": ["ls", "-la"],
				"image": "busybox",
				"memory":512
			},
			"targetNodes": "0:",
			"environment": []
		}
	],
	"numNodes": 2
}
`,
			ConfigurationJson: `
{
	"mainNode": 1,
	"nodeRangeProperties": [
		{
			"container":
			{
				"command": ["ls", "-la"],
				"image": "busybox",
				"memory":512,
				"environment": [
					{
						"name": "EMPTY",
						"value": ""
					}
				]
			},
			"targetNodes": "0:"
		}
	],
	"numNodes": 2
}
`,
			ExpectEquivalent: true,
		},
		"Two Nodes with empty command and mountPoints": {
			ApiJson: `
{
	"mainNode": 1,
	"nodeRangeProperties": [
		{
			"container":
			{
				"image": "busybox",
				"memory":512
			},
			"targetNodes": "0:",
			"environment": [],
			"mountPoints": []
		},
		{
			"container":
			{
				"image": "nginx",
				"memory":128
			},
			"targetNodes": "0:",
			"environment": [],
			"logConfiguration": {
				"logDriver": "awslogs",
				"secretOptions": []
			}
		}
	],
	"numNodes": 2
}
`,
			ConfigurationJson: `
{
	"mainNode": 1,
	"nodeRangeProperties": [
		{
			"container":
			{
				"command": [],
				"image": "busybox",
				"memory":512
			},
			"targetNodes": "0:",
			"environment": [],
			"mountPoints": []
		},
		{
			"container":
			{
				"image": "nginx",
				"memory":128
			},
			"targetNodes": "0:",
			"environment": [],
			"logConfiguration": {
				"logDriver": "awslogs"
			}
		}
	],
	"numNodes": 2
}
`,
			ExpectEquivalent: true,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tfbatch.EquivalentNodePropertiesJSON(testCase.ConfigurationJson, testCase.ApiJson)

			if err != nil && !testCase.ExpectError {
				t.Errorf("got unexpected error: %s", err)
			}

			if err == nil && testCase.ExpectError {
				t.Errorf("expected error, but received none")
			}

			if got != testCase.ExpectEquivalent {
				t.Errorf("got %t, expected %t", got, testCase.ExpectEquivalent)
			}
		})
	}
}
