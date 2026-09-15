// Copyright IBM Corp. 2014, 2026
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

const (
	errCodeOperationError = "operation error"
	errCodeUpdateASG      = "UpdateAutoScalingGroup"
)
