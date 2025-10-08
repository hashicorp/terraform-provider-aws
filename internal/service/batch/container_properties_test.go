// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
)

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
