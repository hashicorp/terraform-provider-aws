package globalaccelerator

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAcceleratorByARN(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.Accelerator, error) {
	input := &globalaccelerator.DescribeAcceleratorInput{
		AcceleratorArn: aws.String(arn),
	}

	return FindAccelerator(ctx, conn, input)
}

func FindAccelerator(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeAcceleratorInput) (*globalaccelerator.Accelerator, error) {
	output, err := conn.DescribeAcceleratorWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Accelerator == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Accelerator, nil
}

func FindAcceleratorAttributesByARN(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.AcceleratorAttributes, error) {
	input := &globalaccelerator.DescribeAcceleratorAttributesInput{
		AcceleratorArn: aws.String(arn),
	}

	return FindAcceleratorAttributes(ctx, conn, input)
}

func FindAcceleratorAttributes(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeAcceleratorAttributesInput) (*globalaccelerator.AcceleratorAttributes, error) {
	output, err := conn.DescribeAcceleratorAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AcceleratorAttributes == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AcceleratorAttributes, nil
}

func FindEndpointGroupByARN(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.EndpointGroup, error) {
	input := &globalaccelerator.DescribeEndpointGroupInput{
		EndpointGroupArn: aws.String(arn),
	}

	return FindEndpointGroup(ctx, conn, input)
}

func FindEndpointGroup(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeEndpointGroupInput) (*globalaccelerator.EndpointGroup, error) {
	output, err := conn.DescribeEndpointGroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeEndpointGroupNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.EndpointGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.EndpointGroup, nil
}

func FindListenerByARN(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.Listener, error) {
	input := &globalaccelerator.DescribeListenerInput{
		ListenerArn: aws.String(arn),
	}

	return FindListener(ctx, conn, input)
}

func FindListener(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeListenerInput) (*globalaccelerator.Listener, error) {
	output, err := conn.DescribeListenerWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeListenerNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Listener == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Listener, nil
}
