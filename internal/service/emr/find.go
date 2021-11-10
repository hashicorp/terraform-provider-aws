package emr

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindClusterByID(conn *emr.EMR, id string) (*emr.Cluster, error) {
	input := &emr.DescribeClusterInput{
		ClusterId: aws.String(id),
	}

	output, err := conn.DescribeCluster(input)

	if tfawserr.ErrCodeEquals(err, "ClusterNotFound") || tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "is not valid") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Cluster == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	status := output.Cluster.Status
	state := aws.StringValue(status.State)

	if state == emr.ClusterStateTerminated || state == emr.ClusterStateTerminatedWithErrors {
		return nil, &resource.NotFoundError{
			Message:     aws.StringValue(status.StateChangeReason.Message),
			LastRequest: input,
		}
	}

	return output.Cluster, nil
}
