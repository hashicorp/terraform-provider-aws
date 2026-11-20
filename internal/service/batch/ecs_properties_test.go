// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"encoding/json"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
)

func TestFlattenECSProperties(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"nil":   {input: "null", want: ""},
		"empty": {input: "{}", want: "{}"},
		"all fields": {
			input: `{"taskProperties":[{
				"containers":[{
					"command":["sh","-c","echo \"hello\" <&>\n"],
					"dependsOn":[{"condition":"START","containerName":"other"}],
					"environment":[{"name":"Z","value":"last"},{"name":"A","value":""}],
					"essential":true,
					"firelensConfiguration":{"options":{"MixedCase.Key":"<&>","empty":""},"type":"fluentbit"},
					"image":"image",
					"linuxParameters":{"devices":[{"containerPath":"/dev/x","hostPath":"/dev/y","permissions":["READ"]}],"initProcessEnabled":true,"maxSwap":1024,"sharedMemorySize":64,"swappiness":60,"tmpfs":[{"containerPath":"/tmp","mountOptions":["rw"],"size":128}]},
					"logConfiguration":{"logDriver":"awslogs","options":{"key":"value"},"secretOptions":[{"name":"log-secret","valueFrom":"log-value"}]},
					"mountPoints":[{"containerPath":"/data","readOnly":true,"sourceVolume":"data"}],
					"name":"main","privileged":true,"readonlyRootFilesystem":true,
					"repositoryCredentials":{"credentialsParameter":"credentials"},
					"resourceRequirements":[{"type":"VCPU","value":"1"}],
					"secrets":[{"name":"Z","valueFrom":"last"},{"name":"A","valueFrom":"first"}],
					"startTimeout":10,"stopTimeout":20,"ulimits":[{"hardLimit":1024,"name":"nofile","softLimit":512}],"user":"1000"
				},{"name":"other","image":"other-image"}],
				"enableExecuteCommand":true,"ephemeralStorage":{"sizeInGiB":21},"executionRoleArn":"execution-role",
				"ipcMode":"task","networkConfiguration":{"assignPublicIp":"ENABLED"},"networkMode":"host","pidMode":"task",
				"platformVersion":"LATEST","runtimePlatform":{"cpuArchitecture":"ARM64","operatingSystemFamily":"LINUX"},
				"taskRoleArn":"task-role","volumes":[{"name":"data","host":{"sourcePath":"/data"}}]
			},{"containers":[{"name":"second-task"}]}]}`,
		},
		"explicit zero values": {
			input: `{"taskProperties":[{
				"containers":[{"command":[""],"dependsOn":[{"condition":"","containerName":""}],"environment":[{"name":"","value":""}],
					"essential":false,"firelensConfiguration":{"options":{"":""}},"image":"","name":"","privileged":false,"readonlyRootFilesystem":false,
					"startTimeout":0,"stopTimeout":0,"user":""}],
				"enableExecuteCommand":false,"executionRoleArn":"","ipcMode":"","networkMode":"","pidMode":"","platformVersion":"","taskRoleArn":""
			}]}`,
		},
		"empty tasks": {input: `{"taskProperties":[]}`},
		"empty collections": {
			input: `{"taskProperties":[{},{"containers":[],"ephemeralStorage":{},"networkConfiguration":{},"runtimePlatform":{},"volumes":[]},{"containers":[{
				"command":[],"dependsOn":[],"environment":[],"firelensConfiguration":{"options":{}},
				"linuxParameters":{},"logConfiguration":{},"mountPoints":[],"repositoryCredentials":{},"resourceRequirements":[],"secrets":[],"ulimits":[]
			}]}]}`,
		},
		"empty elements": {
			input: `{"taskProperties":[{"containers":[{},{"dependsOn":[{}],"firelensConfiguration":{}}]}]}`,
		},
		"empty enums omitted": {
			input: `{"taskProperties":[{"containers":[{"firelensConfiguration":{"type":""}}]}]}`,
			want:  `{"taskProperties":[{"containers":[{"firelensConfiguration":{}}]}]}`,
		},
		"unknown enums retained": {
			input: `{"taskProperties":[{"containers":[{"firelensConfiguration":{"type":"FUTURE"}}]}]}`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var input *awstypes.EcsProperties
			if err := tfjson.DecodeFromString(tc.input, &input); err != nil {
				t.Fatal(err)
			}
			got, err := tfbatch.FlattenECSProperties(input)
			if err != nil {
				t.Fatal(err)
			}
			if input == nil {
				if got != "" {
					t.Fatalf("got %q, want empty string", got)
				}
				return
			}
			want := tc.want
			if want == "" {
				want = tc.input
			}
			if !json.Valid([]byte(got)) {
				t.Fatalf("invalid JSON: %s", got)
			}
			if diff := jsoncmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected JSON (+got, -want): %s", diff)
			}
		})
	}
}

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
