// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusDomainAssociation(ctx context.Context, conn *amplify.Client, appID, domainName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		domainAssociation, err := FindDomainAssociationByAppIDAndDomainName(ctx, conn, appID, domainName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return domainAssociation, string(domainAssociation.DomainStatus), nil
	}
}
