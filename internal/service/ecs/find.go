package ecs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func FindCapacityProviderByARN(conn *ecs.ECS, arn string) (*ecs.CapacityProvider, error) {
	input := &ecs.DescribeCapacityProvidersInput{
		CapacityProviders: aws.StringSlice([]string{arn}),
		Include:           aws.StringSlice([]string{ecs.CapacityProviderFieldTags}),
	}

	output, err := conn.DescribeCapacityProviders(input)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] ECS tagging failed describing Capacity Provider (%s) with tags: %s; retrying without tags", arn, err)

		input.Include = nil
		output, err = conn.DescribeCapacityProviders(input)
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CapacityProviders) == 0 || output.CapacityProviders[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	capacityProvider := output.CapacityProviders[0]

	if status := aws.StringValue(capacityProvider.Status); status == ecs.CapacityProviderStatusInactive {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return capacityProvider, nil
}

func FindClusterByNameOrARN(ctx context.Context, conn *ecs.ECS, nameOrARN string) (*ecs.Cluster, error) {
	input := &ecs.DescribeClustersInput{
		Clusters: aws.StringSlice([]string{nameOrARN}),
		Include:  aws.StringSlice([]string{ecs.ClusterFieldTags, ecs.ClusterFieldConfigurations, ecs.ClusterFieldSettings}),
	}

	output, err := conn.DescribeClustersWithContext(ctx, input)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed describing ECS Cluster (%s) including tags: %s; retrying without tags", nameOrARN, err)

		input.Include = aws.StringSlice([]string{ecs.ClusterFieldConfigurations, ecs.ClusterFieldSettings})
		output, err = conn.DescribeClustersWithContext(ctx, input)
	}

	// Some partitions (i.e., ISO) may not support describe including configuration, giving error
	if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed describing ECS Cluster (%s) including configuration: %s; retrying without configuration", nameOrARN, err)

		input.Include = aws.StringSlice([]string{ecs.ClusterFieldSettings})
		output, err = conn.DescribeClustersWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Clusters) == 0 || output.Clusters[0] == nil {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	if count := len(output.Clusters); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Clusters[0], nil
}
