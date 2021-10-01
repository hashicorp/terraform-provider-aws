package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	//Lex Bot Statuses
	LexModeBuildingServicesStatusBuilding          = "BUILDING"
	LexModeBuildingServicesStatusFailed            = "FAILED"
	LexModeBuildingServicesStatusNotBuilt          = "NOT_BUILT"
	LexModeBuildingServicesStatusReady             = "READY"
	LexModeBuildingServicesStatusReadyBasicTesting = "READY_BASIC_TESTING"

	LexModelBuildingServiceStatusCreated  = "CREATED"
	LexModelBuildingServiceStatusNotFound = "NOTFOUND"
	LexModelBuildingServiceStatusUnknown  = "UNKNOWN"
)

func LexSlotTypeStatus(conn *lexmodelbuildingservice.LexModelBuildingService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetSlotTypeVersions(&lexmodelbuildingservice.GetSlotTypeVersionsInput{
			Name: aws.String(id),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, LexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, LexModelBuildingServiceStatusUnknown, err
		}

		if output == nil || len(output.SlotTypes) == 0 {
			return nil, LexModelBuildingServiceStatusNotFound, nil
		}

		return output, LexModelBuildingServiceStatusCreated, nil
	}
}

func LexIntentStatus(conn *lexmodelbuildingservice.LexModelBuildingService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetIntentVersions(&lexmodelbuildingservice.GetIntentVersionsInput{
			Name: aws.String(id),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, LexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, LexModelBuildingServiceStatusUnknown, err
		}

		if output == nil || len(output.Intents) == 0 {
			return nil, LexModelBuildingServiceStatusNotFound, nil
		}

		return output, LexModelBuildingServiceStatusCreated, nil
	}
}

func LexBotStatus(conn *lexmodelbuildingservice.LexModelBuildingService, id, version string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetBot(&lexmodelbuildingservice.GetBotInput{
			Name:           aws.String(id),
			VersionOrAlias: aws.String(version),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, LexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, LexModelBuildingServiceStatusUnknown, err
		}
		if output == nil {
			return nil, LexModelBuildingServiceStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func LexBotAliasStatus(conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, LexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, LexModelBuildingServiceStatusUnknown, err
		}
		if output == nil {
			return nil, LexModelBuildingServiceStatusNotFound, nil
		}

		return output, LexModelBuildingServiceStatusCreated, nil
	}
}
