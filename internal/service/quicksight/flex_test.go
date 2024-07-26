// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestDataSourcePermissionsDiff(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		oldPermissions  []interface{}
		newPermissions  []interface{}
		expectedGrants  []*quicksight.ResourcePermission
		expectedRevokes []*quicksight.ResourcePermission
	}{
		{
			name:            "no changes;empty",
			oldPermissions:  []interface{}{},
			newPermissions:  []interface{}{},
			expectedGrants:  nil,
			expectedRevokes: nil,
		},
		{
			name: "no changes;same",
			oldPermissions: []interface{}{
				map[string]interface{}{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				}},

			expectedGrants:  nil,
			expectedRevokes: nil,
		},
		{
			name:           "grant only",
			oldPermissions: []interface{}{},
			newPermissions: []interface{}{
				map[string]interface{}{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			expectedGrants: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action1", "action2"}),
					Principal: aws.String("principal1"),
				},
			},
			expectedRevokes: nil,
		},
		{
			name: "revoke only",
			oldPermissions: []interface{}{
				map[string]interface{}{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			newPermissions: []interface{}{},
			expectedGrants: nil,
			expectedRevokes: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action1", "action2"}),
					Principal: aws.String("principal1"),
				},
			},
		},
		{
			name: "grant new action",
			oldPermissions: []interface{}{
				map[string]interface{}{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action1",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			expectedGrants: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action1", "action2"}),
					Principal: aws.String("principal1"),
				},
			},
			expectedRevokes: nil,
		},
		{
			name: "revoke old action",
			oldPermissions: []interface{}{
				map[string]interface{}{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"oldAction",
						"onlyOldAction",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"oldAction",
					}),
				},
			},
			expectedGrants: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"oldAction"}),
					Principal: aws.String("principal1"),
				},
			},
			expectedRevokes: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"onlyOldAction"}),
					Principal: aws.String("principal1"),
				},
			},
		},
		{
			name: "multiple permissions",
			oldPermissions: []interface{}{
				map[string]interface{}{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
				map[string]interface{}{
					names.AttrPrincipal: "principal2",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action3",
						"action4",
					}),
				},
				map[string]interface{}{
					names.AttrPrincipal: "principal3",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action5",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
				map[string]interface{}{
					names.AttrPrincipal: "principal2",
					names.AttrActions: schema.NewSet(schema.HashString, []interface{}{
						"action3",
						"action5",
					}),
				},
			},
			expectedGrants: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action3", "action5"}),
					Principal: aws.String("principal2"),
				},
			},
			expectedRevokes: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action1", "action4"}),
					Principal: aws.String("principal2"),
				},
				{
					Actions:   aws.StringSlice([]string{"action5"}),
					Principal: aws.String("principal3"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			toGrant, toRevoke := tfquicksight.DiffPermissions(testCase.oldPermissions, testCase.newPermissions)
			if !reflect.DeepEqual(toGrant, testCase.expectedGrants) {
				t.Fatalf("Expected: %v, got: %v", testCase.expectedGrants, toGrant)
			}

			if !reflect.DeepEqual(toRevoke, testCase.expectedRevokes) {
				t.Fatalf("Expected: %v, got: %v", testCase.expectedRevokes, toRevoke)
			}
		})
	}
}
