// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
)

func TestEquivalentNodePropertiesJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		apiJSON           string
		configurationJSON string
		wantEquivalent    bool
		wantErr           bool
	}{
		"empty": {
			apiJSON:           ``,
			configurationJSON: ``,
			wantEquivalent:    true,
		},
		"Single Node with empty environment variable": {
			apiJSON: `
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
			configurationJSON: `
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
			wantEquivalent: true,
		},
		"Two Nodes with empty command and mountPoints": {
			apiJSON: `
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
			configurationJSON: `
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
			wantEquivalent: true,
		},
		"Single node ECS Properties with multiple containers": {
			apiJSON: `
{
	"mainNode": 1,
	"nodeRangeProperties": [
		{
			"ecsProperties": {
				"taskProperties": [
				{
					"containers": [
					{
						"name": "container1",
						"image": "my_ecr_image1"
					},
					{
						"name": "container2",
						"image": "my_ecr_image2"
					}
					]
				}
				]
			},
			"targetNodes": "0:",
			"environment": [],
			"mountPoints": []
		}
	],
	"numNodes": 1
}
`,
			configurationJSON: `
{
  "mainNode": 1,
  "nodeRangeProperties": [
    {
      "ecsProperties": {
        "taskProperties": [
          {
            "containers": [
              {
                "name": "container2",
                "image": "my_ecr_image2"
              },
              {
                "name": "container1",
                "image": "my_ecr_image1"
              }
            ]
          }
        ]
      },
      "targetNodes": "0:",
      "environment": [],
      "mountPoints": []
    }
  ],
  "numNodes": 1
}

`,
			wantEquivalent: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			output, err := tfbatch.EquivalentNodePropertiesJSON(testCase.configurationJSON, testCase.apiJSON)
			if got, want := err != nil, testCase.wantErr; !cmp.Equal(got, want) {
				t.Errorf("EquivalentNodePropertiesJSON err %t, want %t", got, want)
			}

			if err == nil {
				if got, want := output, testCase.wantEquivalent; !cmp.Equal(got, want) {
					t.Errorf("EquivalentNodePropertiesJSON equivalent %t, want %t", got, want)
					if want {
						if diff := jsoncmp.Diff(testCase.configurationJSON, testCase.apiJSON); diff != "" {
							t.Errorf("unexpected diff (+wanted, -got): %s", diff)
						}
					}
				}
			}
		})
	}
}
