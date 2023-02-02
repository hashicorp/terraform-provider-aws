package rds

import (
	"time"
)

const (
	ClusterRoleStatusActive  = "ACTIVE"
	ClusterRoleStatusDeleted = "DELETED"
	ClusterRoleStatusPending = "PENDING"
)

const (
	ClusterStatusAvailable                  = "available"
	ClusterStatusBackingUp                  = "backing-up"
	ClusterStatusConfiguringIAMDatabaseAuth = "configuring-iam-database-auth"
	ClusterStatusCreating                   = "creating"
	ClusterStatusDeleting                   = "deleting"
	ClusterStatusMigrating                  = "migrating"
	ClusterStatusModifying                  = "modifying"
	ClusterStatusPreparingDataMigration     = "preparing-data-migration"
	ClusterStatusRebooting                  = "rebooting"
	ClusterStatusRenaming                   = "renaming"
	ClusterStatusResettingMasterCredentials = "resetting-master-credentials"
	ClusterStatusUpgrading                  = "upgrading"
)

const (
	storageTypeStandard = "standard"
	storageTypeGP2      = "gp2"
	storageTypeGP3      = "gp3"
	storageTypeIO1      = "io1"
)

func StorageType_Values() []string {
	return []string{
		storageTypeStandard,
		storageTypeGP2,
		storageTypeGP3,
		storageTypeIO1,
	}
}

const (
	InstanceEngineMariaDB             = "mariadb"
	InstanceEngineMySQL               = "mysql"
	InstanceEngineOracleEnterprise    = "oracle-ee"
	InstanceEngineOracleEnterpriseCDB = "oracle-ee-cdb"
	InstanceEngineOracleStandard2     = "oracle-se2"
	InstanceEngineOracleStandard2CDB  = "oracle-se2-cdb"
	InstanceEnginePostgres            = "postgres"
	InstanceEngineSQLServerEnterprise = "sqlserver-ee"
	InstanceEngineSQLServerExpress    = "sqlserver-ex"
	InstanceEngineSQLServerStandard   = "sqlserver-se"
	InstanceEngineSQLServerWeb        = "sqlserver-ewb"
)

// https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/accessing-monitoring.html#Overview.DBInstance.Status.
const (
	InstanceStatusAvailable                                    = "available"
	InstanceStatusBackingUp                                    = "backing-up"
	InstanceStatusConfiguringEnhancedMonitoring                = "configuring-enhanced-monitoring"
	InstanceStatusConfiguringIAMDatabaseAuth                   = "configuring-iam-database-auth"
	InstanceStatusConfiguringLogExports                        = "configuring-log-exports"
	InstanceStatusConvertingToVPC                              = "converting-to-vpc"
	InstanceStatusCreating                                     = "creating"
	InstanceStatusDeleting                                     = "deleting"
	InstanceStatusFailed                                       = "failed"
	InstanceStatusInaccessibleEncryptionCredentials            = "inaccessible-encryption-credentials"
	InstanceStatusInaccessibleEncryptionCredentialsRecoverable = "inaccessible-encryption-credentials-recoverable"
	InstanceStatusIncompatibleNetwork                          = "incompatible-network"
	InstanceStatusIncompatibleOptionGroup                      = "incompatible-option-group"
	InstanceStatusIncompatibleParameters                       = "incompatible-parameters"
	InstanceStatusIncompatibleRestore                          = "incompatible-restore"
	InstanceStatusInsufficentCapacity                          = "insufficient-capacity"
	InstanceStatusMaintenance                                  = "maintenance"
	InstanceStatusModifying                                    = "modifying"
	InstanceStatusMovingToVPC                                  = "moving-to-vpc"
	InstanceStatusRebooting                                    = "rebooting"
	InstanceStatusResettingMasterCredentials                   = "resetting-master-credentials"
	InstanceStatusRenaming                                     = "renaming"
	InstanceStatusRestoreError                                 = "restore-error"
	InstanceStatusStarting                                     = "starting"
	InstanceStatusStopped                                      = "stopped"
	InstanceStatusStopping                                     = "stopping"
	InstanceStatusStorageFull                                  = "storage-full"
	InstanceStatusStorageOptimization                          = "storage-optimization"
	InstanceStatusUpgrading                                    = "upgrading"
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
	ClusterEngineAurora           = "aurora"
	ClusterEngineAuroraMySQL      = "aurora-mysql"
	ClusterEngineAuroraPostgreSQL = "aurora-postgresql"
	ClusterEngineMySQL            = "mysql"
	ClusterEnginePostgres         = "postgres"
)

func ClusterEngine_Values() []string {
	return []string{
		ClusterEngineAurora,
		ClusterEngineAuroraMySQL,
		ClusterEngineAuroraPostgreSQL,
		ClusterEngineMySQL,
		ClusterEnginePostgres,
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
		ExportableLogTypeUpgrade,
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
	NetworkTypeDual = "DUAL"
	NetworkTypeIPv4 = "IPV4"
)

func NetworkType_Values() []string {
	return []string{
		NetworkTypeDual,
		NetworkTypeIPv4,
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

const (
	ResNameTags = "Tags"
)

const (
	ReservedInstanceStateActive         = "active"
	ReservedInstanceStateRetired        = "retired"
	ReservedInstanceStatePaymentPending = "payment-pending"
)
