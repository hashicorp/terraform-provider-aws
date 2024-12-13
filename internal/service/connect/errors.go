// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/connect/#pkg-constants
const (
	ErrCodeAccessDeniedException = "AccessDeniedException"
)
