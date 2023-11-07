// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAssociationById(ctx context.Context, conn *ssm.SSM, id string) (*ssm.AssociationDescription, error) {
	input := &ssm.DescribeAssociationInput{
		AssociationId: aws.String(id),
	}

	output, err := conn.DescribeAssociationWithContext(ctx, input)
	if tfawserr.ErrCodeContains(err, ssm.ErrCodeAssociationDoesNotExist) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AssociationDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AssociationDescription, nil
}

// FindPatchGroup returns matching SSM Patch Group by Patch Group and BaselineId.
func FindPatchGroup(ctx context.Context, conn *ssm.SSM, patchGroup, baselineId string) (*ssm.PatchGroupPatchBaselineMapping, error) {
	input := &ssm.DescribePatchGroupsInput{}
	var result *ssm.PatchGroupPatchBaselineMapping

	err := conn.DescribePatchGroupsPagesWithContext(ctx, input, func(page *ssm.DescribePatchGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, mapping := range page.Mappings {
			if mapping == nil {
				continue
			}

			if aws.StringValue(mapping.PatchGroup) == patchGroup {
				if mapping.BaselineIdentity != nil && aws.StringValue(mapping.BaselineIdentity.BaselineId) == baselineId {
					result = mapping
					return false
				}
			}
		}

		return !lastPage
	})

	return result, err
}

func FindServiceSettingByID(ctx context.Context, conn *ssm.SSM, id string) (*ssm.ServiceSetting, error) {
	input := &ssm.GetServiceSettingInput{
		SettingId: aws.String(id),
	}

	output, err := conn.GetServiceSettingWithContext(ctx, input)

	if tfawserr.ErrCodeContains(err, ssm.ErrCodeServiceSettingNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServiceSetting == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ServiceSetting, nil
}
