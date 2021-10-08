package rds

const (
	DBClusterRoleStatusActive  = "ACTIVE"
	DBClusterRoleStatusDeleted = "DELETED"
	DBClusterRoleStatusPending = "PENDING"
)

// https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/accessing-monitoring.html#Overview.DBInstance.Status.
const (
	DBInstanceStatusAvailable                     = "available"
	DBInstanceStatusBackingUp                     = "backing-up"
	DBInstanceStatusConfiguringEnhancedMonitoring = "configuring-enhanced-monitoring"
	DBInstanceStatusConfiguringLogExports         = "configuring-log-exports"
	DBInstanceStatusCreating                      = "creating"
	DBInstanceStatusDeleting                      = "deleting"
	DBInstanceStatusIncompatibleParameters        = "incompatible-parameters"
	DBInstanceStatusModifying                     = "modifying"
	DBInstanceStatusStarting                      = "starting"
	DBInstanceStatusStopping                      = "stopping"
	DBInstanceStatusStorageFull                   = "storage-full"
	DBInstanceStatusStorageOptimization           = "storage-optimization"
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
