// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amplify"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amplify/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	domainAssociationCreatedTimeout  = 5 * time.Minute
	domainAssociationVerifiedTimeout = 15 * time.Minute
)

func waitDomainAssociationCreated(ctx context.Context, conn *amplify.Client, appID, domainName string) (*awstypes.DomainAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusCreating, awstypes.DomainStatusInProgress, awstypes.DomainStatusRequestingCertificate),
		Target:  enum.Slice(awstypes.DomainStatusPendingVerification, awstypes.DomainStatusPendingDeployment, awstypes.DomainStatusAvailable),
		Refresh: statusDomainAssociation(ctx, conn, appID, domainName),
		Timeout: domainAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.DomainAssociation); ok {
		if v.DomainStatus == awstypes.DomainStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(v.StatusReason)))
		}

		return v, err
	}

	return nil, err
}

func waitDomainAssociationVerified(ctx context.Context, conn *amplify.Client, appID, domainName string) (*awstypes.DomainAssociation, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusUpdating, awstypes.DomainStatusInProgress, awstypes.DomainStatusPendingVerification),
		Target:  enum.Slice(awstypes.DomainStatusPendingDeployment, awstypes.DomainStatusAvailable),
		Refresh: statusDomainAssociation(ctx, conn, appID, domainName),
		Timeout: domainAssociationVerifiedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.DomainAssociation); ok {
		if v.DomainStatus == awstypes.DomainStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(v.StatusReason)))
		}

		return v, err
	}

	return nil, err
}
