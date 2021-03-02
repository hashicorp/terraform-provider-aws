package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchlogs/lister"
)

func QueryDefinition(conn *cloudwatchlogs.CloudWatchLogs, qName, qId string) (*cloudwatchlogs.QueryDefinition, error) {
	input := &cloudwatchlogs.DescribeQueryDefinitionsInput{
		QueryDefinitionNamePrefix: aws.String(qName),
		MaxResults:                aws.Int64(10),
	}

	var result *cloudwatchlogs.QueryDefinition
	err := lister.DescribeQueryDefinitionsPages(conn, input, func(page *cloudwatchlogs.DescribeQueryDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, qd := range page.QueryDefinitions {
			if aws.StringValue(qd.QueryDefinitionId) == qId {
				result = qd
				return false
			}
		}
		return !lastPage
	})

	return result, err
}
