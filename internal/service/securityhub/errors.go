// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
)

var (
	errCodeAccessDeniedException     = (*types.AccessDeniedException)(nil).ErrorCode()
	errCodeBadRequestException       = "BadRequestException"
	errCodeDataUnavailableException  = "DataUnavailableException"
	errCodeInvalidAccessException    = (*types.InvalidAccessException)(nil).ErrorCode()
	errCodeInvalidInputException     = (*types.InvalidInputException)(nil).ErrorCode()
	errCodeResourceConflictException = (*types.ResourceConflictException)(nil).ErrorCode()
	errCodeResourceNotFoundException = (*types.ResourceNotFoundException)(nil).ErrorCode()
)
