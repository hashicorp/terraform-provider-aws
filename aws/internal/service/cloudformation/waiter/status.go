package waiter

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func ChangeSetStatus(conn *cloudformation.CloudFormation, stackID, changeSetName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(changeSetName),
			StackName:     aws.String(stackID),
		})
		if err != nil {
			log.Printf("[ERROR] Failed to describe CloudFormation change set: %s", err)
			return nil, "", err
		}

		if resp == nil {
			log.Printf("[WARN] Describing CloudFormation change set returned no response")
			return nil, "", nil
		}

		status := aws.StringValue(resp.Status)

		return resp, status, err
	}
}

func StackSetOperationStatus(conn *cloudformation.CloudFormation, stackSetName, operationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.StackSetOperationByStackSetNameAndOperationID(conn, stackSetName, operationID)

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

func StackStatus(conn *cloudformation.CloudFormation, stackName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeStacks(&cloudformation.DescribeStacksInput{
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

func TypeRegistrationProgressStatus(ctx context.Context, conn *cloudformation.CloudFormation, registrationToken string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.TypeRegistrationByToken(ctx, conn, registrationToken)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ProgressStatus), nil
	}
}
