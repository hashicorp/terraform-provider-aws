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

func TestFlattenContainerProperties(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"nil":   {input: "null", want: ""},
		"empty": {input: "{}", want: "{}"},
		"all fields": {
			input: `{
				"command":["sh","-c","echo \"hello\" <&>\n"],
				"enableExecuteCommand":true,
				"environment":[{"name":"A","value":""},{"name":"Z","value":"last"}],
				"ephemeralStorage":{"sizeInGiB":21},
				"executionRoleArn":"execution-role",
				"fargatePlatformConfiguration":{"platformVersion":"LATEST"},
				"image":"image",
				"instanceType":"m5.large",
				"jobRoleArn":"job-role",
				"linuxParameters":{
					"devices":[{"containerPath":"/dev/x","hostPath":"/dev/y","permissions":["READ","WRITE","MKNOD"]}],
					"initProcessEnabled":true,"maxSwap":1024,"sharedMemorySize":64,"swappiness":60,
					"tmpfs":[{"containerPath":"/tmp","mountOptions":["rw","noexec"],"size":128}]
				},
				"logConfiguration":{"logDriver":"awslogs","options":{"MixedCase.Key":"<&>","empty":""},"secretOptions":[{"name":"log-secret","valueFrom":"log-value"}]},
				"memory":512,
				"mountPoints":[{"containerPath":"/data","readOnly":true,"sourceVolume":"data"}],
				"networkConfiguration":{"assignPublicIp":"ENABLED"},
				"privileged":true,"readonlyRootFilesystem":true,
				"repositoryCredentials":{"credentialsParameter":"credentials"},
				"resourceRequirements":[{"type":"VCPU","value":"1"},{"type":"MEMORY","value":"512"}],
				"runtimePlatform":{"cpuArchitecture":"ARM64","operatingSystemFamily":"LINUX"},
				"secrets":[{"name":"secret","valueFrom":"value"}],
				"ulimits":[{"hardLimit":1024,"name":"nofile","softLimit":512}],
				"user":"1000","vcpus":1,
				"volumes":[
					{"name":"data","host":{"sourcePath":"/data"}},
					{"name":"efs","efsVolumeConfiguration":{"authorizationConfig":{"accessPointId":"access-point","iam":"ENABLED"},"fileSystemId":"fs-id","rootDirectory":"/","transitEncryption":"ENABLED","transitEncryptionPort":2049}},
					{"name":"s3","s3filesVolumeConfiguration":{"accessPointArn":"access-point","fileSystemArn":"file-system","rootDirectory":"/","transitEncryptionPort":2049}}
				]
			}`,
		},
		"explicit zero values": {
			input: `{
				"command":[""],"enableExecuteCommand":false,"environment":[{"name":"","value":""}],
				"ephemeralStorage":{"sizeInGiB":0},"executionRoleArn":"",
				"fargatePlatformConfiguration":{"platformVersion":""},"image":"","instanceType":"","jobRoleArn":"",
				"linuxParameters":{"devices":[{"containerPath":"","hostPath":"","permissions":[""]}],"initProcessEnabled":false,"maxSwap":0,"sharedMemorySize":0,"swappiness":0,"tmpfs":[{"containerPath":"","mountOptions":[""],"size":0}]},
				"logConfiguration":{"options":{"":""},"secretOptions":[{"name":"","valueFrom":""}]},
				"memory":0,"mountPoints":[{"containerPath":"","readOnly":false,"sourceVolume":""}],
				"privileged":false,"readonlyRootFilesystem":false,"repositoryCredentials":{"credentialsParameter":""},
				"resourceRequirements":[{"value":""}],"runtimePlatform":{"cpuArchitecture":"","operatingSystemFamily":""},
				"secrets":[{"name":"","valueFrom":""}],"ulimits":[{"hardLimit":0,"name":"","softLimit":0}],"user":"","vcpus":0,
				"volumes":[{"name":"","host":{"sourcePath":""},"efsVolumeConfiguration":{"authorizationConfig":{"accessPointId":""},"fileSystemId":"","rootDirectory":"","transitEncryptionPort":0},"s3filesVolumeConfiguration":{"accessPointArn":"","fileSystemArn":"","rootDirectory":"","transitEncryptionPort":0}}]
			}`,
		},
		"empty collections": {
			input: `{
				"command":[],"environment":[],"ephemeralStorage":{},"fargatePlatformConfiguration":{},
				"linuxParameters":{"devices":[],"tmpfs":[]},"logConfiguration":{"options":{},"secretOptions":[]},
				"mountPoints":[],"networkConfiguration":{},"repositoryCredentials":{},"resourceRequirements":[],
				"runtimePlatform":{},"secrets":[],"ulimits":[],"volumes":[]
			}`,
		},
		"empty elements": {
			input: `{
				"environment":[{}],"linuxParameters":{"devices":[{},{"permissions":[]}],"tmpfs":[{},{"mountOptions":[]}]},
				"logConfiguration":{"secretOptions":[{}]},"mountPoints":[{}],"resourceRequirements":[{}],"secrets":[{}],"ulimits":[{}],
				"volumes":[{},{"host":{},"efsVolumeConfiguration":{"authorizationConfig":{}},"s3filesVolumeConfiguration":{}}]
			}`,
		},
		"empty enums omitted": {
			input: `{"networkConfiguration":{"assignPublicIp":""},"logConfiguration":{"logDriver":""},"resourceRequirements":[{"type":""}],"volumes":[{"efsVolumeConfiguration":{"authorizationConfig":{"iam":""},"transitEncryption":""}}]}`,
			want:  `{"networkConfiguration":{},"logConfiguration":{},"resourceRequirements":[{}],"volumes":[{"efsVolumeConfiguration":{"authorizationConfig":{}}}]}`,
		},
		"unknown enums retained": {
			input: `{"networkConfiguration":{"assignPublicIp":"FUTURE"},"logConfiguration":{"logDriver":"FUTURE"},"resourceRequirements":[{"type":"FUTURE"}],"volumes":[{"efsVolumeConfiguration":{"authorizationConfig":{"iam":"FUTURE"},"transitEncryption":"FUTURE"}}]}`,
		},
		"environment sorted without removing empty values": {
			input: `{"environment":[{"name":"Z","value":"last"},{"name":"A","value":""}]}`,
			want:  `{"environment":[{"name":"A","value":""},{"name":"Z","value":"last"}]}`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var input *awstypes.ContainerProperties
			if err := tfjson.DecodeFromString(tc.input, &input); err != nil {
				t.Fatal(err)
			}
			got, err := tfbatch.FlattenContainerProperties(input)
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

func TestEquivalentContainerPropertiesJSON(t *testing.T) {
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
		"empty ResourceRequirements": {
			apiJSON: `
{
	"command": ["ls", "-la"],
	"environment": [
		{
			"name": "VARNAME",
			"value": "VARVAL"
		}
	],
	"image": "busybox",
	"memory":512,
	"mountPoints": [
		{
			"containerPath": "/tmp",
			"readOnly": false,
			"sourceVolume": "tmp"
		}
	],
	"resourceRequirements": [],
	"ulimits": [
		{
			"hardLimit": 1024,
			"name": "nofile",
			"softLimit": 1024
		}
	],
	"vcpus": 1,
	"volumes": [
		{
			"host": {
				"sourcePath": "/tmp"
			},
			"name": "tmp"
		}
	]
}
`,
			configurationJSON: `
{
	"command": ["ls", "-la"],
	"environment": [
		{
			"name": "VARNAME",
			"value": "VARVAL"
		}
	],
	"image": "busybox",
	"memory":512,
	"mountPoints": [
		{
			"containerPath": "/tmp",
			"readOnly": false,
			"sourceVolume": "tmp"
		}
	],
	"ulimits": [
		{
			"hardLimit": 1024,
			"name": "nofile",
			"softLimit": 1024
		}
	],
	"vcpus": 1,
	"volumes": [
		{
			"host": {
				"sourcePath": "/tmp"
			},
			"name": "tmp"
		}
	]
}
`,
			wantEquivalent: true,
		},
		"reordered Environment": {
			apiJSON: `
{
	"command": ["ls", "-la"],
	"environment": [
		{
			"name": "VARNAME1",
			"value": "VARVAL1"
		},
		{
			"name": "VARNAME2",
			"value": "VARVAL2"
		}
	],
	"image": "busybox",
	"memory":512,
	"mountPoints": [
		{
			"containerPath": "/tmp",
			"readOnly": false,
			"sourceVolume": "tmp"
		}
	],
	"resourceRequirements": [],
	"ulimits": [
		{
			"hardLimit": 1024,
			"name": "nofile",
			"softLimit": 1024
		}
	],
	"vcpus": 1,
	"volumes": [
		{
			"host": {
				"sourcePath": "/tmp"
			},
			"name": "tmp"
		}
	]
}
`,
			configurationJSON: `
{
	"command": ["ls", "-la"],
	"environment": [
		{
			"name": "VARNAME2",
			"value": "VARVAL2"
		},
		{
			"name": "VARNAME1",
			"value": "VARVAL1"
		}
	],
	"image": "busybox",
	"memory":512,
	"mountPoints": [
		{
			"containerPath": "/tmp",
			"readOnly": false,
			"sourceVolume": "tmp"
		}
	],
	"resourceRequirements": [],
	"ulimits": [
		{
			"hardLimit": 1024,
			"name": "nofile",
			"softLimit": 1024
		}
	],
	"vcpus": 1,
	"volumes": [
		{
			"host": {
				"sourcePath": "/tmp"
			},
			"name": "tmp"
		}
	]
}
`,
			wantEquivalent: true,
		},
		"empty environment, mountPoints, ulimits, and volumes": {
			//lintignore:AWSAT005
			apiJSON: `
{
	"image": "example:image",
	"vcpus": 8,
	"memory": 2048,
	"command": ["start.py", "Ref::S3bucket", "Ref::S3key"],
	"jobRoleArn": "arn:aws:iam::123456789012:role/example",
	"volumes": [],
	"environment": [],
	"mountPoints": [],
	"ulimits": [],
	"resourceRequirements": []
}
`,
			//lintignore:AWSAT005
			configurationJSON: `
{
	"command": ["start.py", "Ref::S3bucket", "Ref::S3key"],
	"image": "example:image",
	"memory": 2048,
	"vcpus": 8,
	"jobRoleArn": "arn:aws:iam::123456789012:role/example"
}
`,
			wantEquivalent: true,
		},
		"empty command, logConfiguration.secretOptions, mountPoints, resourceRequirements, secrets, ulimits, volumes": {
			//lintignore:AWSAT003,AWSAT005
			apiJSON: `
{
	"image": "123.dkr.ecr.us-east-1.amazonaws.com/my-app",
	"vcpus": 1,
	"memory": 4096,
	"command": [],
	"jobRoleArn": "arn:aws:iam::123:role/role-test",
	"volumes": [],
	"environment": [{"name":"ENVIRONMENT","value":"test"}],
	"logConfiguration": {
		"logDriver": "awslogs",
		"secretOptions": []
	},
	"mountPoints": [],
	"ulimits": [],
	"resourceRequirements": [],
	"secrets": []
}
`,
			//lintignore:AWSAT003,AWSAT005
			configurationJSON: `
{
    "image": "123.dkr.ecr.us-east-1.amazonaws.com/my-app",
    "memory": 4096,
    "vcpus": 1,
    "jobRoleArn": "arn:aws:iam::123:role/role-test",
    "environment": [
      {
        "name": "ENVIRONMENT",
        "value": "test"
      }
   ],
   "logConfiguration": {
		"logDriver": "awslogs"
	}
}
`,
			wantEquivalent: true,
		},
		"no fargatePlatformConfiguration": {
			//lintignore:AWSAT003,AWSAT005
			apiJSON: `
{
	"image": "123.dkr.ecr.us-east-1.amazonaws.com/my-app",
	"resourceRequirements": [
	  {
		"type": "MEMORY",
		"value": "512"
	  },
	  {
		"type": "VCPU",
		"value": "0.25"
	  }
	],
	"fargatePlatformConfiguration": {
		"platformVersion": "LATEST"
	}
}
`,
			//lintignore:AWSAT003,AWSAT005
			configurationJSON: `
{
	"image": "123.dkr.ecr.us-east-1.amazonaws.com/my-app",
	"resourceRequirements": [
	  {
		  "type": "MEMORY",
		  "value": "512"
	  },
	  {
		"type": "VCPU",
		"value": "0.25"
	  }
	]
}
`,
			wantEquivalent: true,
		},
		"empty linuxParameters.devices, linuxParameters.tmpfs, logConfiguration.options": {
			//lintignore:AWSAT003,AWSAT005
			apiJSON: `
{
	"image": "123.dkr.ecr.us-east-1.amazonaws.com/my-app",
	"vcpus": 1,
	"memory": 4096,
	"jobRoleArn": "arn:aws:iam::123:role/role-test",
	"environment": [{"name":"ENVIRONMENT","value":"test"}],
    "linuxParameters": {
		"devices": [],
		"initProcessEnabled": true,
		"tmpfs": []
	},
	"logConfiguration": {
		"logDriver": "awslogs",
		"options": {}
	}
}
`,
			//lintignore:AWSAT003,AWSAT005
			configurationJSON: `
{
	"image": "123.dkr.ecr.us-east-1.amazonaws.com/my-app",
	"vcpus": 1,
	"memory": 4096,
	"jobRoleArn": "arn:aws:iam::123:role/role-test",
	"environment": [{"name":"ENVIRONMENT","value":"test"}],
    "linuxParameters": {
		"initProcessEnabled": true
	},
	"logConfiguration": {
		"logDriver": "awslogs"
	}
}
`,
			wantEquivalent: true,
		},
		"empty linuxParameters.devices.permissions, linuxParameters.tmpfs.mountOptions": {
			//lintignore:AWSAT003,AWSAT005
			apiJSON: `
{
	"image": "123.dkr.ecr.us-east-1.amazonaws.com/my-app",
	"vcpus": 1,
	"memory": 4096,
	"jobRoleArn": "arn:aws:iam::123:role/role-test",
	"environment": [{"name":"ENVIRONMENT","value":"test"}],
    "linuxParameters": {
		"devices": [{
			"containerPath": "/test",
			"hostPath": "/tmp",
			"permissions": []
		}],
		"initProcessEnabled": true,
		"tmpfs": [{
			"containerPath": "/tmp",
			"mountOptions": [],
			"size": 4096
		}]
	}
}
`,
			//lintignore:AWSAT003,AWSAT005
			configurationJSON: `
{
	"image": "123.dkr.ecr.us-east-1.amazonaws.com/my-app",
	"vcpus": 1,
	"memory": 4096,
	"jobRoleArn": "arn:aws:iam::123:role/role-test",
	"environment": [{"name":"ENVIRONMENT","value":"test"}],
    "linuxParameters": {
		"devices": [{
			"containerPath": "/test",
			"hostPath": "/tmp"
		}],
		"initProcessEnabled": true,
		"tmpfs": [{
			"containerPath": "/tmp",
			"size": 4096
		}]
	}
}
`,
			wantEquivalent: true,
		},
		"empty environment variables": {
			//lintignore:AWSAT005
			apiJSON: `
{
	"image": "example:image",
	"vcpus": 8,
	"memory": 2048,
	"command": ["start.py", "Ref::S3bucket", "Ref::S3key"],
	"environment": [
		{
			"name": "VALUE",
			"value": "test"
		}
	],
	"jobRoleArn": "arn:aws:iam::123456789012:role/example",
	"volumes": [],
	"mountPoints": [],
	"ulimits": [],
	"resourceRequirements": []
}`,
			//lintignore:AWSAT005
			configurationJSON: `
{
	"command": ["start.py", "Ref::S3bucket", "Ref::S3key"],
	"image": "example:image",
	"memory": 2048,
	"vcpus": 8,
	"environment": [
		{
			"name": "EMPTY",
			"value": ""
		},
		{
			"name": "VALUE",
			"value": "test"
		}
	],
	"jobRoleArn": "arn:aws:iam::123456789012:role/example"
}`,
			wantEquivalent: true,
		},
		"empty environment variable": {
			//lintignore:AWSAT005
			apiJSON: `
{
	"image": "example:image",
	"vcpus": 8,
	"memory": 2048,
	"command": ["start.py", "Ref::S3bucket", "Ref::S3key"],
	"environment": [],
	"jobRoleArn": "arn:aws:iam::123456789012:role/example",
	"volumes": [],
	"mountPoints": [],
	"ulimits": [],
	"resourceRequirements": []
}`,
			//lintignore:AWSAT005
			configurationJSON: `
{
	"command": ["start.py", "Ref::S3bucket", "Ref::S3key"],
	"image": "example:image",
	"memory": 2048,
	"vcpus": 8,
	"environment": [
		{
			"name": "EMPTY",
			"value": ""
		}
	],
	"jobRoleArn": "arn:aws:iam::123456789012:role/example"
}`,
			wantEquivalent: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			output, err := tfbatch.EquivalentContainerPropertiesJSON(testCase.configurationJSON, testCase.apiJSON)
			if got, want := err != nil, testCase.wantErr; !cmp.Equal(got, want) {
				t.Errorf("EquivalentContainerPropertiesJSON err %t, want %t", got, want)
			}

			if err == nil {
				if got, want := output, testCase.wantEquivalent; !cmp.Equal(got, want) {
					t.Errorf("EquivalentContainerPropertiesJSON equivalent %t, want %t", got, want)
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
