package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	PrincipalAssociationTimeout    = 3 * time.Minute
	PrincipalDisassociationTimeout = 3 * time.Minute
)

// WaitResourceShareInvitationAccepted waits for a ResourceShareInvitation to return ACCEPTED
func WaitResourceShareInvitationAccepted(conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShareInvitation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareInvitationStatusPending},
		Target:  []string{ram.ResourceShareInvitationStatusAccepted},
		Refresh: StatusResourceShareInvitation(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShareInvitation); ok {
		return v, err
	}

	return nil, err
}

// WaitResourceShareOwnedBySelfDisassociated waits for a ResourceShare owned by own account to be disassociated
func WaitResourceShareOwnedBySelfDisassociated(conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShare, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated},
		Target:  []string{},
		Refresh: StatusResourceShareOwnerSelf(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShare); ok {
		return v, err
	}

	return nil, err
}

// WaitResourceShareOwnedBySelfActive waits for a ResourceShare owned by own account to return ACTIVE
func WaitResourceShareOwnedBySelfActive(conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShare, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareStatusPending},
		Target:  []string{ram.ResourceShareStatusActive},
		Refresh: StatusResourceShareOwnerSelf(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShare); ok {
		return v, err
	}

	return nil, err
}

// WaitResourceShareOwnedBySelfDeleted waits for a ResourceShare owned by own account to return DELETED
func WaitResourceShareOwnedBySelfDeleted(conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShare, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareStatusDeleting},
		Target:  []string{ram.ResourceShareStatusDeleted},
		Refresh: StatusResourceShareOwnerSelf(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShare); ok {
		return v, err
	}

	return nil, err
}

func WaitResourceSharePrincipalAssociated(conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociating, PrincipalAssociationStatusNotFound},
		Target:  []string{ram.ResourceShareAssociationStatusAssociated},
		Refresh: StatusResourceSharePrincipalAssociation(conn, resourceShareARN, principal),
		Timeout: PrincipalAssociationTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShareAssociation); ok {
		return v, err
	}

	return nil, err
}

func WaitResourceSharePrincipalDisassociated(conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated, ram.ResourceShareAssociationStatusDisassociating},
		Target:  []string{ram.ResourceShareAssociationStatusDisassociated, PrincipalAssociationStatusNotFound},
		Refresh: StatusResourceSharePrincipalAssociation(conn, resourceShareARN, principal),
		Timeout: PrincipalDisassociationTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShareAssociation); ok {
		return v, err
	}

	return nil, err
}
