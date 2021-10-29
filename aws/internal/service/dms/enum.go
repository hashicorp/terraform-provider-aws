package dms

const (
	EndpointStatusDeleting = "deleting"
)

const (
	EngineNameAurora                     = "aurora"
	EngineNameAuroraPostgresql           = "aurora-postgresql"
	EngineNameAuroraPostgresqlServerless = "aurora-postgresql-serverless"
	EngineNameAuroraServerless           = "aurora-serverless"
	EngineNameAzuredb                    = "azuredb"
	EngineNameDb2                        = "db2"
	EngineNameDmsTransfer                = "dms-transfer"
	EngineNameDocdb                      = "docdb"
	EngineNameDynamodb                   = "dynamodb"
	EngineNameElasticsearch              = "elasticsearch"
	EngineNameKafka                      = "kafka"
	EngineNameKinesis                    = "kinesis"
	EngineNameMariadb                    = "mariadb"
	EngineNameMongodb                    = "mongodb"
	EngineNameMysql                      = "mysql"
	EngineNameNeptune                    = "neptune"
	EngineNameOpensearch                 = "opensearch"
	EngineNameOracle                     = "oracle"
	EngineNamePostgres                   = "postgres"
	EngineNameRedis                      = "redis"
	EngineNameRedshift                   = "redshift"
	EngineNameS3                         = "s3"
	EngineNameSqlServer                  = "sqlserver"
	EngineNameSybase                     = "sybase"
)

func EngineName_Values() []string {
	return []string{
		EngineNameAurora,
		EngineNameAuroraPostgresql,
		EngineNameAuroraPostgresqlServerless,
		EngineNameAuroraServerless,
		EngineNameAzuredb,
		EngineNameDb2,
		EngineNameDmsTransfer,
		EngineNameDocdb,
		EngineNameDynamodb,
		EngineNameElasticsearch,
		EngineNameKafka,
		EngineNameKinesis,
		EngineNameMariadb,
		EngineNameMongodb,
		EngineNameMysql,
		EngineNameNeptune,
		EngineNameOpensearch,
		EngineNameOracle,
		EngineNamePostgres,
		EngineNameRedis,
		EngineNameRedshift,
		EngineNameS3,
		EngineNameSqlServer,
		EngineNameSybase,
	}
}

const (
	KafkaDefaultTopic = "kafka-default-topic"
)

// https://github.com/aws/aws-sdk-go/issues/2522.
const (
	MongoDbAuthMechanismValueDefault   = "default"
	MongoDbAuthMechanismValueMongodbCr = "mongodb-cr"
	MongoDbAuthMechanismValueScramSha1 = "scram-sha-1"
)

func MongoDbAuthMechanismValue_Values() []string {
	return []string{
		MongoDbAuthMechanismValueDefault,
		MongoDbAuthMechanismValueMongodbCr,
		MongoDbAuthMechanismValueScramSha1,
	}
}

const (
	MongoDbAuthSourceAdmin = "admin"
)

const (
	S3SettingsCompressionTypeGzip = "GZIP"
	S3SettingsCompressionTypeNone = "NONE"
)

func S3SettingsCompressionType_Values() []string {
	return []string{
		S3SettingsCompressionTypeGzip,
		S3SettingsCompressionTypeNone,
	}
}

const (
	S3SettingsEncryptionModeSseKms = "SSE_KMS"
	S3SettingsEncryptionModeSseS3  = "SSE_S3"
)

func S3SettingsEncryptionMode_Values() []string {
	return []string{
		S3SettingsEncryptionModeSseKms,
		S3SettingsEncryptionModeSseS3,
	}
}
