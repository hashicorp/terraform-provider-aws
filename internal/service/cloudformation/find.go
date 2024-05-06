// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindChangeSetByStackIDAndChangeSetName(ctx context.Context, conn *cloudformation.Client, stackID, changeSetName string) (*cloudformation.DescribeChangeSetOutput, error) {
	input := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeSetName),
		StackName:     aws.String(stackID),
	}

	output, err := conn.DescribeChangeSet(ctx, input)

	if errs.IsA[*awstypes.ChangeSetNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindStackInstanceSummariesByOrgIDs(ctx context.Context, conn *cloudformation.Client, stackSetName, region, callAs string, orgIDs []string) ([]awstypes.StackInstanceSummary, error) {
	input := &cloudformation.ListStackInstancesInput{
		StackInstanceRegion: aws.String(region),
		StackSetName:        aws.String(stackSetName),
	}

	if callAs != "" {
		input.CallAs = awstypes.CallAs(callAs)
	}

	var result []awstypes.StackInstanceSummary

	pages := cloudformation.NewListStackInstancesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.StackSetNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, s := range page.Summaries {
			for _, orgID := range orgIDs {
				if aws.ToString(s.OrganizationalUnitId) == orgID {
					result = append(result, s)
				}
			}
		}
	}

	return result, nil
}

func FindStackInstanceByName(ctx context.Context, conn *cloudformation.Client, stackSetName, accountID, region, callAs string) (*awstypes.StackInstance, error) {
	input := &cloudformation.DescribeStackInstanceInput{
		StackInstanceAccount: aws.String(accountID),
		StackInstanceRegion:  aws.String(region),
		StackSetName:         aws.String(stackSetName),
	}

	if callAs != "" {
		input.CallAs = awstypes.CallAs(callAs)
	}

	output, err := conn.DescribeStackInstance(ctx, input)

	if errs.IsA[*awstypes.StackInstanceNotFoundException](err) || errs.IsA[*awstypes.StackSetNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.StackInstance == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.StackInstance, nil
}

func FindStackSetByName(ctx context.Context, conn *cloudformation.Client, name, callAs string) (*awstypes.StackSet, error) {
	input := &cloudformation.DescribeStackSetInput{
		StackSetName: aws.String(name),
	}

	if callAs != "" {
		input.CallAs = awstypes.CallAs(callAs)
	}

	output, err := conn.DescribeStackSet(ctx, input)

	if errs.IsA[*awstypes.StackSetNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if callAs == string(awstypes.CallAsDelegatedAdmin) && tfawserr.ErrMessageContains(err, errCodeValidationError, "Failed to check account is Delegated Administrator") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.StackSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.StackSet, nil
}

func FindStackSetOperationByStackSetNameAndOperationID(ctx context.Context, conn *cloudformation.Client, stackSetName, operationID, callAs string) (*awstypes.StackSetOperation, error) {
	input := &cloudformation.DescribeStackSetOperationInput{
		OperationId:  aws.String(operationID),
		StackSetName: aws.String(stackSetName),
	}

	if callAs != "" {
		input.CallAs = awstypes.CallAs(callAs)
	}

	output, err := conn.DescribeStackSetOperation(ctx, input)

	if errs.IsA[*awstypes.OperationNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.StackSetOperation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.StackSetOperation, nil
}
