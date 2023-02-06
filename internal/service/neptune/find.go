package neptune

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindEndpointByID(ctx context.Context, conn *neptune.Neptune, id string) (*neptune.DBClusterEndpoint, error) {
	clusterId, endpointId, err := readClusterEndpointID(id)
	if err != nil {
		return nil, err
	}
	input := &neptune.DescribeDBClusterEndpointsInput{
		DBClusterIdentifier:         aws.String(clusterId),
		DBClusterEndpointIdentifier: aws.String(endpointId),
	}

	output, err := conn.DescribeDBClusterEndpointsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterEndpointNotFoundFault) ||
		tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	endpoints := output.DBClusterEndpoints
	if len(endpoints) == 0 {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return endpoints[0], nil
}

func findGlobalClusterByARN(ctx context.Context, conn *neptune.Neptune, dbClusterARN string) (*neptune.GlobalCluster, error) {
	// Input currently has no filter support, maybe this will change in the future.
	input := &neptune.DescribeGlobalClustersInput{}

	for {
		log.Printf("[DEBUG] Reading Neptune Global Cluster (%s): %s", dbClusterARN, input)
		output, err := conn.DescribeGlobalClustersWithContext(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, gc := range output.GlobalClusters {
			if gc == nil {
				continue
			}
			for _, gcm := range gc.GlobalClusterMembers {
				if gcm == nil {
					continue
				}

				if aws.StringValue(gcm.DBClusterArn) == dbClusterARN {
					return gc, nil
				}
			}
		}

		if output.Marker == nil {
			break
		}

		input.Marker = output.Marker
	}

	// We didn't find the global cluster
	return nil, nil
}

func FindGlobalClusterById(ctx context.Context, conn *neptune.Neptune, globalClusterID string) (*neptune.GlobalCluster, error) {
	input := &neptune.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	for {
		log.Printf("[DEBUG] Reading Neptune Global Cluster (%s): %s", globalClusterID, input)
		output, err := conn.DescribeGlobalClustersWithContext(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, gc := range output.GlobalClusters {
			if gc == nil {
				continue
			}

			if aws.StringValue(gc.GlobalClusterIdentifier) == aws.StringValue(input.GlobalClusterIdentifier) {
				return gc, nil
			}
		}

		if output.Marker == nil {
			break
		}

		input.Marker = output.Marker
	}

	return nil, nil
}

func findClusterByClusterARN(ctx context.Context, conn *neptune.Neptune, arn string) (*neptune.DBCluster, error) {
	input := &neptune.DescribeDBClustersInput{}
	for {
		log.Printf("[DEBUG] Reading Neptune Cluster (%s): %s", arn, input)
		output, err := conn.DescribeDBClustersWithContext(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, dbc := range output.DBClusters {
			if dbc == nil {
				continue
			}

			if aws.StringValue(dbc.DBClusterArn) == arn {
				return dbc, nil
			}
		}

		if output.Marker == nil {
			break
		}

		input.Marker = output.Marker
	}

	// We didn't find the cluster
	return nil, errors.New(neptune.ErrCodeDBClusterNotFoundFault)
}
