package equivalency

import (
	"testing"
)

func TestEquivalentBatchContainerPropertiesJSON(t *testing.T) {
	testCases := []struct {
		Name              string
		ApiJson           string
		ConfigurationJson string
		ExpectEquivalent  bool
		ExpectError       bool
	}{
		{
			Name:              "empty",
			ApiJson:           ``,
			ConfigurationJson: ``,
			ExpectEquivalent:  true,
		},
		{
			Name: "empty ResourceRequirements",
			ApiJson: `
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
			ConfigurationJson: `
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
			ExpectEquivalent: true,
		},
		{
			Name: "reordered Environment",
			ApiJson: `
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
			ConfigurationJson: `
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
			ExpectEquivalent: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got, err := EquivalentBatchContainerPropertiesJSON(testCase.ConfigurationJson, testCase.ApiJson)

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
