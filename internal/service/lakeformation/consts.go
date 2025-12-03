// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"time"
)

const (
	TableNameAllTables   = "ALL_TABLES"
	IAMAllowedPrincipals = "IAM_ALLOWED_PRINCIPALS"
)

const (
	IAMPropagationTimeout = 2 * time.Minute
)
