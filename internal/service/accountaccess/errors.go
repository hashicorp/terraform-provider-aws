// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func isResourceNotFoundError(err error) bool {
	return errs.IsA[*awstypes.ResourceNotFoundException](err)
}

func isEntitlementNotFoundError(err error) bool {
	return isResourceNotFoundError(err) ||
		errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Entitlement not found")
}
