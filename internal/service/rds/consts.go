package rds

import "time"

const (
	ClusterRoleStatusActive  = "ACTIVE"
	ClusterRoleStatusDeleted = "DELETED"
	ClusterRoleStatusPending = "PENDING"
)

const (
	StorageTypeStandard = "standard"
	StorageTypeGp2      = "gp2"
	StorageTypeIo1      = "io1"
)

func StorageType_Values() []string {
	return []string{
		StorageTypeStandard,
		StorageTypeGp2,
		StorageTypeIo1,
	}
}

// https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/accessing-monitoring.html#Overview.DBInstance.Status.
const (
	InstanceStatusAvailable                     = "available"
	InstanceStatusBackingUp                     = "backing-up"
	InstanceStatusConfiguringEnhancedMonitoring = "configuring-enhanced-monitoring"
	InstanceStatusConfiguringLogExports         = "configuring-log-exports"
	InstanceStatusCreating                      = "creating"
	InstanceStatusDeleting                      = "deleting"
	InstanceStatusIncompatibleParameters        = "incompatible-parameters"
	InstanceStatusIncompatibleRestore           = "incompatible-restore"
	InstanceStatusModifying                     = "modifying"
	InstanceStatusStarting                      = "starting"
	InstanceStatusStopping                      = "stopping"
	InstanceStatusStorageFull                   = "storage-full"
	InstanceStatusStorageOptimization           = "storage-optimization"
)

const (
	InstanceAutomatedBackupStatusPending     = "pending"
	InstanceAutomatedBackupStatusReplicating = "replicating"
	InstanceAutomatedBackupStatusRetained    = "retained"
)

const (
	EventSubscriptionStatusActive    = "active"
	EventSubscriptionStatusCreating  = "creating"
	EventSubscriptionStatusDeleting  = "deleting"
	EventSubscriptionStatusModifying = "modifying"
)

const (
	EngineAurora           = "aurora"
	EngineAuroraMySQL      = "aurora-mysql"
	EngineAuroraPostgreSQL = "aurora-postgresql"
	EngineMySQL            = "mysql"
	EnginePostgres         = "postgres"
)

func Engine_Values() []string {
	return []string{
		EngineAurora,
		EngineAuroraMySQL,
		EngineAuroraPostgreSQL,
		EngineMySQL,
		EnginePostgres,
	}
}

const (
	EngineModeGlobal        = "global"
	EngineModeMultiMaster   = "multimaster"
	EngineModeParallelQuery = "parallelquery"
	EngineModeProvisioned   = "provisioned"
	EngineModeServerless    = "serverless"
)

func EngineMode_Values() []string {
	return []string{
		EngineModeGlobal,
		EngineModeMultiMaster,
		EngineModeParallelQuery,
		EngineModeProvisioned,
		EngineModeServerless,
	}
}

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

func ClusterExportableLogType_Values() []string {
	return []string{
		ExportableLogTypeAudit,
		ExportableLogTypeError,
		ExportableLogTypeGeneral,
		ExportableLogTypePostgreSQL,
		ExportableLogTypeSlowQuery,
	}
}

func InstanceExportableLogType_Values() []string {
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

const (
	RestoreTypeCopyOnWrite = "copy-on-write"
	RestoreTypeFullCopy    = "full-copy"
)

func RestoreType_Values() []string {
	return []string{
		RestoreTypeCopyOnWrite,
		RestoreTypeFullCopy,
	}
}

const (
	TimeoutActionForceApplyCapacityChange = "ForceApplyCapacityChange"
	TimeoutActionRollbackCapacityChange   = "RollbackCapacityChange"
)

func TimeoutAction_Values() []string {
	return []string{
		TimeoutActionForceApplyCapacityChange,
		TimeoutActionRollbackCapacityChange,
	}
}

const (
	propagationTimeout = 2 * time.Minute
)
