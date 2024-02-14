// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
)

var (
	errCodeResourceNotFoundException = (*types.ResourceNotFoundException)(nil).ErrorCode()
)
