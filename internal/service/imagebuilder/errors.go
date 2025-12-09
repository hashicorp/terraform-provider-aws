// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
)

var (
	errCodeInvalidParameterValueException = (*awstypes.InvalidParameterValueException)(nil).ErrorCode()
	errCodeResourceNotFoundException      = (*awstypes.ResourceNotFoundException)(nil).ErrorCode()
)
