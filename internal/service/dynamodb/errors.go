// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	errCodeThrottlingException       = "ThrottlingException"
	errCodeUnknownOperationException = "UnknownOperationException"
	errCodeValidationException       = "ValidationException"
	errCodeResourceNotFoundException = "ResourceNotFoundException"
)

var (
	errCodeTableNotFoundException = (*awstypes.TableNotFoundException)(nil).ErrorCode()
)
