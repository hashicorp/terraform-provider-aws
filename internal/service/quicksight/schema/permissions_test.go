// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestDiffPermissions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		oldPermissions  []any
		newPermissions  []any
		expectedGrants  []awstypes.ResourcePermission
		expectedRevokes []awstypes.ResourcePermission
	}{
		{
			name:            "no changes;empty",
			oldPermissions:  []any{},
			newPermissions:  []any{},
			expectedGrants:  nil,
			expectedRevokes: nil,
		},
		{
			name: "no changes;same",
			oldPermissions: []any{
				map[string]any{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action1",
						"action2",
					}),
				},
			},
			newPermissions: []any{
				map[string]any{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action1",
						"action2",
					}),
				}},

			expectedGrants:  nil,
			expectedRevokes: nil,
		},
		{
			name:           "grant only",
			oldPermissions: []any{},
			newPermissions: []any{
				map[string]any{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action1",
						"action2",
					}),
				},
			},
			expectedGrants: []awstypes.ResourcePermission{
				{
					Actions:   []string{"action1", "action2"},
					Principal: aws.String("principal1"),
				},
			},
			expectedRevokes: nil,
		},
		{
			name: "revoke only",
			oldPermissions: []any{
				map[string]any{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action1",
						"action2",
					}),
				},
			},
			newPermissions: []any{},
			expectedGrants: nil,
			expectedRevokes: []awstypes.ResourcePermission{
				{
					Actions:   []string{"action1", "action2"},
					Principal: aws.String("principal1"),
				},
			},
		},
		{
			name: "grant new action",
			oldPermissions: []any{
				map[string]any{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action1",
					}),
				},
			},
			newPermissions: []any{
				map[string]any{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action1",
						"action2",
					}),
				},
			},
			expectedGrants: []awstypes.ResourcePermission{
				{
					Actions:   []string{"action1", "action2"},
					Principal: aws.String("principal1"),
				},
			},
			expectedRevokes: nil,
		},
		{
			name: "revoke old action",
			oldPermissions: []any{
				map[string]any{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"oldAction",
						"onlyOldAction",
					}),
				},
			},
			newPermissions: []any{
				map[string]any{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"oldAction",
					}),
				},
			},
			expectedGrants: []awstypes.ResourcePermission{
				{
					Actions:   []string{"oldAction"},
					Principal: aws.String("principal1"),
				},
			},
			expectedRevokes: []awstypes.ResourcePermission{
				{
					Actions:   []string{"onlyOldAction"},
					Principal: aws.String("principal1"),
				},
			},
		},
		{
			name: "multiple permissions",
			oldPermissions: []any{
				map[string]any{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action1",
						"action2",
					}),
				},
				map[string]any{
					names.AttrPrincipal: "principal2",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action1",
						"action3",
						"action4",
					}),
				},
				map[string]any{
					names.AttrPrincipal: "principal3",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action5",
					}),
				},
			},
			newPermissions: []any{
				map[string]any{
					names.AttrPrincipal: "principal1",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action1",
						"action2",
					}),
				},
				map[string]any{
					names.AttrPrincipal: "principal2",
					names.AttrActions: schema.NewSet(schema.HashString, []any{
						"action3",
						"action5",
					}),
				},
			},
			expectedGrants: []awstypes.ResourcePermission{
				{
					Actions:   []string{"action3", "action5"},
					Principal: aws.String("principal2"),
				},
			},
			expectedRevokes: []awstypes.ResourcePermission{
				{
					Actions:   []string{"action1", "action4"},
					Principal: aws.String("principal2"),
				},
				{
					Actions:   []string{"action5"},
					Principal: aws.String("principal3"),
				},
			},
		},
	}

	ignoreExportedOpts := cmpopts.IgnoreUnexported(
		awstypes.ResourcePermission{},
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			toGrant, toRevoke := DiffPermissions(testCase.oldPermissions, testCase.newPermissions)
			if diff := cmp.Diff(toGrant, testCase.expectedGrants, ignoreExportedOpts); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
			if diff := cmp.Diff(toRevoke, testCase.expectedRevokes, ignoreExportedOpts); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
