package redshift

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusCluster(conn *redshift.Redshift, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ClusterStatus), nil
	}
}

func availabilityZoneRelocationStatus(cluster *redshift.Cluster) (bool, error) {
	// AvailabilityZoneRelocation is not returned by the API, and AvailabilityZoneRelocationStatus is not implemented as Const at this time.
	switch availabilityZoneRelocationStatus := *cluster.AvailabilityZoneRelocationStatus; availabilityZoneRelocationStatus {
	case "enabled", "pending_enabling":
		return true, nil
	case "disabled", "pending_disabling":
		return false, nil
	default:
		return false, errors.New(fmt.Sprintf("unexpected AvailabilityZoneRelocationStatus attribute value: %s", availabilityZoneRelocationStatus))
	}
}
