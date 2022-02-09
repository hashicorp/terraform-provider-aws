package cloudwatchlogs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
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
