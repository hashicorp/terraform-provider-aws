package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindChangeSetByStackIDAndChangeSetName(conn *cloudformation.CloudFormation, stackID, changeSetName string) (*cloudformation.DescribeChangeSetOutput, error) {
	input := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeSetName),
		StackName:     aws.String(stackID),
	}

	output, err := conn.DescribeChangeSet(input)

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeChangeSetNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindStackByID(conn *cloudformation.CloudFormation, id string) (*cloudformation.Stack, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(id),
	}

	output, err := conn.DescribeStacks(input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "does not exist") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Stacks) == 0 || output.Stacks[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Stacks); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	stack := output.Stacks[0]

	if status := aws.StringValue(stack.StackStatus); status == cloudformation.StackStatusDeleteComplete {
		return nil, &resource.NotFoundError{
			LastRequest: input,
			Message:     status,
		}
	}

	return stack, nil
}

func FindStackInstanceAccountIdByOrgIDs(conn *cloudformation.CloudFormation, stackSetName, region, callAs string, orgIDs []string) (string, error) {
	input := &cloudformation.ListStackInstancesInput{
		StackInstanceRegion: aws.String(region),
		StackSetName:        aws.String(stackSetName),
	}

	if callAs != "" {
		input.CallAs = aws.String(callAs)
	}

	var result string

	err := conn.ListStackInstancesPages(input, func(page *cloudformation.ListStackInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, s := range page.Summaries {
			if s == nil {
				continue
			}

			for _, orgID := range orgIDs {
				if aws.StringValue(s.OrganizationalUnitId) == orgID {
					result = aws.StringValue(s.Account)
					return false
				}
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeStackSetNotFoundException) {
		return "", &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", err
	}

	return result, nil
}

func FindStackInstanceByName(conn *cloudformation.CloudFormation, stackSetName, accountID, region, callAs string) (*cloudformation.StackInstance, error) {
	input := &cloudformation.DescribeStackInstanceInput{
		StackInstanceAccount: aws.String(accountID),
		StackInstanceRegion:  aws.String(region),
		StackSetName:         aws.String(stackSetName),
	}

	if callAs != "" {
		input.CallAs = aws.String(callAs)
	}

	output, err := conn.DescribeStackInstance(input)

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeStackInstanceNotFoundException) || tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeStackSetNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindStackSetByName(conn *cloudformation.CloudFormation, name, callAs string) (*cloudformation.StackSet, error) {
	input := &cloudformation.DescribeStackSetInput{
		StackSetName: aws.String(name),
	}

	if callAs != "" {
		input.CallAs = aws.String(callAs)
	}

	output, err := conn.DescribeStackSet(input)

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeStackSetNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindStackSetOperationByStackSetNameAndOperationID(conn *cloudformation.CloudFormation, stackSetName, operationID, callAs string) (*cloudformation.StackSetOperation, error) {
	input := &cloudformation.DescribeStackSetOperationInput{
		OperationId:  aws.String(operationID),
		StackSetName: aws.String(stackSetName),
	}

	if callAs != "" {
		input.CallAs = aws.String(callAs)
	}

	output, err := conn.DescribeStackSetOperation(input)

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeOperationNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindTypeByARN(ctx context.Context, conn *cloudformation.CloudFormation, arn string) (*cloudformation.DescribeTypeOutput, error) {
	input := &cloudformation.DescribeTypeInput{
		Arn: aws.String(arn),
	}

	return FindType(ctx, conn, input)
}

func FindTypeByName(ctx context.Context, conn *cloudformation.CloudFormation, name string) (*cloudformation.DescribeTypeOutput, error) {
	input := &cloudformation.DescribeTypeInput{
		Type:     aws.String(cloudformation.RegistryTypeResource),
		TypeName: aws.String(name),
	}

	return FindType(ctx, conn, input)
}

func FindType(ctx context.Context, conn *cloudformation.CloudFormation, input *cloudformation.DescribeTypeInput) (*cloudformation.DescribeTypeOutput, error) {
	output, err := conn.DescribeTypeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeTypeNotFoundException) {
		return nil, &resource.NotFoundError{
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

	if status := aws.StringValue(output.DeprecatedStatus); status == cloudformation.DeprecatedStatusDeprecated {
		return nil, &resource.NotFoundError{
			LastRequest: input,
			Message:     status,
		}
	}

	return output, nil
}

func FindTypeRegistrationByToken(ctx context.Context, conn *cloudformation.CloudFormation, registrationToken string) (*cloudformation.DescribeTypeRegistrationOutput, error) {
	input := &cloudformation.DescribeTypeRegistrationInput{
		RegistrationToken: aws.String(registrationToken),
	}

	output, err := conn.DescribeTypeRegistrationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, cloudformation.ErrCodeCFNRegistryException, "No registration token matches") {
		return nil, &resource.NotFoundError{
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
