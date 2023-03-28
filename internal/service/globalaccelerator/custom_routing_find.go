package globalaccelerator

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// FindCustomRoutingAcceleratorByARN returns the custom routing accelerator corresponding to the specified ARN.
// Returns NotFoundError if no custom routing accelerator is found.
func FindCustomRoutingAcceleratorByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.CustomRoutingAccelerator, error) {
	input := &globalaccelerator.DescribeCustomRoutingAcceleratorInput{
		AcceleratorArn: aws.String(arn),
	}

	return FindCustomRoutingAccelerator(conn, input)
}

// FindCustomRoutingAccelerator returns the custom routing accelerator corresponding to the specified input.
// Returns NotFoundError if no custom routing accelerator is found.
func FindCustomRoutingAccelerator(conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeCustomRoutingAcceleratorInput) (*globalaccelerator.CustomRoutingAccelerator, error) {
	output, err := conn.DescribeCustomRoutingAccelerator(input)

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

// FindCustomRoutingAcceleratorAttributesByARN returns the custom routing accelerator attributes corresponding to the specified ARN.
// Returns NotFoundError if no custom routing accelerator is found.
func FindCustomRoutingAcceleratorAttributesByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.CustomRoutingAcceleratorAttributes, error) {
	input := &globalaccelerator.DescribeCustomRoutingAcceleratorAttributesInput{
		AcceleratorArn: aws.String(arn),
	}

	return FindCustomRoutingAcceleratorAttributes(conn, input)
}

// FindCustomRoutingAcceleratorAttributes returns the custom routing accelerator attributes corresponding to the specified input.
// Returns NotFoundError if no custom routing accelerator is found.
func FindCustomRoutingAcceleratorAttributes(conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeCustomRoutingAcceleratorAttributesInput) (*globalaccelerator.CustomRoutingAcceleratorAttributes, error) {
	output, err := conn.DescribeCustomRoutingAcceleratorAttributes(input)

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

// FindCustomRoutingEndpointGroupByARN returns the custom routing endpoint group corresponding to the specified ARN.
// Returns NotFoundError if no custom routing endpoint group is found.
func FindCustomRoutingEndpointGroupByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.CustomRoutingEndpointGroup, error) {
	input := &globalaccelerator.DescribeCustomRoutingEndpointGroupInput{
		EndpointGroupArn: aws.String(arn),
	}

	return FindCustomRoutingEndpointGroup(conn, input)
}

// FindCustomRoutingEndpointGroup returns the custom routing endpoint group corresponding to the specified input.
// Returns NotFoundError if no custom routing endpoint group is found.
func FindCustomRoutingEndpointGroup(conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeCustomRoutingEndpointGroupInput) (*globalaccelerator.CustomRoutingEndpointGroup, error) {
	output, err := conn.DescribeCustomRoutingEndpointGroup(input)

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

// FindCustomRoutingListenerByARN returns the custom routing listener corresponding to the specified ARN.
// Returns NotFoundError if no custom routing listener is found.
func FindCustomRoutingListenerByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.CustomRoutingListener, error) {
	input := &globalaccelerator.DescribeCustomRoutingListenerInput{
		ListenerArn: aws.String(arn),
	}

	return FindCustomRoutingListener(conn, input)
}

// FindCustomRoutingListener returns the custom routing listener corresponding to the specified input.
// Returns NotFoundError if no custom routing listener is found.
func FindCustomRoutingListener(conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeCustomRoutingListenerInput) (*globalaccelerator.CustomRoutingListener, error) {
	output, err := conn.DescribeCustomRoutingListener(input)

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
