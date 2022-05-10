package lexmodels

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	lexBotAliasDeletedTimeout = 5 * time.Minute
	lexIntentDeletedTimeout   = 5 * time.Minute
	lexSlotTypeDeletedTimeout = 5 * time.Minute
)

func waitBotVersionCreated(conn *lexmodelbuildingservice.LexModelBuildingService, name, version string, timeout time.Duration) (*lexmodelbuildingservice.GetBotOutput, error) { //nolint:unparam
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{lexmodelbuildingservice.StatusBuilding},
		Target: []string{
			lexmodelbuildingservice.StatusNotBuilt,
			lexmodelbuildingservice.StatusReady,
			lexmodelbuildingservice.StatusReadyBasicTesting,
		},
		Refresh: statusBotVersion(conn, name, version),
		Timeout: timeout,
	}

	outputRaw, err := stateChangeConf.WaitForState()

	if output, ok := outputRaw.(*lexmodelbuildingservice.GetBotOutput); ok {
		if status := aws.StringValue(output.Status); status == lexmodelbuildingservice.StatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitBotDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, name string, timeout time.Duration) (*lexmodelbuildingservice.GetBotOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{
			lexmodelbuildingservice.StatusNotBuilt,
			lexmodelbuildingservice.StatusReady,
			lexmodelbuildingservice.StatusReadyBasicTesting,
		},
		Target:  []string{},
		Refresh: statusBotVersion(conn, name, BotVersionLatest),
		Timeout: timeout,
	}

	outputRaw, err := stateChangeConf.WaitForState()

	if output, ok := outputRaw.(*lexmodelbuildingservice.GetBotOutput); ok {
		if status := aws.StringValue(output.Status); status == lexmodelbuildingservice.StatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitBotAliasDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) (*lexmodelbuildingservice.GetBotAliasOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{lexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: statusBotAlias(conn, botAliasName, botName),
		Timeout: lexBotAliasDeletedTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetBotAliasOutput); ok {
		return v, err
	}

	return nil, err
}

func waitIntentDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, intentId string) (*lexmodelbuildingservice.GetIntentVersionsOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{lexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: statusIntent(conn, intentId),
		Timeout: lexIntentDeletedTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetIntentVersionsOutput); ok {
		return v, err
	}

	return nil, err
}

func waitSlotTypeDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, name string) (*lexmodelbuildingservice.GetSlotTypeOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{lexModelBuildingServiceStatusCreated},
		Target:  []string{},
		Refresh: statusSlotType(conn, name, SlotTypeVersionLatest),
		Timeout: lexSlotTypeDeletedTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetSlotTypeOutput); ok {
		return v, err
	}

	return nil, err
}
