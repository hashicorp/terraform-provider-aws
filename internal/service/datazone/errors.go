// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

const (
	errCodeAccessDenied = "AccessDeniedException"
)

func isResourceMissing(err error) bool {
	// DataZone returns a 403 when the domain does not exist
	// AccessDeniedException: User is not permitted to perform operation: GetDomain
	return errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err)
}

// isResourceNotFound reports only a genuine not-found response. Unlike
// isResourceMissing it does NOT treat AccessDeniedException as absence, so an
// authorization failure is propagated to the caller instead of silently
// removing a still-existing resource from state.
func isResourceNotFound(err error) bool {
	return errs.IsA[*awstypes.ResourceNotFoundException](err)
}
