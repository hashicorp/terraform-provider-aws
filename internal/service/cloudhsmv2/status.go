package cloudhsmv2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func statusClusterState(conn *cloudhsmv2.CloudHSMV2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := FindCluster(conn, id)

		if err != nil {
			return nil, "", err
		}

		if cluster == nil {
			return nil, "", nil
		}

		return cluster, aws.StringValue(cluster.State), err
	}
}

func statusHSMState(conn *cloudhsmv2.CloudHSMV2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		hsm, err := FindHSM(conn, id, "")

		if err != nil {
			return nil, "", err
		}

		if hsm == nil {
			return nil, "", nil
		}

		return hsm, aws.StringValue(hsm.State), err
	}
}
