package finder

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
)

func ConnectionSummaryByName(ctx context.Context, conn *apprunner.AppRunner, name string) (*apprunner.ConnectionSummary, error) {
	input := &apprunner.ListConnectionsInput{
		ConnectionName: aws.String(name),
	}

	var cs *apprunner.ConnectionSummary

	err := conn.ListConnectionsPagesWithContext(ctx, input, func(page *apprunner.ListConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, c := range page.ConnectionSummaryList {
			if c == nil {
				continue
			}

			if aws.StringValue(c.ConnectionName) == name {
				cs = c
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if cs == nil {
		return nil, nil
	}

	return cs, nil
}
