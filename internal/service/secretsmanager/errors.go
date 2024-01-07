// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

var (
	errCodeMalformedPolicyDocumentException = (*types.MalformedPolicyDocumentException)(nil).ErrorCode()
	errCodeResourceNotFoundException        = (*types.ResourceNotFoundException)(nil).ErrorCode()
	errCodeInvalidRequestException          = (*types.InvalidRequestException)(nil).ErrorCode()
)
