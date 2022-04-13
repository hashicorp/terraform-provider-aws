package cloudcontrol

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudcontrolapi"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindProgressEventByRequestToken(ctx context.Context, conn *cloudcontrolapi.CloudControlApi, requestToken string) (*cloudcontrolapi.ProgressEvent, error) {
	input := &cloudcontrolapi.GetResourceRequestStatusInput{
		RequestToken: aws.String(requestToken),
	}

	output, err := conn.GetResourceRequestStatusWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudcontrolapi.ErrCodeRequestTokenNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ProgressEvent == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ProgressEvent, nil
}

func FindResourceByID(ctx context.Context, conn *cloudcontrolapi.CloudControlApi, resourceID, typeName, typeVersionID, roleARN string) (*cloudcontrolapi.ResourceDescription, error) {
	input := &cloudcontrolapi.GetResourceInput{
		Identifier: aws.String(resourceID),
		TypeName:   aws.String(typeName),
	}
	if roleARN != "" {
		input.RoleArn = aws.String(roleARN)
	}
	if typeVersionID != "" {
		input.TypeVersionId = aws.String(typeVersionID)
	}

	output, err := conn.GetResourceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudcontrolapi.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	// TEMPORARY:
	// Some CloudFormation Resources do not correctly re-map "not found" errors, instead returning a HandlerFailureException.
	// These should be reported and fixed upstream over time, but for now work around the issue.
	if tfawserr.ErrMessageContains(err, cloudcontrolapi.ErrCodeHandlerFailureException, "not found") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResourceDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResourceDescription, nil
}
