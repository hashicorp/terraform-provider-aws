package connect

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// ConnectInstanceCreateTimeout Timeout for connect instance creation
	instanceCreatedTimeout = 5 * time.Minute
	instanceDeletedTimeout = 5 * time.Minute

	botAssociationCreateTimeout = 5 * time.Minute

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

	outputRaw, err := stateConf.WaitForState()

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

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*connect.DescribeInstanceOutput); ok {
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

	outputRaw, err := stateConf.WaitForState()

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

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*connect.DescribeVocabularyOutput); ok {
		return v, err
	}

	return nil, err
}
