package connect

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// ConnectInstanceCreateTimeout Timeout for connect instance creation
	instanceCreatedTimeout = 5 * time.Minute
	instanceDeletedTimeout = 5 * time.Minute

	botAssociationCreateTimeout = 5 * time.Minute

	phoneNumberCreatedTimeout = 2 * time.Minute
	phoneNumberUpdatedTimeout = 2 * time.Minute
	phoneNumberDeletedTimeout = 2 * time.Minute

	vocabularyCreatedTimeout = 5 * time.Minute
	// It takes about 90 minutes for Amazon Connect to delete a vocabulary.
	// https://docs.aws.amazon.com/connect/latest/adminguide/add-custom-vocabulary.html
	vocabularyDeletedTimeout = 100 * time.Minute
)

func waitInstanceCreated(ctx context.Context, conn *connect.Connect, timeout time.Duration, instanceId string) (*connect.DescribeInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.InstanceStatusCreationInProgress},
		Target:  []string{connect.InstanceStatusActive},
		Refresh: statusInstance(ctx, conn, instanceId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribeInstanceOutput); ok {
		return v, err
	}

	return nil, err
}

// We don't have a PENDING_DELETION or DELETED for the Connect instance.
// If the Connect Instance has an associated EXISTING DIRECTORY, removing the connect instance
// will cause an error because it is still has authorized applications.
func waitInstanceDeleted(ctx context.Context, conn *connect.Connect, timeout time.Duration, instanceId string) (*connect.DescribeInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.InstanceStatusActive},
		Target:  []string{connect.ErrCodeResourceNotFoundException},
		Refresh: statusInstance(ctx, conn, instanceId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribeInstanceOutput); ok {
		return v, err
	}

	return nil, err
}

func waitPhoneNumberCreated(ctx context.Context, conn *connect.Connect, timeout time.Duration, phoneNumberId string) (*connect.DescribePhoneNumberOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.PhoneNumberWorkflowStatusInProgress},
		Target:  []string{connect.PhoneNumberWorkflowStatusClaimed},
		Refresh: statusPhoneNumber(ctx, conn, phoneNumberId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*connect.DescribePhoneNumberOutput); ok {
		if aws.StringValue(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Status) == connect.PhoneNumberWorkflowStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitPhoneNumberUpdated(ctx context.Context, conn *connect.Connect, timeout time.Duration, phoneNumberId string) (*connect.DescribePhoneNumberOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.PhoneNumberWorkflowStatusInProgress},
		Target:  []string{connect.PhoneNumberWorkflowStatusClaimed},
		Refresh: statusPhoneNumber(ctx, conn, phoneNumberId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*connect.DescribePhoneNumberOutput); ok {
		if aws.StringValue(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Status) == connect.PhoneNumberWorkflowStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitPhoneNumberDeleted(ctx context.Context, conn *connect.Connect, timeout time.Duration, phoneNumberId string) (*connect.DescribePhoneNumberOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.PhoneNumberWorkflowStatusInProgress},
		Target:  []string{connect.ErrCodeResourceNotFoundException},
		Refresh: statusPhoneNumber(ctx, conn, phoneNumberId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribePhoneNumberOutput); ok {
		return v, err
	}

	return nil, err
}

func waitVocabularyCreated(ctx context.Context, conn *connect.Connect, timeout time.Duration, instanceId, vocabularyId string) (*connect.DescribeVocabularyOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.VocabularyStateCreationInProgress},
		Target:  []string{connect.VocabularyStateActive, connect.VocabularyStateCreationFailed},
		Refresh: statusVocabulary(ctx, conn, instanceId, vocabularyId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribeVocabularyOutput); ok {
		return v, err
	}

	return nil, err
}

func waitVocabularyDeleted(ctx context.Context, conn *connect.Connect, timeout time.Duration, instanceId, vocabularyId string) (*connect.DescribeVocabularyOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.VocabularyStateDeleteInProgress},
		Target:  []string{connect.ErrCodeResourceNotFoundException},
		Refresh: statusVocabulary(ctx, conn, instanceId, vocabularyId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribeVocabularyOutput); ok {
		return v, err
	}

	return nil, err
}
