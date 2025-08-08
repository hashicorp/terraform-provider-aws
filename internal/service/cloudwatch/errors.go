// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

var (
	errCodeResourceNotFound = (*types.ResourceNotFound)(nil).ErrorCode()
)
