package workmail

import (
	"time"

	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	OrganizationDeletedTimeout = 1 * time.Minute
	OrganizationActiveTimeout  = 1 * time.Minute

	organizationStateCreating  = "Creating"
	organizationStateRequested = "Requested"
	organizationStateActive    = "Active"
	organizationStateDeleted   = "Deleted"
)

func waitOrganizationActive(conn *workmail.WorkMail, id string) (*workmail.DescribeOrganizationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			organizationStateCreating,
			organizationStateRequested,
		},
		Target: []string{
			organizationStateActive,
		},
		Refresh: statusOrganizationState(conn, id),
		Timeout: OrganizationActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*workmail.DescribeOrganizationOutput); ok {
		return v, err
	}

	return nil, err
}

func waitOrganizationDeleted(conn *workmail.WorkMail, id string) (*workmail.DescribeOrganizationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			organizationStateActive,
		},
		Target: []string{
			organizationStateDeleted,
		},
		Refresh: statusOrganizationState(conn, id),
		Timeout: OrganizationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*workmail.DescribeOrganizationOutput); ok {
		return v, err
	}

	return nil, err
}
