// Copyright (c) HashiCorp, Inc.
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
	clusterStatusAvailable                  = "available"
	clusterStatusBackingUp                  = "backing-up"
	clusterStatusConfiguringIAMDatabaseAuth = "configuring-iam-database-auth"
	clusterStatusCreating                   = "creating"
	clusterStatusDeleting                   = "deleting"
	clusterStatusMigrating                  = "migrating"
	clusterStatusModifying                  = "modifying"
	clusterStatusPreparingDataMigration     = "preparing-data-migration"
	clusterStatusPromoting                  = "promoting"
	clusterStatusRebooting                  = "rebooting"
	clusterStatusRenaming                   = "renaming"
	clusterStatusResettingMasterCredentials = "resetting-master-credentials"
	clusterStatusScalingCompute             = "scaling-compute"
	clusterStatusUpgrading                  = "upgrading"

	// Non-standard status values.
	clusterStatusAvailableWithPendingModifiedValues = "tf-available-with-pending-modified-values"
)

const (
	clusterSnapshotStatusAvailable = "available"
	clusterSnapshotStatusCreating  = "creating"
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

func StorageType_Values() []string {
	return []string{
		storageTypeStandard,
		storageTypeGP2,
		storageTypeGP3,
		storageTypeIO1,
		storageTypeIO2,
		storageTypeAuroraIOPT1,
	}
}

const (
	InstanceEngineAuroraMySQL         = "aurora-mysql"
	InstanceEngineAuroraPostgreSQL    = "aurora-postgresql"
	InstanceEngineCustomPrefix        = "custom-"
	InstanceEngineDB2Advanced         = "db2-ae"
	InstanceEngineDB2Standard         = "db2-se"
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
	InstanceEngineSQLServerWeb        = "sqlserver-web"
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
	InstanceStatusDeletePreCheck                               = "delete-precheck"
	InstanceStatusDeleting                                     = "deleting"
	InstanceStatusFailed                                       = "failed"
	InstanceStatusInaccessibleEncryptionCredentials            = "inaccessible-encryption-credentials"
	InstanceStatusInaccessibleEncryptionCredentialsRecoverable = "inaccessible-encryption-credentials-recoverable"
	InstanceStatusIncompatiblCreate                            = "incompatible-create"
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
	GlobalClusterStatusAvailable = "available"
	GlobalClusterStatusCreating  = "creating"
	GlobalClusterStatusDeleting  = "deleting"
	GlobalClusterStatusModifying = "modifying"
	GlobalClusterStatusUpgrading = "upgrading"
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
	ClusterEngineAuroraMySQL      = "aurora-mysql"
	ClusterEngineAuroraPostgreSQL = "aurora-postgresql"
	ClusterEngineMySQL            = "mysql"
	ClusterEnginePostgres         = "postgres"
	ClusterEngineCustomPrefix     = "custom-"
)

func ClusterEngine_Values() []string {
	return []string{
		ClusterEngineAuroraMySQL,
		ClusterEngineAuroraPostgreSQL,
		ClusterEngineMySQL,
		ClusterEnginePostgres,
	}
}

func ClusterInstanceEngine_Values() []string {
	return []string{
		ClusterEngineAuroraMySQL,
		ClusterEngineAuroraPostgreSQL,
		ClusterEngineMySQL,
		ClusterEnginePostgres,
	}
}

const (
	globalClusterEngineAurora           = "aurora"
	globalClusterEngineAuroraMySQL      = "aurora-mysql"
	globalClusterEngineAuroraPostgreSQL = "aurora-postgresql"
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
	ExportableLogTypeAgent      = "agent"
	ExportableLogTypeAlert      = "alert"
	ExportableLogTypeAudit      = "audit"
	ExportableLogTypeDiagLog    = "diag.log"
	ExportableLogTypeError      = "error"
	ExportableLogTypeGeneral    = "general"
	ExportableLogTypeListener   = "listener"
	ExportableLogTypeNotifyLog  = "notify.log"
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
		ExportableLogTypeDiagLog,
		ExportableLogTypeError,
		ExportableLogTypeGeneral,
		ExportableLogTypeListener,
		ExportableLogTypeNotifyLog,
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
	ResNameTags = "Tags"
)

const (
	ReservedInstanceStateActive         = "active"
	ReservedInstanceStateRetired        = "retired"
	ReservedInstanceStatePaymentPending = "payment-pending"
)

const (
	parameterSourceEngineDefault = "engine-default"
	parameterSourceSystem        = "system"
	parameterSourceUser          = "user"
)
