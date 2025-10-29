// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"time"
)

type TableType string

const (
	TableTypeTable            TableType = "Table"
	TableTypeTableWithColumns TableType = "TableWithColumns"
)

const (
	TableNameAllTables   = "ALL_TABLES"
	IAMAllowedPrincipals = "IAM_ALLOWED_PRINCIPALS"
)

const (
	IAMPropagationTimeout = 2 * time.Minute
)
