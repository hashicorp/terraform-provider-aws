package rds

const (
	DBClusterRoleStatusActive  = "ACTIVE"
	DBClusterRoleStatusDeleted = "DELETED"
	DBClusterRoleStatusPending = "PENDING"
)

const (
	EventSubscriptionStatusActive    = "active"
	EventSubscriptionStatusCreating  = "creating"
	EventSubscriptionStatusDeleting  = "deleting"
	EventSubscriptionStatusModifying = "modifying"
)

const (
	ExportableLogTypeAgent      = "agent"
	ExportableLogTypeAlert      = "alert"
	ExportableLogTypeAudit      = "audit"
	ExportableLogTypeError      = "error"
	ExportableLogTypeGeneral    = "general"
	ExportableLogTypeListener   = "listener"
	ExportableLogTypeOEMAgent   = "oemagent"
	ExportableLogTypePostgreSQL = "postgresql"
	ExportableLogTypeSlowQuery  = "slowquery"
	ExportableLogTypeTrace      = "trace"
	ExportableLogTypeUpgrade    = "upgrade"
)

func ExportableLogType_Values() []string {
	return []string{
		ExportableLogTypeAgent,
		ExportableLogTypeAlert,
		ExportableLogTypeAudit,
		ExportableLogTypeError,
		ExportableLogTypeGeneral,
		ExportableLogTypeListener,
		ExportableLogTypeOEMAgent,
		ExportableLogTypePostgreSQL,
		ExportableLogTypeSlowQuery,
		ExportableLogTypeTrace,
		ExportableLogTypeUpgrade,
	}
}
