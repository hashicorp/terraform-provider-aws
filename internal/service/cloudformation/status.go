package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusChangeSet(ctx context.Context, conn *cloudformation.CloudFormation, stackID, changeSetName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindChangeSetByStackIDAndChangeSetName(ctx, conn, stackID, changeSetName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusStackSetOperation(ctx context.Context, conn *cloudformation.CloudFormation, stackSetName, operationID, callAs string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStackSetOperationByStackSetNameAndOperationID(ctx, conn, stackSetName, operationID, callAs)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

const (
	stackStatusError    = "Error"
	stackStatusNotFound = "NotFound"
)

func StatusStack(ctx context.Context, conn *cloudformation.CloudFormation, stackName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeStacksWithContext(ctx, &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			return nil, stackStatusError, err
		}

		if resp.Stacks == nil || len(resp.Stacks) == 0 {
			return nil, stackStatusNotFound, nil
		}

		return resp.Stacks[0], aws.StringValue(resp.Stacks[0].StackStatus), err
	}
}

func StatusTypeRegistrationProgress(ctx context.Context, conn *cloudformation.CloudFormation, registrationToken string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTypeRegistrationByToken(ctx, conn, registrationToken)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ProgressStatus), nil
	}
}
