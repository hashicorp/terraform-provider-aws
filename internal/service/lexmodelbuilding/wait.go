package lexmodelbuilding

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tflex "github.com/hashicorp/terraform-provider-aws/aws/internal/service/lex"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tflexmodelbuilding "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodelbuilding"
	tflexmodelbuilding "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodelbuilding"
)

const (
	lexBotAliasDeletedTimeout = 5 * time.Minute
	lexIntentDeletedTimeout   = 5 * time.Minute
	lexSlotTypeDeletedTimeout = 5 * time.Minute
)

func waitBotVersionCreated(conn *lexmodelbuildingservice.LexModelBuildingService, name, version string, timeout time.Duration) (*lexmodelbuildingservice.GetBotOutput, error) {
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
		Refresh: statusBotVersion(conn, name, tflexmodelbuilding.BotVersionLatest),
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

func waitLexBotAliasDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) (*lexmodelbuildingservice.GetBotAliasOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{lexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: statusLexBotAlias(conn, botAliasName, botName),
		Timeout: lexBotAliasDeletedTimeout,
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
		Timeout: lexIntentDeletedTimeout,
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
		Timeout: lexSlotTypeDeletedTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetSlotTypeVersionsOutput); ok {
		return v, err
	}

	return nil, err
}
