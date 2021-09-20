package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// FindAcceleratorByARN returns the accelerator corresponding to the specified ARN.
// Returns NotFoundError if no accelerator is found.
func FindAcceleratorByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.Accelerator, error) {
	input := &globalaccelerator.DescribeAcceleratorInput{
		AcceleratorArn: aws.String(arn),
	}

	return FindAccelerator(conn, input)
}

// FindAccelerator returns the accelerator corresponding to the specified input.
// Returns NotFoundError if no accelerator is found.
func FindAccelerator(conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeAcceleratorInput) (*globalaccelerator.Accelerator, error) {
	output, err := conn.DescribeAccelerator(input)

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
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Accelerator, nil
}

// FindAcceleratorAttributesByARN returns the accelerator attributes corresponding to the specified ARN.
// Returns NotFoundError if no accelerator is found.
func FindAcceleratorAttributesByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.AcceleratorAttributes, error) {
	input := &globalaccelerator.DescribeAcceleratorAttributesInput{
		AcceleratorArn: aws.String(arn),
	}

	return FindAcceleratorAttributes(conn, input)
}

// FindAcceleratorAttributes returns the accelerator attributes corresponding to the specified input.
// Returns NotFoundError if no accelerator is found.
func FindAcceleratorAttributes(conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeAcceleratorAttributesInput) (*globalaccelerator.AcceleratorAttributes, error) {
	output, err := conn.DescribeAcceleratorAttributes(input)

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
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.AcceleratorAttributes, nil
}

// FindEndpointGroupByARN returns the endpoint group corresponding to the specified ARN.
// Returns NotFoundError if no endpoint group is found.
func FindEndpointGroupByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.EndpointGroup, error) {
	input := &globalaccelerator.DescribeEndpointGroupInput{
		EndpointGroupArn: aws.String(arn),
	}

	return FindEndpointGroup(conn, input)
}

// FindEndpointGroup returns the endpoint group corresponding to the specified input.
// Returns NotFoundError if no endpoint group is found.
func FindEndpointGroup(conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeEndpointGroupInput) (*globalaccelerator.EndpointGroup, error) {
	output, err := conn.DescribeEndpointGroup(input)

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
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.EndpointGroup, nil
}

// FindListenerByARN returns the listener corresponding to the specified ARN.
// Returns NotFoundError if no listener is found.
func FindListenerByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.Listener, error) {
	input := &globalaccelerator.DescribeListenerInput{
		ListenerArn: aws.String(arn),
	}

	return FindListener(conn, input)
}

// FindListener returns the listener corresponding to the specified input.
// Returns NotFoundError if no listener is found.
func FindListener(conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeListenerInput) (*globalaccelerator.Listener, error) {
	output, err := conn.DescribeListener(input)

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
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Listener, nil
}
