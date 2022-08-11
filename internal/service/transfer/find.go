package transfer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAccessByServerIDAndExternalID(conn *transfer.Transfer, serverID, externalID string) (*transfer.DescribedAccess, error) {
	input := &transfer.DescribeAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	}

	output, err := conn.DescribeAccess(input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Access == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Access, nil
}

func FindServerByID(conn *transfer.Transfer, id string) (*transfer.DescribedServer, error) {
	input := &transfer.DescribeServerInput{
		ServerId: aws.String(id),
	}

	output, err := conn.DescribeServer(input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Server == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Server, nil
}

func FindUserByServerIDAndUserName(conn *transfer.Transfer, serverID, userName string) (*transfer.DescribedUser, error) {
	input := &transfer.DescribeUserInput{
		ServerId: aws.String(serverID),
		UserName: aws.String(userName),
	}

	output, err := conn.DescribeUser(input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.User == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.User, nil
}

func FindWorkflowByID(conn *transfer.Transfer, id string) (*transfer.DescribedWorkflow, error) {
	input := &transfer.DescribeWorkflowInput{
		WorkflowId: aws.String(id),
	}

	output, err := conn.DescribeWorkflow(input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workflow == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workflow, nil
}
