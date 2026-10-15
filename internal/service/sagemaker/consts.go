// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"time"
)

const (
	propagationTimeout = 2 * time.Minute

	// endpointCreateRetryTimeout bounds the create-time rollback-and-retry loop that
	// recovers from IAM eventual consistency when an endpoint's execution role was just
	// created and is not yet assumable by SageMaker. The loop pauses between attempts (see
	// resourceEndpointCreate) so SageMaker's cached negative role-assumability result can
	// expire rather than being continuously refreshed by back-to-back recreations, so the
	// budget must span several such paced cycles.
	endpointCreateRetryTimeout = 15 * time.Minute
)
