// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func PermissionsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MinItems: 1,
		MaxItems: 64,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrActions: {
					Type:     schema.TypeSet,
					Required: true,
					MinItems: 1,
					MaxItems: 20,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				names.AttrPrincipal: stringLenBetweenSchema(attrRequired, 1, 256),
			},
		},
	}
}

func PermissionsDataSourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrActions: {
					Type:     schema.TypeSet,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				names.AttrPrincipal: {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func ExpandResourcePermissions(tfList []any) []awstypes.ResourcePermission {
	apiObjects := make([]awstypes.ResourcePermission, len(tfList))

	for i, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := awstypes.ResourcePermission{
			Actions:   flex.ExpandStringValueSet(tfMap[names.AttrActions].(*schema.Set)),
			Principal: aws.String(tfMap[names.AttrPrincipal].(string)),
		}

		apiObjects[i] = apiObject
	}

	return apiObjects
}

func FlattenPermissions(apiObjects []awstypes.ResourcePermission) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		if apiObject.Actions != nil {
			tfMap[names.AttrActions] = apiObject.Actions
		}

		if apiObject.Principal != nil {
			tfMap[names.AttrPrincipal] = aws.ToString(apiObject.Principal)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func DiffPermissions(o, n []any) ([]awstypes.ResourcePermission, []awstypes.ResourcePermission) {
	old := ExpandResourcePermissions(o)
	new := ExpandResourcePermissions(n)

	var toGrant, toRevoke []awstypes.ResourcePermission

	for _, op := range old {
		found := false

		for _, np := range new {
			if aws.ToString(np.Principal) != aws.ToString(op.Principal) {
				continue
			}

			found = true
			newActions := flex.FlattenStringValueSet(np.Actions)
			oldActions := flex.FlattenStringValueSet(op.Actions)

			if newActions.Equal(oldActions) {
				break
			}

			toRemove := oldActions.Difference(newActions)

			if toRemove.Len() > 0 {
				toRevoke = append(toRevoke, awstypes.ResourcePermission{
					Actions:   flex.ExpandStringValueSet(toRemove),
					Principal: np.Principal,
				})
			}

			if newActions.Len() > 0 {
				toGrant = append(toGrant, awstypes.ResourcePermission{
					Actions:   flex.ExpandStringValueSet(newActions),
					Principal: np.Principal,
				})
			}
		}

		if !found {
			toRevoke = append(toRevoke, op)
		}
	}

	for _, np := range new {
		found := false

		for _, op := range old {
			if aws.ToString(np.Principal) == aws.ToString(op.Principal) {
				found = true
				break
			}
		}

		if !found {
			toGrant = append(toGrant, np)
		}
	}

	return toGrant, toRevoke
}
