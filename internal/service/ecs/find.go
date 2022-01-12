package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCapacityProviderByARN(conn *ecs.ECS, arn string) (*ecs.CapacityProvider, error) {
	input := &ecs.DescribeCapacityProvidersInput{
		CapacityProviders: aws.StringSlice([]string{arn}),
		Include:           aws.StringSlice([]string{ecs.CapacityProviderFieldTags}),
	}

	output, err := conn.DescribeCapacityProviders(input)

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

func FindClusterByNameOrARN(conn *ecs.ECS, nameOrARN string) (*ecs.DescribeClustersOutput, error) {
	input := &ecs.DescribeClustersInput{
		Clusters: aws.StringSlice([]string{nameOrARN}),
		Include:  aws.StringSlice([]string{ecs.ClusterFieldTags, ecs.ClusterFieldConfigurations, ecs.ClusterFieldSettings}),
	}

	output, err := conn.DescribeClusters(input)

	if err != nil {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
