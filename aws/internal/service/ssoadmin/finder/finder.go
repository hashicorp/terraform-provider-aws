package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
)

func AttachedManagedPolicy(conn *ssoadmin.SSOAdmin, permissionSetArn, instanceArn, managedPolicyArn string) (*ssoadmin.AttachedManagedPolicy, error) {
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
			if aws.StringValue(policy.Arn) == managedPolicyArn {
				attachedPolicy = policy
				return false
			}
		}
		return !lastPage
	})

	return attachedPolicy, err
}

func InlinePolicy(conn *ssoadmin.SSOAdmin, instanceArn, permissionSetArn string) (*string, error) {
	input := &ssoadmin.GetInlinePolicyForPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	output, err := conn.GetInlinePolicyForPermissionSet(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.InlinePolicy, nil
}
