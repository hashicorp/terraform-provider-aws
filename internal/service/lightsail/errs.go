// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// Some Operations do not properly return the types.NotFoundException error
// This function matches on the types.NotFoundException or if the error text contains "DoesNotExist"
func IsANotFoundError(err error) bool {
	return errs.IsA[*types.NotFoundException](err) ||
		errs.IsAErrorMessageContains[*errs.ErrorWithMessage](errs.NewErrorWithMessage(err), "DoesNotExist")
}
