// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"time"
)

const (
	propagationTimeout = 2 * time.Minute
)

const (
	endpointStatusDeleting = "deleting"

	replicationInstanceStatusAvailable = "available"
	replicationInstanceStatusCreating  = "creating"
	replicationInstanceStatusDeleting  = "deleting"
	replicationInstanceStatusModifying = "modifying"
	replicationInstanceStatusUpgrading = "upgrading"

	replicationTaskStatusCreating  = "creating"
	replicationTaskStatusDeleting  = "deleting"
	replicationTaskStatusFailed    = "failed"
	replicationTaskStatusModifying = "modifying"
	replicationTaskStatusMoving    = "moving"
	replicationTaskStatusReady     = "ready"
	replicationTaskStatusStopped   = "stopped"
	replicationTaskStatusStopping  = "stopping"
	replicationTaskStatusRunning   = "running"
	replicationTaskStatusStarting  = "starting"
)

const (
	engineNameAurora                     = "aurora"
	engineNameAuroraPostgresql           = "aurora-postgresql"
	engineNameAuroraPostgresqlServerless = "aurora-postgresql-serverless"
	engineNameAuroraServerless           = "aurora-serverless"
	engineNameAzuredb                    = "azuredb"
	engineNameAzureSQLManagedInstance    = "azure-sql-managed-instance"
	engineNameBabelfish                  = "babelfish"
	engineNameDB2                        = "db2"
	engineNameDB2zOS                     = "db2-zos"
	engineNameTransfer                   = "dms-transfer"
	engineNameDocDB                      = "docdb"
	engineNameDynamoDB                   = "dynamodb"
	engineNameElasticsearch              = "elasticsearch"
	engineNameKafka                      = "kafka"
	engineNameKinesis                    = "kinesis"
	engineNameMariadb                    = "mariadb"
	engineNameMongodb                    = "mongodb"
	engineNameMySQL                      = "mysql"
	engineNameNeptune                    = "neptune"
	engineNameOpenSearch                 = "opensearch"
	engineNameOracle                     = "oracle"
	engineNamePostgres                   = "postgres"
	engineNameRedis                      = "redis"
	engineNameRedshift                   = "redshift"
	engineNameRedshiftServerless         = "redshift-serverless"
	engineNameS3                         = "s3"
	engineNameSQLServer                  = "sqlserver"
	engineNameSybase                     = "sybase"
)

func engineName_Values() []string {
	return []string{
		engineNameAurora,
		engineNameAuroraPostgresql,
		engineNameAuroraPostgresqlServerless,
		engineNameAuroraServerless,
		engineNameAzuredb,
		engineNameAzureSQLManagedInstance,
		engineNameBabelfish,
		engineNameDB2,
		engineNameDB2zOS,
		engineNameTransfer,
		engineNameDocDB,
		engineNameDynamoDB,
		engineNameElasticsearch,
		engineNameKafka,
		engineNameKinesis,
		engineNameMariadb,
		engineNameMongodb,
		engineNameMySQL,
		engineNameNeptune,
		engineNameOpenSearch,
		engineNameOracle,
		engineNamePostgres,
		engineNameRedis,
		engineNameRedshift,
		engineNameRedshiftServerless,
		engineNameSQLServer,
		engineNameSybase,
	}
}

const (
	kafkaDefaultTopic = "kafka-default-topic"
)

// https://github.com/aws/aws-sdk-go/issues/2522.
const (
	mongoDBAuthMechanismValueDefault   = "default"
	mongoDBAuthMechanismValueMongodbCr = "mongodb-cr"
	mongoDBAuthMechanismValueScramSha1 = "scram-sha-1"
)

func mongoDBAuthMechanismValue_Values() []string {
	return []string{
		mongoDBAuthMechanismValueDefault,
		mongoDBAuthMechanismValueMongodbCr,
		mongoDBAuthMechanismValueScramSha1,
	}
}

const (
	mongoDBAuthSourceAdmin = "admin"
)

const (
	encryptionModeSseKMS = "SSE_KMS"
	encryptionModeSseS3  = "SSE_S3"
)

func encryptionMode_Values() []string {
	return []string{
		encryptionModeSseKMS,
		encryptionModeSseS3,
	}
}

const (
	replicationStatusCreated              = "created"
	replicationStatusReady                = "ready"
	replicationStatusRunning              = "running"
	replicationStatusStopping             = "stopping"
	replicationStatusStopped              = "stopped"
	replicationStatusFailed               = "failed"
	replicationStatusInitialising         = "initializing"
	replicationStatusMetadataResources    = "preparing_metadata_resources"
	replicationStatusTestingConnection    = "testing_connection"
	replicationStatusFetchingMetadata     = "fetching_metadata"
	replicationStatusCalculatingCapacity  = "calculating_capacity"
	replicationStatusProvisioningCapacity = "provisioning_capacity"
	replicationStatusReplicationStarting  = "replication_starting"
)

const (
	replicationTypeValueStartReplication = "creating"
	replicationTypeValueResumeProcessing = "resume-processing"
)

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
	eventSubscriptionStatusActive    = "active"
	eventSubscriptionStatusCreating  = "creating"
	eventSubscriptionStatusDeleting  = "deleting"
	eventSubscriptionStatusModifying = "modifying"
)
