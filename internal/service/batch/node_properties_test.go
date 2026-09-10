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

func TestFlattenNodeProperties(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"nil":   {input: "null", want: ""},
		"empty": {input: "{}", want: "{}"},
		"all fields": {
			input: `{
				"mainNode":1,"numNodes":3,
				"nodeRangeProperties":[
					{
						"consumableResourceProperties":{"consumableResourceList":[{"consumableResource":"resource","quantity":4294967296}]},
						"container":{"image":"image","environment":[{"name":"Z","value":"last"},{"name":"A","value":""}]},
						"instanceTypes":["m5.large","m5.xlarge"],"targetNodes":"0:1"
					},
					{
						"ecsProperties":{"taskProperties":[{"containers":[{"name":"Z","image":"image"},{"name":"A","image":"other"}]}]},
						"targetNodes":"2"
					},
					{
						"eksProperties":{"podProperties":{
							"containers":[{
								"args":["hello \"world\" <&>\n\u2603"],"command":["sh","-c"],
								"env":[{"name":"Z","value":"last"},{"name":"A","value":""}],
								"image":"image","imagePullPolicy":"Always","name":"main",
								"resources":{"limits":{"cpu":"2","memory":"2Gi","example.com/GPU":"1"},"requests":{"cpu":"1","memory":"1Gi"}},
								"securityContext":{"allowPrivilegeEscalation":true,"privileged":true,"readOnlyRootFilesystem":true,"runAsGroup":4294967296,"runAsNonRoot":true,"runAsUser":4294967297},
								"volumeMounts":[{"mountPath":"/data","name":"data","readOnly":true,"subPath":"subdir"}]
							},{"name":"sidecar","image":"sidecar-image"}],
							"dnsPolicy":"ClusterFirst","hostNetwork":true,
							"imagePullSecrets":[{"name":"registry-secret"}],
							"initContainers":[{"name":"init","image":"init-image","command":["setup"],"args":["--init"]}],
							"metadata":{"annotations":{"MixedCase.Key":"<&>","empty":""},"labels":{"app":"batch"},"namespace":"jobs"},
							"serviceAccountName":"batch","shareProcessNamespace":true,
							"volumes":[
								{"name":"scratch","emptyDir":{"medium":"Memory","sizeLimit":"1Gi"}},
								{"name":"data","hostPath":{"path":"/data"}},
								{"name":"persistent","persistentVolumeClaim":{"claimName":"claim","readOnly":true}},
								{"name":"secret","secret":{"optional":true,"secretName":"secret"}}
							]
						}},
						"targetNodes":"0:"
					}
				]
			}`,
		},
		"explicit zero values": {
			input: `{
				"mainNode":0,"numNodes":0,
				"nodeRangeProperties":[{
					"consumableResourceProperties":{"consumableResourceList":[{"consumableResource":"","quantity":0}]},
					"container":{"image":"","memory":0,"privileged":false},
					"ecsProperties":{"taskProperties":[{"containers":[{"essential":false}]}]},
					"eksProperties":{"podProperties":{
						"containers":[{
							"args":[""],"command":[""],"env":[{"name":"","value":""}],"image":"","imagePullPolicy":"","name":"",
							"resources":{"limits":{"":""},"requests":{"":""}},
							"securityContext":{"allowPrivilegeEscalation":false,"privileged":false,"readOnlyRootFilesystem":false,"runAsGroup":0,"runAsNonRoot":false,"runAsUser":0},
							"volumeMounts":[{"mountPath":"","name":"","readOnly":false,"subPath":""}]
						}],
						"dnsPolicy":"","hostNetwork":false,"imagePullSecrets":[{"name":""}],"initContainers":[{"image":""}],
						"metadata":{"annotations":{"":""},"labels":{"":""},"namespace":""},"serviceAccountName":"","shareProcessNamespace":false,
						"volumes":[{"name":"","emptyDir":{"medium":"","sizeLimit":""},"hostPath":{"path":""},"persistentVolumeClaim":{"claimName":"","readOnly":false},"secret":{"optional":false,"secretName":""}}]
					}},
					"instanceTypes":[""],"targetNodes":""
				}]
			}`,
		},
		"empty node ranges": {input: `{"nodeRangeProperties":[]}`},
		"empty nested objects": {
			input: `{"nodeRangeProperties":[{},{"consumableResourceProperties":{},"container":{},"ecsProperties":{},"eksProperties":{}},{"eksProperties":{"podProperties":{}}}]}`,
		},
		"empty collections": {
			input: `{"nodeRangeProperties":[{
				"consumableResourceProperties":{"consumableResourceList":[]},"instanceTypes":[],
				"container":{"command":[]},"ecsProperties":{"taskProperties":[]},
				"eksProperties":{"podProperties":{"containers":[],"initContainers":[],"imagePullSecrets":[],"metadata":{"annotations":{},"labels":{}},"volumes":[]}}
			},{
				"eksProperties":{"podProperties":{"containers":[{"args":[],"command":[],"env":[],"resources":{"limits":{},"requests":{}},"volumeMounts":[]}]}}
			}]}`,
		},
		"empty elements": {
			input: `{"nodeRangeProperties":[{
				"consumableResourceProperties":{"consumableResourceList":[{}]},
				"eksProperties":{"podProperties":{
					"containers":[{},{"env":[{}],"resources":{},"securityContext":{},"volumeMounts":[{}]}],
					"initContainers":[{}],"imagePullSecrets":[{}],"metadata":{},
					"volumes":[{},{"emptyDir":{},"hostPath":{},"persistentVolumeClaim":{},"secret":{}}]
				}}
			}]}`,
		},
		"nil fields omitted": {
			input: `{"mainNode":null,"numNodes":null,"nodeRangeProperties":[{"consumableResourceProperties":null,"container":null,"ecsProperties":null,"eksProperties":{"podProperties":{"containers":null,"metadata":null,"volumes":null}},"instanceTypes":null,"targetNodes":null}]}`,
			want:  `{"nodeRangeProperties":[{"eksProperties":{"podProperties":{}}}]}`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var input *awstypes.NodeProperties
			if err := tfjson.DecodeFromString(tc.input, &input); err != nil {
				t.Fatal(err)
			}
			got, err := tfbatch.FlattenNodeProperties(input)
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
