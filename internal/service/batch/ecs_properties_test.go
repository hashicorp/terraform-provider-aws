// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"testing"

	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
)

func TestEquivalentECSPropertiesJSON(t *testing.T) {
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
		"reordered containers": {
			ApiJson: `
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
			ConfigurationJson: `
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
			ExpectEquivalent: true,
		},
		"reordered environment": {
			ApiJson: `
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
			ConfigurationJson: `
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
			ExpectEquivalent: true,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tfbatch.EquivalentECSPropertiesJSON(testCase.ConfigurationJson, testCase.ApiJson)

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
