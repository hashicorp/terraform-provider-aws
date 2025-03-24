// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
)

func TestEquivalentEKSPropertiesJSON(t *testing.T) {
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
		    "podProperties": {
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
		  }
					`,
			configurationJSON: `
		{
		    "podProperties": {
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
		  }
					`,
			wantEquivalent: true,
		},
		"reordered environment": {
			apiJSON: `
		{
		  "podProperties": {
		      "containers": [
		        {
		          "name": "container1",
		          "image": "my_ecr_image1",
		          "env": [
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
		          "env": []
		        }
		      ]
		  }
		}
					`,
			configurationJSON: `
		{
		  "podProperties": {
		      "containers": [
		        {
		          "name": "container1",
		          "image": "my_ecr_image1",
		          "env": [
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
		}
					`,
			wantEquivalent: true,
		},
		"full": {
			apiJSON: `
    {
      "podProperties": {
        "containers": [
          {
            "command": ["sleep", "60"],
            "env": [
              {
                "name": "test",
                "value": "Environment Variable"
              }
            ],
            "image": "public.ecr.aws/amazonlinux/amazonlinux:1",
            "volumeMounts": [],
            "name": "container_a",
            "resources": {
              "requests": {
                "cpu": "1.0",
                "memory": "2048"
              }
            }
          },
          {
            "command": ["sleep", "360"],
            "env": [],
            "image": "public.ecr.aws/amazonlinux/amazonlinux:1",
            "volumeMounts": [],
            "name": "container_b",
            "resources": {
              "requests": {
                "cpu": "1.0",
                "memory": "2048"
              }
            }
          }
        ],
        "volumes": []
      }
    }
		      `,
			configurationJSON: `
		{
      "podProperties": {
        "containers": [
          {
            "command": ["sleep", "60"],
            "env": [
              {
                "name": "test",
                "value": "Environment Variable"
              }
            ],
            "image": "public.ecr.aws/amazonlinux/amazonlinux:1",
            "name": "container_a",
            "resources": {
              "requests": {
                "cpu": "1.0",
                "memory": "2048"
              }
            }
          },
          {
            "command": ["sleep", "360"],
            "image": "public.ecr.aws/amazonlinux/amazonlinux:1",
            "name": "container_b",
            "resources": {
              "requests": {
                "cpu": "1.0",
                "memory": "2048"
              }
            }
          }
        ]
      }
    }
		      `,
			wantEquivalent: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			output, err := tfbatch.EquivalentEKSPropertiesJSON(testCase.configurationJSON, testCase.apiJSON)
			if got, want := err != nil, testCase.wantErr; !cmp.Equal(got, want) {
				t.Errorf("EquivalentEKSPropertiesJSON err %t, want %t", got, want)
			}

			if err == nil {
				if got, want := output, testCase.wantEquivalent; !cmp.Equal(got, want) {
					t.Errorf("EquivalentEKSPropertiesJSON equivalent %t, want %t", got, want)
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
