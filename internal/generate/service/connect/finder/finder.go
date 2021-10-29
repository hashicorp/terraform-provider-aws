package finder

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	tfconnect "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func LambdaFunctionAssociationByArnWithContext(ctx context.Context, conn *connect.Connect, instanceID string, functionArn string) (string, error) {
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

func BotAssociationV1ByNameWithContext(ctx context.Context, conn *connect.Connect, instanceID string, name string) (*connect.LexBot, error) {
	var result *connect.LexBot

	input := &connect.ListBotsInput{
		InstanceId: aws.String(instanceID),
		LexVersion: aws.String(tfconnect.LexBotV1Version),
		MaxResults: aws.Int64(tfconnect.ListBotsMaxResults),
	}

	err := conn.ListBotsPagesWithContext(ctx, input, func(page *connect.ListBotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		for _, cf := range page.LexBots {
			if cf == nil && cf.LexBot == nil {
				continue
			}
			if aws.StringValue(cf.LexBot.Name) == name {
				result = cf.LexBot
				return false
			}
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}
	return result, nil
}

func BotAssociationV2ByAliasArnWithContext(ctx context.Context, conn *connect.Connect, instanceID string, aliasArn string) (*connect.LexV2Bot, error) {
	var result *connect.LexV2Bot

	input := &connect.ListBotsInput{
		InstanceId: aws.String(instanceID),
		LexVersion: aws.String(tfconnect.LexBotV2Version),
		MaxResults: aws.Int64(tfconnect.ListBotsMaxResults),
	}

	err := conn.ListBotsPagesWithContext(ctx, input, func(page *connect.ListBotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		for _, cf := range page.LexBots {
			if cf == nil && cf.LexV2Bot == nil {
				continue
			}
			if aws.StringValue(cf.LexV2Bot.AliasArn) == aliasArn {
				result = cf.LexV2Bot
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
