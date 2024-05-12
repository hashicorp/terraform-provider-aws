// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

const (
	ErrorCodeAccessDenied = "AccessDeniedException"
)

func isResourceMissing(err error) bool {
	// DataZone returns a 403 when the domain does not exist
	// AccessDeniedException: User is not permitted to perform operation: GetDomain
	return errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not permitted to perform")
}
