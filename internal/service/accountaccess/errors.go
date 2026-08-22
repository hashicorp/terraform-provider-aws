// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// isNotFoundError reports whether err indicates that an Application does not
// exist.
func isNotFoundError(err error) bool {
	return errs.IsA[*awstypes.ResourceNotFoundException](err)
}
