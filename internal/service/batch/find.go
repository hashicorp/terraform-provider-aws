package batch

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindComputeEnvironmentDetailByName(ctx context.Context, conn *batch.Batch, name string) (*batch.ComputeEnvironmentDetail, error) {
	input := &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: aws.StringSlice([]string{name}),
	}

	computeEnvironmentDetail, err := FindComputeEnvironmentDetail(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(computeEnvironmentDetail.Status); status == batch.CEStatusDeleted {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return computeEnvironmentDetail, nil
}

func FindComputeEnvironmentDetail(ctx context.Context, conn *batch.Batch, input *batch.DescribeComputeEnvironmentsInput) (*batch.ComputeEnvironmentDetail, error) {
	output, err := conn.DescribeComputeEnvironmentsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ComputeEnvironments) == 0 || output.ComputeEnvironments[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO if len(output.ComputeEnvironments) > 1

	return output.ComputeEnvironments[0], nil
}

func FindJobDefinitionByARN(ctx context.Context, conn *batch.Batch, arn string) (*batch.JobDefinition, error) {
	input := &batch.DescribeJobDefinitionsInput{
		JobDefinitions: aws.StringSlice([]string{arn}),
	}

	jobDefinition, err := FindJobDefinition(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(jobDefinition.Status); status == jobDefinitionStatusInactive {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return jobDefinition, nil
}

func FindJobDefinition(ctx context.Context, conn *batch.Batch, input *batch.DescribeJobDefinitionsInput) (*batch.JobDefinition, error) {
	output, err := conn.DescribeJobDefinitionsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.JobDefinitions) == 0 || output.JobDefinitions[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO if len(output.JobDefinitions) > 1

	return output.JobDefinitions[0], nil
}
