// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
)

var (
	errCodeResourceNotFoundException = (*awstypes.ResourceNotFoundException)(nil).ErrorCode()
	errCodeValidationException       = (*awstypes.ValidationException)(nil).ErrorCode()
)
