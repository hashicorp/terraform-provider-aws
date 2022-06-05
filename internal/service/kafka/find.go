package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindClusterByARN(ctx context.Context, conn *kafka.Kafka, arn string) (*kafka.ClusterInfo, error) {
	input := &kafka.DescribeClusterInput{
		ClusterArn: aws.String(arn),
	}

	output, err := conn.DescribeClusterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ClusterInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ClusterInfo, nil
}

func FindClusterOperationByARN(ctx context.Context, conn *kafka.Kafka, arn string) (*kafka.ClusterOperationInfo, error) {
	input := &kafka.DescribeClusterOperationInput{
		ClusterOperationArn: aws.String(arn),
	}

	output, err := conn.DescribeClusterOperationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ClusterOperationInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ClusterOperationInfo, nil
}

func FindConfigurationByARN(conn *kafka.Kafka, arn string) (*kafka.DescribeConfigurationOutput, error) {
	input := &kafka.DescribeConfigurationInput{
		Arn: aws.String(arn),
	}

	output, err := conn.DescribeConfiguration(input)

	if tfawserr.ErrMessageContains(err, kafka.ErrCodeBadRequestException, "Configuration ARN does not exist") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// FindScramSecrets returns the matching MSK Cluster's associated secrets
func FindScramSecrets(conn *kafka.Kafka, clusterArn string) ([]*string, error) {
	input := &kafka.ListScramSecretsInput{
		ClusterArn: aws.String(clusterArn),
	}

	var scramSecrets []*string
	err := conn.ListScramSecretsPages(input, func(page *kafka.ListScramSecretsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		scramSecrets = append(scramSecrets, page.SecretArnList...)
		return !lastPage
	})

	return scramSecrets, err
}
