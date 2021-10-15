package amplify

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	domainAssociationCreatedTimeout  = 5 * time.Minute
	domainAssociationVerifiedTimeout = 15 * time.Minute
)

func waitDomainAssociationCreated(conn *amplify.Amplify, appID, domainName string) (*amplify.DomainAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{amplify.DomainStatusCreating, amplify.DomainStatusInProgress, amplify.DomainStatusRequestingCertificate},
		Target:  []string{amplify.DomainStatusPendingVerification, amplify.DomainStatusPendingDeployment, amplify.DomainStatusAvailable},
		Refresh: statusDomainAssociation(conn, appID, domainName),
		Timeout: domainAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*amplify.DomainAssociation); ok {
		if status := aws.StringValue(v.DomainStatus); status == amplify.DomainStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(v.StatusReason)))
		}

		return v, err
	}

	return nil, err
}

func waitDomainAssociationVerified(conn *amplify.Amplify, appID, domainName string) (*amplify.DomainAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{amplify.DomainStatusUpdating, amplify.DomainStatusInProgress, amplify.DomainStatusPendingVerification},
		Target:  []string{amplify.DomainStatusPendingDeployment, amplify.DomainStatusAvailable},
		Refresh: statusDomainAssociation(conn, appID, domainName),
		Timeout: domainAssociationVerifiedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*amplify.DomainAssociation); ok {
		if v != nil && aws.StringValue(v.DomainStatus) == amplify.DomainStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(v.StatusReason)))
		}

		return v, err
	}

	return nil, err
}
