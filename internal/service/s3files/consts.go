// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
)

// TODO: Replace with awstypes.LifeCycleStateMisconfigured once available in the AWS SDK for Go v2.
const (
	lifeCycleStateMisconfigured awstypes.LifeCycleState = "misconfigured"
)
