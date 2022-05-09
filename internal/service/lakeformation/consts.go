package lakeformation

import "time"

const (
	TableNameAllTables        = "ALL_TABLES"
	TableTypeTable            = "Table"
	TableTypeTableWithColumns = "TableWithColumns"
	IAMAllowedPrincipals      = "IAM_ALLOWED_PRINCIPALS"
)

const (
	IAMPropagationTimeout = 2 * time.Minute
)
