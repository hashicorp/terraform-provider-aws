// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func FindPolicyAttachmentByTwoPartKey(ctx context.Context, conn *organizations.Organizations, targetID, policyID string) (*organizations.PolicyTargetSummary, error) {
	input := &organizations.ListTargetsForPolicyInput{
		PolicyId: aws.String(policyID),
	}
	var output *organizations.PolicyTargetSummary

	err := conn.ListTargetsForPolicyPagesWithContext(ctx, input, func(page *organizations.ListTargetsForPolicyOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Targets {
			if aws.StringValue(v.TargetId) == targetID {
				output = v
				return true
			}
		}
		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeTargetNotFoundException, organizations.ErrCodePolicyNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}
