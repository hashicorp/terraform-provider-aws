package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	lexBotDeleteTimeout      = 5 * time.Minute
	lexIntentDeleteTimeout   = 5 * time.Minute
	lexSlotTypeDeleteTimeout = 5 * time.Minute
)

func waitLexBotDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, botId string) (*lexmodelbuildingservice.GetBotVersionsOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{lexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: statusLexBot(conn, botId),
		Timeout: lexBotDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetBotVersionsOutput); ok {
		return v, err
	}

	return nil, err
}

func waitLexBotAliasDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) (*lexmodelbuildingservice.GetBotAliasOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{lexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: statusLexBotAlias(conn, botAliasName, botName),
		Timeout: lexBotDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetBotAliasOutput); ok {
		return v, err
	}

	return nil, err
}

func waitLexIntentDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, intentId string) (*lexmodelbuildingservice.GetIntentVersionsOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{lexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: statusLexIntent(conn, intentId),
		Timeout: lexIntentDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetIntentVersionsOutput); ok {
		return v, err
	}

	return nil, err
}

func waitLexSlotTypeDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, slotTypeId string) (*lexmodelbuildingservice.GetSlotTypeVersionsOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{lexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: statusLexSlotType(conn, slotTypeId),
		Timeout: lexSlotTypeDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetSlotTypeVersionsOutput); ok {
		return v, err
	}

	return nil, err
}
