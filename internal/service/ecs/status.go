package ecs

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// AWS will likely add consts for these at some point
	serviceStatusInactive = "INACTIVE"
	serviceStatusActive   = "ACTIVE"
	serviceStatusDraining = "DRAINING"

	serviceStatusError = "ERROR"
	serviceStatusNone  = "NONE"

	clusterStatusError = "ERROR"
	clusterStatusNone  = "NONE"
)

func statusCapacityProvider(conn *ecs.ECS, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCapacityProviderByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusCapacityProviderUpdate(conn *ecs.ECS, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCapacityProviderByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.UpdateStatus), nil
	}
}

func statusService(conn *ecs.ECS, id, cluster string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ecs.DescribeServicesInput{
			Services: aws.StringSlice([]string{id}),
			Cluster:  aws.String(cluster),
		}

		output, err := conn.DescribeServices(input)

		if tfawserr.ErrCodeEquals(err, ecs.ErrCodeServiceNotFoundException) {
			return nil, serviceStatusNone, nil
		}

		if err != nil {
			return nil, serviceStatusError, err
		}

		if len(output.Services) == 0 {
			return nil, serviceStatusNone, nil
		}

		log.Printf("[DEBUG] ECS service (%s) is currently %q", id, *output.Services[0].Status)
		return output, aws.StringValue(output.Services[0].Status), err
	}
}

func statusCluster(conn *ecs.ECS, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByARN(conn, arn)

		if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
			return nil, clusterStatusNone, nil
		}

		if err != nil {
			return nil, clusterStatusError, err
		}

		if len(output.Clusters) == 0 {
			return nil, clusterStatusNone, nil
		}

		return output, aws.StringValue(output.Clusters[0].Status), err
	}
}
