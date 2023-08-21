// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// FindManagedPolicy returns the managed policy attached to a permission set within a specified SSO instance.
// Returns an error if no managed policy is found.
func FindManagedPolicy(ctx context.Context, conn *ssoadmin.SSOAdmin, managedPolicyArn, permissionSetArn, instanceArn string) (*ssoadmin.AttachedManagedPolicy, error) {
	input := &ssoadmin.ListManagedPoliciesInPermissionSetInput{
		PermissionSetArn: aws.String(permissionSetArn),
		InstanceArn:      aws.String(instanceArn),
	}

	var attachedPolicy *ssoadmin.AttachedManagedPolicy
	err := conn.ListManagedPoliciesInPermissionSetPagesWithContext(ctx, input, func(page *ssoadmin.ListManagedPoliciesInPermissionSetOutput, lastPage bool) bool {
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

// FindPermissionsBoundary returns the permissions boundary attached to a permission set within a specified SSO instance.
// Returns an error if no permissions boundary is found.
func FindPermissionsBoundary(ctx context.Context, conn *ssoadmin.SSOAdmin, permissionSetArn, instanceArn string) (*ssoadmin.PermissionsBoundary, error) {
	input := &ssoadmin.GetPermissionsBoundaryForPermissionSetInput{
		PermissionSetArn: aws.String(permissionSetArn),
		InstanceArn:      aws.String(instanceArn),
	}

	output, err := conn.GetPermissionsBoundaryForPermissionSetWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output.PermissionsBoundary, nil
}
