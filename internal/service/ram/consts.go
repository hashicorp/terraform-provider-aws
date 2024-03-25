// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"time"
)

const (
	resourceShareInvitationPropagationTimeout = 2 * time.Minute
	resourceSharePropagationTimeout           = 1 * time.Minute
)
