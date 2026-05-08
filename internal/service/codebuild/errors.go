// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
)

var (
	errCodeResourceNotFoundException = (*types.ResourceNotFoundException)(nil).ErrorCode()
)
