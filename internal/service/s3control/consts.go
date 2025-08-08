// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"time"
)

const (
	// General timeout for S3 changes to propagate.
	// See https://docs.aws.amazon.com/AmazonS3/latest/userguide/Welcome.html#ConsistencyModel.
	s3PropagationTimeout = 2 * time.Minute
)
