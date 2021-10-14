package waiter

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tflex "github.com/hashicorp/terraform-provider-aws/aws/internal/service/lex"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	LexBotAliasDeletedTimeout = 5 * time.Minute
	LexIntentDeletedTimeout   = 5 * time.Minute
	LexSlotTypeDeletedTimeout = 5 * time.Minute
)

func BotVersionCreated(conn *lexmodelbuildingservice.LexModelBuildingService, name, version string, timeout time.Duration) (*lexmodelbuildingservice.GetBotOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{lexmodelbuildingservice.StatusBuilding},
		Target: []string{
			lexmodelbuildingservice.StatusNotBuilt,
			lexmodelbuildingservice.StatusReady,
			lexmodelbuildingservice.StatusReadyBasicTesting,
		},
		Refresh: BotVersionStatus(conn, name, version),
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

func BotDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, name string, timeout time.Duration) (*lexmodelbuildingservice.GetBotOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{
			lexmodelbuildingservice.StatusNotBuilt,
			lexmodelbuildingservice.StatusReady,
			lexmodelbuildingservice.StatusReadyBasicTesting,
		},
		Target:  []string{},
		Refresh: BotVersionStatus(conn, name, tflex.LexBotVersionLatest),
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

func LexBotAliasDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) (*lexmodelbuildingservice.GetBotAliasOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{LexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: LexBotAliasStatus(conn, botAliasName, botName),
		Timeout: LexBotAliasDeletedTimeout,
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
		Timeout: LexIntentDeletedTimeout,
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
		Timeout: LexSlotTypeDeletedTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetSlotTypeVersionsOutput); ok {
		return v, err
	}

	return nil, err
}
