package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
)

// ManagedPolicy returns the managed policy attached to a permission set within a specified SSO instance.
// Returns an error if no managed policy is found.
func ManagedPolicy(conn *ssoadmin.SSOAdmin, managedPolicyArn, permissionSetArn, instanceArn string) (*ssoadmin.AttachedManagedPolicy, error) {
	input := &ssoadmin.ListManagedPoliciesInPermissionSetInput{
		PermissionSetArn: aws.String(permissionSetArn),
		InstanceArn:      aws.String(instanceArn),
	}

	var attachedPolicy *ssoadmin.AttachedManagedPolicy
	err := conn.ListManagedPoliciesInPermissionSetPages(input, func(page *ssoadmin.ListManagedPoliciesInPermissionSetOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, policy := range page.AttachedManagedPolicies {
			if policy == nil {
				continue
			}

			if aws.StringValue(policy.Arn) == managedPolicyArn {
				attachedPolicy = policy
				return false
			}
		}
		return !lastPage
	})

	return attachedPolicy, err
}
