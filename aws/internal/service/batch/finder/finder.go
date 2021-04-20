package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfbatch "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch"
)

func ComputeEnvironmentDetailByName(conn *batch.Batch, name string) (*batch.ComputeEnvironmentDetail, error) {
	input := &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: aws.StringSlice([]string{name}),
	}

	return ComputeEnvironmentDetail(conn, input)
}

func ComputeEnvironmentDetail(conn *batch.Batch, input *batch.DescribeComputeEnvironmentsInput) (*batch.ComputeEnvironmentDetail, error) {
	output, err := conn.DescribeComputeEnvironments(input)

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

	computeEnvironment := output.ComputeEnvironments[0]

	if status := aws.StringValue(computeEnvironment.Status); status == batch.CEStatusDeleted {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return computeEnvironment, nil
}

func JobDefinitionByARN(conn *batch.Batch, arn string) (*batch.JobDefinition, error) {
	input := &batch.DescribeJobDefinitionsInput{
		JobDefinitions: aws.StringSlice([]string{arn}),
	}

	return JobDefinition(conn, input)
}

func JobDefinition(conn *batch.Batch, input *batch.DescribeJobDefinitionsInput) (*batch.JobDefinition, error) {
	output, err := conn.DescribeJobDefinitions(input)

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

	jobDefinition := output.JobDefinitions[0]

	if status := aws.StringValue(jobDefinition.Status); status == tfbatch.JobDefinitionStatusInactive {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return jobDefinition, nil
}
