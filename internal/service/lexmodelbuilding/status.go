package lexmodelbuilding

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	lexModelBuildingServiceStatusCreated  = "Created"
	lexModelBuildingServiceStatusNotFound = "NotFound"
	lexModelBuildingServiceStatusUnknown  = "Unknown"
)

func statusLexSlotType(conn *lexmodelbuildingservice.LexModelBuildingService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetSlotTypeVersions(&lexmodelbuildingservice.GetSlotTypeVersionsInput{
			Name: aws.String(id),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, lexModelBuildingServiceStatusUnknown, err
		}

		if output == nil || len(output.SlotTypes) == 0 {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}

		return output, lexModelBuildingServiceStatusCreated, nil
	}
}

func statusLexIntent(conn *lexmodelbuildingservice.LexModelBuildingService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetIntentVersions(&lexmodelbuildingservice.GetIntentVersionsInput{
			Name: aws.String(id),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, lexModelBuildingServiceStatusUnknown, err
		}

		if output == nil || len(output.Intents) == 0 {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}

		return output, lexModelBuildingServiceStatusCreated, nil
	}
}

func statusLexBot(conn *lexmodelbuildingservice.LexModelBuildingService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetBotVersions(&lexmodelbuildingservice.GetBotVersionsInput{
			Name: aws.String(id),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, lexModelBuildingServiceStatusUnknown, err
		}

		if output == nil || len(output.Bots) == 0 {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}

		return output, lexModelBuildingServiceStatusCreated, nil
	}
}

func statusLexBotAlias(conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, lexModelBuildingServiceStatusUnknown, err
		}

		if output == nil {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}

		return output, lexModelBuildingServiceStatusCreated, nil
	}
}
