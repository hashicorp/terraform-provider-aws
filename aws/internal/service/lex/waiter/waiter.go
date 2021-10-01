package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tflex "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lex"
)

func LexBotCreated(conn *lexmodelbuildingservice.LexModelBuildingService, botId, version string, timeout time.Duration) (*lexmodelbuildingservice.GetBotOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{LexModeBuildingServicesStatusBuilding},
		Target:  []string{LexModeBuildingServicesStatusFailed, LexModeBuildingServicesStatusNotBuilt, LexModeBuildingServicesStatusReady, LexModeBuildingServicesStatusReadyBasicTesting},
		Refresh: LexBotStatus(conn, botId, version),
		Timeout: timeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetBotOutput); ok {
		return v, err
	}

	return nil, err
}

func LexBotDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, botId string, timeout time.Duration) (*lexmodelbuildingservice.GetBotOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: lexmodelbuildingservice.StatusType_Values(),
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: LexBotStatus(conn, botId, tflex.LexBotVersionLatest),
		Timeout: timeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetBotOutput); ok {
		return v, err
	}

	return nil, err
}

func LexBotAliasDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) (*lexmodelbuildingservice.GetBotAliasOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{LexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: LexBotAliasStatus(conn, botAliasName, botName),
		Timeout: tflex.LexBotDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetBotAliasOutput); ok {
		return v, err
	}

	return nil, err
}

func LexIntentDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, intentId string) (*lexmodelbuildingservice.GetIntentVersionsOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{LexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: LexIntentStatus(conn, intentId),
		Timeout: tflex.LexIntentDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetIntentVersionsOutput); ok {
		return v, err
	}

	return nil, err
}

func LexSlotTypeDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, slotTypeId string) (*lexmodelbuildingservice.GetSlotTypeVersionsOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{LexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: LexSlotTypeStatus(conn, slotTypeId),
		Timeout: tflex.LexSlotTypeDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetSlotTypeVersionsOutput); ok {
		return v, err
	}

	return nil, err
}
