// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandResourcePermissions(tfList []interface{}) []awstypes.ResourcePermission {
	apiObjects := make([]awstypes.ResourcePermission, len(tfList))

	for i, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := awstypes.ResourcePermission{
			Actions:   flex.ExpandStringValueSet(tfMap[names.AttrActions].(*schema.Set)),
			Principal: aws.String(tfMap[names.AttrPrincipal].(string)),
		}

		apiObjects[i] = apiObject
	}

	return apiObjects
}

func flattenPermissions(apiObjects []awstypes.ResourcePermission) []interface{} {
	if len(apiObjects) == 0 {
		return []interface{}{}
	}

	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]interface{})

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

func diffPermissions(o, n []interface{}) ([]awstypes.ResourcePermission, []awstypes.ResourcePermission) {
	old := expandResourcePermissions(o)
	new := expandResourcePermissions(n)

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
