package ssm

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	documentDeleteTimeout = 2 * time.Minute
	documentActiveTimeout = 2 * time.Minute
)

func waitAssociationSuccess(ctx context.Context, conn *ssm.SSM, id string, timeout time.Duration) (*ssm.AssociationDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssm.AssociationStatusNamePending},
		Target:  []string{ssm.AssociationStatusNameSuccess},
		Refresh: statusAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssm.AssociationDescription); ok && output.Overview != nil {
		if status := aws.StringValue(output.Overview.Status); status == ssm.AssociationStatusNameFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.Overview.DetailedStatus)))
		}
		return output, err
	}

	return nil, err
}

// waitDocumentDeleted waits for an Document to return Deleted
func waitDocumentDeleted(ctx context.Context, conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssm.DocumentStatusDeleting},
		Target:  []string{},
		Refresh: statusDocument(ctx, conn, name),
		Timeout: documentDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssm.DocumentDescription); ok {
		return output, err
	}

	return nil, err
}

// waitDocumentActive waits for an Document to return Active
func waitDocumentActive(ctx context.Context, conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssm.DocumentStatusCreating, ssm.DocumentStatusUpdating},
		Target:  []string{ssm.DocumentStatusActive},
		Refresh: statusDocument(ctx, conn, name),
		Timeout: documentActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssm.DocumentDescription); ok {
		return output, err
	}

	return nil, err
}

func waitServiceSettingUpdated(ctx context.Context, conn *ssm.SSM, id string, timeout time.Duration) (*ssm.ServiceSetting, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"PendingUpdate", ""},
		Target:  []string{"Customized", "Default"},
		Refresh: statusServiceSetting(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssm.ServiceSetting); ok {
		return output, err
	}

	return nil, err
}

func waitServiceSettingReset(ctx context.Context, conn *ssm.SSM, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Customized", "PendingUpdate", ""},
		Target:  []string{"Default"},
		Refresh: statusServiceSetting(ctx, conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
