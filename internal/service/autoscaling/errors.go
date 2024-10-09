// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
)

const (
	errCodeValidationError = "ValidationError"
)

var (
	errCodeResourceInUseFault             = (*awstypes.ResourceInUseFault)(nil).ErrorCode()
	errCodeScalingActivityInProgressFault = (*awstypes.ScalingActivityInProgressFault)(nil).ErrorCode()
)
