package codestarconnections

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindConnectionByARN(ctx context.Context, conn *codestarconnections.CodeStarConnections, arn string) (*codestarconnections.Connection, error) {
	input := &codestarconnections.GetConnectionInput{
		ConnectionArn: aws.String(arn),
	}

	output, err := conn.GetConnectionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Connection == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Connection, nil
}

func FindHostByARN(ctx context.Context, conn *codestarconnections.CodeStarConnections, arn string) (*codestarconnections.GetHostOutput, error) {
	input := &codestarconnections.GetHostInput{
		HostArn: aws.String(arn),
	}

	output, err := conn.GetHostWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
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
