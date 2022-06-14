package datasync

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAgentByARN(conn *datasync.DataSync, arn string) (*datasync.DescribeAgentOutput, error) {
	input := &datasync.DescribeAgentInput{
		AgentArn: aws.String(arn),
	}

	output, err := conn.DescribeAgent(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "does not exist") {
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

func FindTaskByARN(conn *datasync.DataSync, arn string) (*datasync.DescribeTaskOutput, error) {
	input := &datasync.DescribeTaskInput{
		TaskArn: aws.String(arn),
	}

	output, err := conn.DescribeTask(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
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

func FindLocationHDFSByARN(conn *datasync.DataSync, arn string) (*datasync.DescribeLocationHdfsOutput, error) {
	input := &datasync.DescribeLocationHdfsInput{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationHdfs(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
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

func FindFSxLustreLocationByARN(conn *datasync.DataSync, arn string) (*datasync.DescribeLocationFsxLustreOutput, error) {
	input := &datasync.DescribeLocationFsxLustreInput{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationFsxLustre(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
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

func FindFSxOpenZFSLocationByARN(conn *datasync.DataSync, arn string) (*datasync.DescribeLocationFsxOpenZfsOutput, error) {
	input := &datasync.DescribeLocationFsxOpenZfsInput{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationFsxOpenZfs(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
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
