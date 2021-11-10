package emr

import (
	"github.com/aws/aws-sdk-go/aws"
	emr "github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusCluster(conn *emr.EMR, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.State), nil
	}
}
