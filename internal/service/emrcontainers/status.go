package emrcontainers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	statusVirtualClusterNotFound = "NotFound"
	statusVirtualClusterUnknown  = "Unknown"
)

// statusVirtualCluster fetches the virtual cluster and its status
func statusVirtualCluster(conn *emrcontainers.EMRContainers, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vc, err := findVirtualClusterById(conn, id)

		if tfawserr.ErrCodeEquals(err, emrcontainers.ErrCodeResourceNotFoundException) {
			return nil, statusVirtualClusterNotFound, nil
		}

		if err != nil {
			return nil, statusVirtualClusterUnknown, err
		}

		if vc == nil {
			return nil, statusVirtualClusterNotFound, nil
		}

		return vc, aws.StringValue(vc.State), nil
	}
}
