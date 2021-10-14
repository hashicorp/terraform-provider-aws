package globalaccelerator

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// statusAccelerator fetches the Accelerator and its Status
func statusAccelerator(conn *globalaccelerator.GlobalAccelerator, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		accelerator, err := FindAcceleratorByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return accelerator, aws.StringValue(accelerator.Status), nil
	}
}
