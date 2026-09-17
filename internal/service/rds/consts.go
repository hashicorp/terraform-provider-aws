// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"time"

	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	clusterRoleStatusActive  = "ACTIVE"
	clusterRoleStatusDeleted = "DELETED"
	clusterRoleStatusPending = "PENDING"
)

const (
	clusterStatusAvailable                     = "available"
	clusterStatusBackingUp                     = "backing-up"
	clusterStatusConfiguringEnhancedMonitoring = "configuring-enhanced-monitoring"
	clusterStatusConfiguringIAMDatabaseAuth    = "configuring-iam-database-auth"
	clusterStatusCreating                      = "creating"
	clusterStatusDeleting                      = "deleting"
	clusterStatusMigrating                     = "migrating"
	clusterStatusModifying                     = "modifying"
	clusterStatusPreparingDataMigration        = "preparing-data-migration"
	clusterStatusPromoting                     = "promoting"
	clusterStatusRebooting                     = "rebooting"
	clusterStatusRenaming                      = "renaming"
	clusterStatusResettingMasterCredentials    = "resetting-master-credentials"
	clusterStatusScalingCompute                = "scaling-compute"
	clusterStatusScalingStorage                = "scaling-storage"
	clusterStatusUpgrading                     = "upgrading"

	// Non-standard status values.
	clusterStatusAvailableWithPendingModifiedValues = "tf-available-with-pending-modified-values"
)

const (
	clusterSnapshotStatusAvailable = "available"
	clusterSnapshotStatusCreating  = "creating"
	clusterSnapshotStatusCopying   = "copying"
)

const (
	clusterSnapshotAttributeNameRestore = "restore"
)

const (
	clusterEndpointStatusAvailable = "available"
	clusterEndpointStatusCreating  = "creating"
	clusterEndpointStatusDeleting  = "deleting"
)

const (
	storageTypeStandard    = "standard"
	storageTypeGP2         = "gp2"
	storageTypeGP3         = "gp3"
	storageTypeIO1         = "io1"
	storageTypeIO2         = "io2"
	storageTypeAuroraIOPT1 = "aurora-iopt1"
)

const (
	instanceEngineAuroraMySQL         = "aurora-mysql"
	instanceEngineAuroraPostgreSQL    = "aurora-postgresql"
	instanceEngineDB2Advanced         = "db2-ae"
	instanceEngineDB2Standard         = "db2-se"
	instanceEngineMariaDB             = "mariadb"
	instanceEngineMySQL               = "mysql"
	instanceEngineOracleEnterprise    = "oracle-ee"
	instanceEngineOracleEnterpriseCDB = "oracle-ee-cdb"
	instanceEngineOracleStandard2     = "oracle-se2"
	instanceEngineOracleStandard2CDB  = "oracle-se2-cdb"
	instanceEnginePostgres            = "postgres"
	instanceEngineSQLServerEnterprise = "sqlserver-ee"
	instanceEngineSQLServerExpress    = "sqlserver-ex"
	instanceEngineSQLServerStandard   = "sqlserver-se"
	instanceEngineSQLServerWeb        = "sqlserver-web"

	// Not valid for RDS instances.
	instanceEngineDocDB   = "docdb"
	instanceEngineNeptune = "neptune"

	instanceEngineCustomPrefix = "custom-"
)

// https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/accessing-monitoring.html#Overview.DBInstance.Status.
const (
	instanceStatusAvailable                                    = "available"
	instanceStatusBackingUp                                    = "backing-up"
	instanceStatusConfiguringEnhancedMonitoring                = "configuring-enhanced-monitoring"
	instanceStatusConfiguringIAMDatabaseAuth                   = "configuring-iam-database-auth"
	instanceStatusConfiguringLogExports                        = "configuring-log-exports"
	instanceStatusConvertingToVPC                              = "converting-to-vpc"
	instanceStatusCreating                                     = "creating"
	instanceStatusDeletePreCheck                               = "delete-precheck"
	instanceStatusDeleting                                     = "deleting"
	instanceStatusFailed                                       = "failed"
	instanceStatusInaccessibleEncryptionCredentials            = "inaccessible-encryption-credentials"
	instanceStatusInaccessibleEncryptionCredentialsRecoverable = "inaccessible-encryption-credentials-recoverable"
	instanceStatusIncompatiblCreate                            = "incompatible-create"
	instanceStatusIncompatibleNetwork                          = "incompatible-network"
	instanceStatusIncompatibleOptionGroup                      = "incompatible-option-group"
	instanceStatusIncompatibleParameters                       = "incompatible-parameters"
	instanceStatusIncompatibleRestore                          = "incompatible-restore"
	instanceStatusInsufficentCapacity                          = "insufficient-capacity"
	instanceStatusMaintenance                                  = "maintenance"
	instanceStatusModifying                                    = "modifying"
	instanceStatusMovingToVPC                                  = "moving-to-vpc"
	instanceStatusRebooting                                    = "rebooting"
	instanceStatusResettingMasterCredentials                   = "resetting-master-credentials"
	instanceStatusRenaming                                     = "renaming"
	instanceStatusRestoreError                                 = "restore-error"
	instanceStatusStarting                                     = "starting"
	instanceStatusStopped                                      = "stopped"
	instanceStatusStopping                                     = "stopping"
	instanceStatusStorageConfigUpgrade                         = "storage-config-upgrade"
	instanceStatusStorageFull                                  = "storage-full"
	instanceStatusStorageInitialization                        = "storage-initialization"
	instanceStatusStorageOptimization                          = "storage-optimization"
	instanceStatusUpgrading                                    = "upgrading"
)

const (
	globalClusterStatusAvailable = "available"
	globalClusterStatusCreating  = "creating"
	globalClusterStatusDeleting  = "deleting"
	globalClusterStatusModifying = "modifying"
	globalClusterStatusUpgrading = "upgrading"
)

const (
	eventSubscriptionStatusActive    = "active"
	eventSubscriptionStatusCreating  = "creating"
	eventSubscriptionStatusDeleting  = "deleting"
	eventSubscriptionStatusModifying = "modifying"
)

const (
	dbSnapshotAvailable = "available"
	dbSnapshotCreating  = "creating"
)

const (
	dbSnapshotAttributeNameRestore = "restore"
)

const (
	clusterEngineAuroraMySQL      = "aurora-mysql"
	clusterEngineAuroraPostgreSQL = "aurora-postgresql"
	clusterEngineMySQL            = "mysql"
	clusterEnginePostgres         = "postgres"

	// Not valid for RDS clusters.
	clusterEngineDocDB   = "docdb"
	clusterEngineNeptune = "neptune"
)

func clusterEngine_Values() []string {
	return []string{
		clusterEngineAuroraMySQL,
		clusterEngineAuroraPostgreSQL,
		clusterEngineMySQL,
		clusterEnginePostgres,
	}
}

func clusterInstanceEngine_Values() []string {
	return []string{
		instanceEngineAuroraMySQL,
		instanceEngineAuroraPostgreSQL,
		instanceEngineMySQL,
		instanceEnginePostgres,
	}
}

const (
	globalClusterEngineAurora           = "aurora"
	globalClusterEngineAuroraMySQL      = "aurora-mysql"
	globalClusterEngineAuroraPostgreSQL = "aurora-postgresql"

	// Not valid for RDS global clusters.
	globalClusterEngineDocDB   = "docdb"
	globalClusterEngineNeptune = "neptune"
)

func globalClusterEngine_Values() []string {
	return []string{
		globalClusterEngineAurora,
		globalClusterEngineAuroraMySQL,
		globalClusterEngineAuroraPostgreSQL,
	}
}

const (
	engineModeGlobal        = "global"
	engineModeMultiMaster   = "multimaster"
	engineModeParallelQuery = "parallelquery"
	engineModeProvisioned   = "provisioned"
	engineModeServerless    = "serverless"
)

func engineMode_Values() []string {
	return []string{
		engineModeGlobal,
		engineModeMultiMaster,
		engineModeParallelQuery,
		engineModeProvisioned,
		engineModeServerless,
	}
}

const (
	engineLifecycleSupport         = "open-source-rds-extended-support"
	engineLifecycleSupportDisabled = "open-source-rds-extended-support-disabled"
)

func engineLifecycleSupport_Values() []string {
	return []string{
		engineLifecycleSupport,
		engineLifecycleSupportDisabled,
	}
}

const (
	exportableLogTypeAgent          = "agent"
	exportableLogTypeAlert          = "alert"
	exportableLogTypeAudit          = "audit"
	exportableLogTypeDiagLog        = "diag.log"
	exportableLogTypeError          = "error"
	exportableLogTypeGeneral        = "general"
	exportableLogTypeIAMDBAuthError = "iam-db-auth-error"
	exportableLogTypeInstance       = "instance"
	exportableLogTypeListener       = "listener"
	exportableLogTypeNotifyLog      = "notify.log"
	exportableLogTypeOEMAgent       = "oemagent"
	exportableLogTypePostgreSQL     = "postgresql"
	exportableLogTypeSlowQuery      = "slowquery"
	exportableLogTypeTrace          = "trace"
	exportableLogTypeUpgrade        = "upgrade"
)

func clusterExportableLogType_Values() []string {
	return []string{
		exportableLogTypeAudit,
		exportableLogTypeError,
		exportableLogTypeGeneral,
		exportableLogTypeIAMDBAuthError,
		exportableLogTypeInstance,
		exportableLogTypePostgreSQL,
		exportableLogTypeSlowQuery,
		exportableLogTypeUpgrade,
	}
}

func instanceExportableLogType_Values() []string {
	return []string{
		exportableLogTypeAgent,
		exportableLogTypeAlert,
		exportableLogTypeAudit,
		exportableLogTypeDiagLog,
		exportableLogTypeError,
		exportableLogTypeGeneral,
		exportableLogTypeIAMDBAuthError,
		exportableLogTypeListener,
		exportableLogTypeNotifyLog,
		exportableLogTypeOEMAgent,
		exportableLogTypePostgreSQL,
		exportableLogTypeSlowQuery,
		exportableLogTypeTrace,
		exportableLogTypeUpgrade,
	}
}

const (
	networkTypeDual = "DUAL"
	networkTypeIPv4 = "IPV4"
)

func networkType_Values() []string {
	return []string{
		networkTypeDual,
		networkTypeIPv4,
	}
}

const (
	restoreTypeCopyOnWrite = "copy-on-write"
	restoreTypeFullCopy    = "full-copy"
)

func restoreType_Values() []string {
	return []string{
		restoreTypeCopyOnWrite,
		restoreTypeFullCopy,
	}
}

const (
	timeoutActionForceApplyCapacityChange = "ForceApplyCapacityChange"
	timeoutActionRollbackCapacityChange   = "RollbackCapacityChange"
)

func timeoutAction_Values() []string {
	return []string{
		timeoutActionForceApplyCapacityChange,
		timeoutActionRollbackCapacityChange,
	}
}

const (
	backupTargetOutposts = "outposts"
	backupTargetRegion   = names.AttrRegion
)

func backupTarget_Values() []string {
	return []string{
		backupTargetOutposts,
		names.AttrRegion,
	}
}

const (
	propagationTimeout = 2 * time.Minute
)

const (
	reservedInstanceStateActive         = "active"
	reservedInstanceStateRetired        = "retired"
	reservedInstanceStatePaymentPending = "payment-pending"
)

const (
	parameterSourceEngineDefault = "engine-default"
	parameterSourceSystem        = "system"
	parameterSourceUser          = "user"
)
