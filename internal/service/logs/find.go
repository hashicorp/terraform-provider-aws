package logs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindQueryDefinition(ctx context.Context, conn *cloudwatchlogs.CloudWatchLogs, name, queryDefinitionID string) (*cloudwatchlogs.QueryDefinition, error) {
	input := &cloudwatchlogs.DescribeQueryDefinitionsInput{}
	if name != "" {
		input.QueryDefinitionNamePrefix = aws.String(name)
	}

	var result *cloudwatchlogs.QueryDefinition
	err := describeQueryDefinitionsPagesWithContext(ctx, conn, input, func(page *cloudwatchlogs.DescribeQueryDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, qd := range page.QueryDefinitions {
			if aws.StringValue(qd.QueryDefinitionId) == queryDefinitionID {
				result = qd
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func FindSubscriptionFilter(conn *cloudwatchlogs.CloudWatchLogs, logGroupName, name string) (*cloudwatchlogs.SubscriptionFilter, error) {
	input := &cloudwatchlogs.DescribeSubscriptionFiltersInput{
		LogGroupName:     aws.String(logGroupName),
		FilterNamePrefix: aws.String(name),
	}

	output, err := conn.DescribeSubscriptionFilters(input)
	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
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

	filters := output.SubscriptionFilters

	if len(filters) == 0 || filters[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return filters[0], nil
}
