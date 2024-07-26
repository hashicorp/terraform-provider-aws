// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandResourcePermissions(tfList []interface{}) []*quicksight.ResourcePermission {
	permissions := make([]*quicksight.ResourcePermission, len(tfList))

	for i, tfListRaw := range tfList {
		tfMap := tfListRaw.(map[string]interface{})

		permission := &quicksight.ResourcePermission{
			Actions:   flex.ExpandStringSet(tfMap[names.AttrActions].(*schema.Set)),
			Principal: aws.String(tfMap[names.AttrPrincipal].(string)),
		}

		permissions[i] = permission
	}
	return permissions
}

func DiffPermissions(o, n []interface{}) ([]*quicksight.ResourcePermission, []*quicksight.ResourcePermission) {
	old := expandResourcePermissions(o)
	new := expandResourcePermissions(n)

	var toGrant, toRevoke []*quicksight.ResourcePermission

	for _, op := range old {
		found := false

		for _, np := range new {
			if aws.StringValue(np.Principal) != aws.StringValue(op.Principal) {
				continue
			}

			found = true
			newActions := flex.FlattenStringSet(np.Actions)
			oldActions := flex.FlattenStringSet(op.Actions)

			if newActions.Equal(oldActions) {
				break
			}

			toRemove := oldActions.Difference(newActions)

			if toRemove.Len() > 0 {
				toRevoke = append(toRevoke, &quicksight.ResourcePermission{
					Actions:   flex.ExpandStringSet(toRemove),
					Principal: np.Principal,
				})
			}

			if newActions.Len() > 0 {
				toGrant = append(toGrant, &quicksight.ResourcePermission{
					Actions:   flex.ExpandStringSet(newActions),
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
			if aws.StringValue(np.Principal) == aws.StringValue(op.Principal) {
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

func flattenPermissions(perms []*quicksight.ResourcePermission) []interface{} {
	if len(perms) == 0 {
		return []interface{}{}
	}

	values := make([]interface{}, 0)

	for _, p := range perms {
		if p == nil {
			continue
		}

		perm := make(map[string]interface{})

		if p.Principal != nil {
			perm[names.AttrPrincipal] = aws.StringValue(p.Principal)
		}

		if p.Actions != nil {
			perm[names.AttrActions] = flex.FlattenStringList(p.Actions)
		}

		values = append(values, perm)
	}

	return values
}
