package emr

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCluster(conn *emr.EMR, input *emr.DescribeClusterInput) (*emr.Cluster, error) {
	output, err := conn.DescribeCluster(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeClusterNotFound) || tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "is not valid") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Cluster == nil || output.Cluster.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Cluster, nil
}

func FindClusterByID(conn *emr.EMR, id string) (*emr.Cluster, error) {
	input := &emr.DescribeClusterInput{
		ClusterId: aws.String(id),
	}

	output, err := FindCluster(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.Status.State); state == emr.ClusterStateTerminated || state == emr.ClusterStateTerminatedWithErrors {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return output, nil
}
