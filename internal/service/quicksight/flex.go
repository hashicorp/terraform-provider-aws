// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandResourcePermissions(tfList []interface{}) []types.ResourcePermission {
	permissions := make([]types.ResourcePermission, len(tfList))

	for i, tfListRaw := range tfList {
		tfMap := tfListRaw.(map[string]interface{})

		permission := &types.ResourcePermission{
			Actions:   flex.ExpandStringValueSet(tfMap["actions"].(*schema.Set)),
			Principal: aws.String(tfMap["principal"].(string)),
		}

		permissions[i] = *permission
	}
	return permissions
}

func DiffPermissions(o, n []interface{}) ([]types.ResourcePermission, []types.ResourcePermission) {
	old := expandResourcePermissions(o)
	new := expandResourcePermissions(n)

	var toGrant, toRevoke []types.ResourcePermission

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
				toRevoke = append(toRevoke, types.ResourcePermission{
					Actions:   flex.ExpandStringValueSet(toRemove),
					Principal: np.Principal,
				})
			}

			if newActions.Len() > 0 {
				toGrant = append(toGrant, types.ResourcePermission{
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

func flattenPermissions(perms []types.ResourcePermission) []interface{} {
	if len(perms) == 0 {
		return []interface{}{}
	}

	values := make([]interface{}, 0)

	for _, p := range perms {
		if p.Actions == nil && p.Principal == nil {
			continue
		}

		perm := make(map[string]interface{})

		if p.Principal != nil {
			perm["principal"] = p.Principal
		}

		if p.Actions != nil {
			perm["actions"] = p.Actions
		}

		values = append(values, perm)
	}

	return values
}
