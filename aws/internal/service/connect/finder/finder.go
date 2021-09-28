package finder

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	tfconnect "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect"
)

func LambdaFunctionAssociationByFunctionArn(ctx context.Context, conn *connect.Connect, instanceID string, functionArn string) (string, error) {
	var result string

	input := &connect.ListLambdaFunctionsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(tfconnect.ListLambdaFunctionsMaxResults),
	}

	err := conn.ListLambdaFunctionsPagesWithContext(ctx, input, func(page *connect.ListLambdaFunctionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.LambdaFunctions {
			if cf == nil {
				continue
			}

			if aws.StringValue(cf) == functionArn {
				result = functionArn
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	return result, nil
}

func LexBotAssociationByName(ctx context.Context, conn *connect.Connect, instanceID string, name string) (*connect.LexBot, error) {
	var result *connect.LexBot

	input := &connect.ListLexBotsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(tfconnect.ListLexBotsMaxResults),
	}

	err := conn.ListLexBotsPagesWithContext(ctx, input, func(page *connect.ListLexBotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.LexBots {
			if cf == nil {
				continue
			}

			if aws.StringValue(cf.Name) == name {
				result = cf
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
