package waiter

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

const (
	DirectoryCreatedTimeout = 60 * time.Minute
	DirectoryDeletedTimeout = 60 * time.Minute
)

func DirectoryCreated(conn *directoryservice.DirectoryService, id string) (*directoryservice.DirectoryDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageRequested, directoryservice.DirectoryStageCreating, directoryservice.DirectoryStageCreated},
		Target:  []string{directoryservice.DirectoryStageActive},
		Refresh: DirectoryStage(conn, id),
		Timeout: DirectoryCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directoryservice.DirectoryDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StageReason)))

		return output, err
	}

	return nil, err
}

func DirectoryDeleted(conn *directoryservice.DirectoryService, id string) (*directoryservice.DirectoryDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageActive, directoryservice.DirectoryStageDeleting},
		Target:  []string{},
		Refresh: DirectoryStage(conn, id),
		Timeout: DirectoryDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directoryservice.DirectoryDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StageReason)))

		return output, err
	}

	return nil, err
}
