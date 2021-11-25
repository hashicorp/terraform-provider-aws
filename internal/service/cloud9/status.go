package cloud9

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusEnvironmentStatus(conn *cloud9.Cloud9, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindEnvironmentByID(conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		lifecycle := out.Lifecycle
		status := aws.StringValue(lifecycle.Status)
		if status == cloud9.EnvironmentStatusError && lifecycle.Reason != nil {
			return out, status, fmt.Errorf("Reason: %s", aws.StringValue(lifecycle.Reason))
		}

		return out, status, nil
	}
}
