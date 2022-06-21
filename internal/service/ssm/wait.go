package ssm

import (
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

func waitAssociationSuccess(conn *ssm.SSM, id string, timeout time.Duration) (*ssm.AssociationDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssm.AssociationStatusNamePending},
		Target:  []string{ssm.AssociationStatusNameSuccess},
		Refresh: statusAssociation(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ssm.AssociationDescription); ok && output.Overview != nil {
		if status := aws.StringValue(output.Overview.Status); status == ssm.AssociationStatusNameFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.Overview.DetailedStatus)))
		}
		return output, err
	}

	return nil, err
}

// waitDocumentDeleted waits for an Document to return Deleted
func waitDocumentDeleted(conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssm.DocumentStatusDeleting},
		Target:  []string{},
		Refresh: statusDocument(conn, name),
		Timeout: documentDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ssm.DocumentDescription); ok {
		return output, err
	}

	return nil, err
}

// waitDocumentActive waits for an Document to return Active
func waitDocumentActive(conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssm.DocumentStatusCreating, ssm.DocumentStatusUpdating},
		Target:  []string{ssm.DocumentStatusActive},
		Refresh: statusDocument(conn, name),
		Timeout: documentActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ssm.DocumentDescription); ok {
		return output, err
	}

	return nil, err
}
