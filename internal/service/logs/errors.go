// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

var (
	errCodeValidationException = (*awstypes.ValidationException)(nil).ErrorCode()
)
