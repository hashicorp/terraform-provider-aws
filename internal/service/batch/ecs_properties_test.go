// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
)

func TestEquivalentECSPropertiesJSON(t *testing.T) {
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
		"reordered containers": {
			apiJSON: `
{
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
  }
			`,
			configurationJSON: `
{
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
  }
			`,
			wantEquivalent: true,
		},
		"reordered environment": {
			apiJSON: `
{
  "taskProperties": [
    {
      "containers": [
        {
          "name": "container1",
          "image": "my_ecr_image1",
          "environment": [
            {
              "name": "VARNAME1",
              "value": "VARVAL1"
            },
            {
              "name": "VARNAME2",
              "value": "VARVAL2"
            }
          ]
        },
        {
          "name": "container2",
          "image": "my_ecr_image2",
          "environment": []
        }
      ]
    }
  ]
}
			`,
			configurationJSON: `
{
  "taskProperties": [
    {
      "containers": [
        {
          "name": "container1",
          "image": "my_ecr_image1",
          "environment": [
            {
              "name": "VARNAME2",
              "value": "VARVAL2"
            },
            {
              "name": "VARNAME1",
              "value": "VARVAL1"
            }
          ]
        },
        {
          "name": "container2",
          "image": "my_ecr_image2"
        }
      ]
    }
  ]
}
			`,
			wantEquivalent: true,
		},
		"full": {
			apiJSON: `
{
  "taskProperties": [
    {
      "containers": [
        {
          "command": [
            "sleep",
            "60"
          ],
          "dependsOn": [
            {
              "condition": "COMPLETE",
              "containerName": "container_b"
            }
          ],
          "environment": [
            {
              "name": "test",
              "value": "Environment Variable"
            }
          ],
          "essential": true,
          "image": "public.ecr.aws/amazonlinux/amazonlinux:1",
          "logConfiguration": {
            "logDriver": "awslogs",
            "options": {
              "awslogs-stream-prefix": "ecs",
              "awslogs-group": "ewbankkit-test-003",
              "awslogs-region": "region-2"
            },
            "secretOptions": []
          },
          "mountPoints": [],
          "name": "container_a",
          "privileged": false,
          "readonlyRootFilesystem": false,
          "resourceRequirements": [
            {
              "type": "VCPU",
              "value": "1.0"
            },
            {
              "type": "MEMORY",
              "value": "2048"
            }
          ],
          "secrets": [
            {
              "name": "TEST",
              "valueFrom": "DUMMY"
            }
          ],
          "ulimits": []
        },
        {
          "command": [
            "sleep",
            "360"
          ],
          "dependsOn": [],
          "environment": [],
          "essential": false,
          "image": "public.ecr.aws/amazonlinux/amazonlinux:1",
          "mountPoints": [],
          "name": "container_b",
          "resourceRequirements": [
            {
              "type": "VCPU",
              "value": "1.0"
            },
            {
              "type": "MEMORY",
              "value": "2048"
            }
          ],
          "secrets": [],
          "ulimits": []
        }
      ],
      "executionRoleArn": "role1",
      "platformVersion": "LATEST",
      "volumes": []
    }
  ]
}
      `,
			configurationJSON: `
{
  "taskProperties": [
    {
      "containers": [
        {
          "command": [
            "sleep",
            "60"
          ],
          "dependsOn": [
            {
              "condition": "COMPLETE",
              "containerName": "container_b"
            }
          ],
          "environment": [
            {
              "name": "test",
              "value": "Environment Variable"
            }
          ],
          "essential": true,
          "image": "public.ecr.aws/amazonlinux/amazonlinux:1",
          "logConfiguration": {
            "logDriver": "awslogs",
            "options": {
              "awslogs-group": "ewbankkit-test-003",
              "awslogs-region": "region-2",
              "awslogs-stream-prefix": "ecs"
            }
          },
          "name": "container_a",
          "privileged": false,
          "readonlyRootFilesystem": false,
          "resourceRequirements": [
            {
              "type": "VCPU",
              "value": "1.0"
            },
            {
              "type": "MEMORY",
              "value": "2048"
            }
          ],
          "secrets": [
            {
              "name": "TEST",
              "valueFrom": "DUMMY"
            }
          ]
        },
        {
          "command": [
            "sleep",
            "360"
          ],
          "essential": false,
          "image": "public.ecr.aws/amazonlinux/amazonlinux:1",
          "name": "container_b",
          "resourceRequirements": [
            {
              "type": "VCPU",
              "value": "1.0"
            },
            {
              "type": "MEMORY",
              "value": "2048"
            }
          ]
        }
      ],
      "executionRoleArn": "role1"
    }
  ]
}
      `,
			wantEquivalent: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			output, err := tfbatch.EquivalentECSPropertiesJSON(testCase.configurationJSON, testCase.apiJSON)
			if got, want := err != nil, testCase.wantErr; !cmp.Equal(got, want) {
				t.Errorf("EquivalentECSPropertiesJSON err %t, want %t", got, want)
			}

			if err == nil {
				if got, want := output, testCase.wantEquivalent; !cmp.Equal(got, want) {
					t.Errorf("EquivalentECSPropertiesJSON equivalent %t, want %t", got, want)
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
