package globalaccelerator

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

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
