package quicksight

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DiffPermissions(o, n []interface{}) ([]*quicksight.ResourcePermission, []*quicksight.ResourcePermission) {
	old := expandDataSourcePermissions(o)
	new := expandDataSourcePermissions(n)

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
			toAdd := newActions.Difference(oldActions)

			if toRemove.Len() > 0 {
				toRevoke = append(toRevoke, &quicksight.ResourcePermission{
					Actions:   flex.ExpandStringSet(toRemove),
					Principal: np.Principal,
				})
			}

			if toAdd.Len() > 0 {
				toGrant = append(toGrant, &quicksight.ResourcePermission{
					Actions:   flex.ExpandStringSet(toAdd),
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
