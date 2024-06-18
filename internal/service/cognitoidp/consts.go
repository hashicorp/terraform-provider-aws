// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"time"
)

const (
	ResNameResourceServer    = "Resource Server"
	ResNameRiskConfiguration = "Risk Configuration"
	ResNameUserPoolClient    = "User Pool Client"
	ResNameUserPoolDomain    = "User Pool Domain"
	ResNameUser              = "User"
)

const (
	propagationTimeout = 2 * time.Minute
)
