package ssm

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	documentDeleteTimeout = 2 * time.Minute
	documentActiveTimeout = 2 * time.Minute
)

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
func waitDocumentActive(conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) {
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
