// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	domainAssociationCreatedTimeout  = 5 * time.Minute
	domainAssociationVerifiedTimeout = 15 * time.Minute
)

func waitDomainAssociationCreated(ctx context.Context, conn *amplify.Amplify, appID, domainName string) (*amplify.DomainAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{amplify.DomainStatusCreating, amplify.DomainStatusInProgress, amplify.DomainStatusRequestingCertificate},
		Target:  []string{amplify.DomainStatusPendingVerification, amplify.DomainStatusPendingDeployment, amplify.DomainStatusAvailable},
		Refresh: statusDomainAssociation(ctx, conn, appID, domainName),
		Timeout: domainAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*amplify.DomainAssociation); ok {
		if status := aws.StringValue(v.DomainStatus); status == amplify.DomainStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(v.StatusReason)))
		}

		return v, err
	}

	return nil, err
}

func waitDomainAssociationVerified(ctx context.Context, conn *amplify.Amplify, appID, domainName string) (*amplify.DomainAssociation, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{amplify.DomainStatusUpdating, amplify.DomainStatusInProgress, amplify.DomainStatusPendingVerification},
		Target:  []string{amplify.DomainStatusPendingDeployment, amplify.DomainStatusAvailable},
		Refresh: statusDomainAssociation(ctx, conn, appID, domainName),
		Timeout: domainAssociationVerifiedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*amplify.DomainAssociation); ok {
		if v != nil && aws.StringValue(v.DomainStatus) == amplify.DomainStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(v.StatusReason)))
		}

		return v, err
	}

	return nil, err
}
