package dms_test

import (
	"fmt"
	"regexp"
	"testing"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDMSEndpoint_basic(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointBasicConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointBasicConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, "password", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "port", "3303"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "username", "tftestupdate"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_s3(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-s3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointS3Config(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.external_table_definition", ""),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.csv_row_delimiter", "\\n"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.csv_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.bucket_folder", ""),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.bucket_name", "bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.data_format", "csv"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.parquet_version", "parquet-1-0"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.parquet_timestamp_in_millisecond", "false"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.encryption_mode", "SSE_S3"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.server_side_encryption_kms_key_id", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointS3ConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexp.MustCompile(`key=value;`)),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.external_table_definition", "new-external_table_definition"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.csv_row_delimiter", "\\r"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.csv_delimiter", "."),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.bucket_folder", "new-bucket_folder"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.bucket_name", "new-bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.compression_type", "GZIP"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/8009
func TestAccDMSEndpoint_S3_extraConnectionAttributes(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-s3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointS3ExtraConnectionAttributesConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexp.MustCompile(`dataFormat=parquet;`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDMSEndpoint_dynamoDB(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-dynamodb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointDynamoDbConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointDynamoDbConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_elasticSearch(t *testing.T) {
	resourceName := "aws_dms_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointElasticsearchConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
					testAccCheckResourceAttrRegionalHostname(resourceName, "elasticsearch_settings.0.endpoint_uri", "es", "search-estest"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.full_load_error_percentage", "10"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.error_retry_duration", "300"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

// TestAccDMSEndpoint_Elasticsearch_extraConnectionAttributes validates
// extra_connection_attributes handling for "elasticsearch" engine not affected
// by changes made specific to suppressing diffs in the case of "s3"/"mongodb" engine
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/8009
func TestAccDMSEndpoint_Elasticsearch_extraConnectionAttributes(t *testing.T) {
	resourceName := "aws_dms_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointElasticsearchExtraConnectionAttributesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "errorRetryDuration=400;"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDMSEndpoint_Elasticsearch_errorRetryDuration(t *testing.T) {
	resourceName := "aws_dms_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointElasticsearchConfigErrorRetryDuration(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.error_retry_duration", "60"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Resource needs additional creation retry handling for the following:
			// Error creating DMS endpoint: ResourceAlreadyExistsFault: ReplicationEndpoint "xxx" already in use
			// {
			// 	Config: dmsEndpointElasticsearchConfigErrorRetryDuration(rName, 120),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckEndpointExists(resourceName),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.error_retry_duration", "120"),
			// 	),
			// },
		},
	})
}

func TestAccDMSEndpoint_Elasticsearch_fullLoadErrorPercentage(t *testing.T) {
	resourceName := "aws_dms_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointElasticsearchConfigFullLoadErrorPercentage(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.full_load_error_percentage", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Resource needs additional creation retry handling for the following:
			// Error creating DMS endpoint: ResourceAlreadyExistsFault: ReplicationEndpoint "xxx" already in use
			// {
			// 	Config: dmsEndpointElasticsearchConfigFullLoadErrorPercentage(rName, 2),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckEndpointExists(resourceName),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.full_load_error_percentage", "2"),
			// 	),
			// },
		},
	})
}

func TestAccDMSEndpoint_kafka(t *testing.T) {
	domainName := acctest.RandomSubdomain()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointKafkaConfig(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_control_details", "false"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_null_and_empty", "false"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_partition_value", "false"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_table_alter_operations", "false"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_transaction_details", "false"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.message_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.message_max_bytes", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.no_hex_prefix", "false"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.partition_include_schema_table", "false"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.sasl_password", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.sasl_username", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.security_protocol", "plaintext"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.ssl_ca_certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.ssl_client_certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.ssl_client_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.ssl_client_key_password", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.topic", "kafka-default-topic"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointKafkaConfigUpdate(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_control_details", "true"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_null_and_empty", "true"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_partition_value", "true"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_table_alter_operations", "true"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_transaction_details", "true"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.message_format", "json-unformatted"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.message_max_bytes", "500000"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.no_hex_prefix", "true"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.partition_include_schema_table", "true"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.sasl_password", "tftest-new"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.sasl_username", "tftest-new"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.security_protocol", "sasl-ssl"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.ssl_ca_certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.ssl_client_certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.ssl_client_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.ssl_client_key_password", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.topic", "topic1"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_kinesis(t *testing.T) {
	resourceName := "aws_dms_endpoint.test"
	iamRoleResourceName := "aws_iam_role.test"
	stream1ResourceName := "aws_kinesis_stream.test1"
	stream2ResourceName := "aws_kinesis_stream.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointKinesisConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_control_details", "false"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_null_and_empty", "false"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_partition_value", "false"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_table_alter_operations", "true"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_transaction_details", "true"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.message_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.partition_include_schema_table", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_settings.0.service_access_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_settings.0.stream_arn", stream1ResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointKinesisConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_control_details", "true"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_null_and_empty", "true"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_partition_value", "true"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_table_alter_operations", "false"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_transaction_details", "false"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.message_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.partition_include_schema_table", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_settings.0.service_access_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_settings.0.stream_arn", stream2ResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_mongoDB(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-mongodb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointMongoDbConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

// TestAccDMSEndpoint_MongoDB_update validates engine-specific
// configured fields and extra_connection_attributes now set in the resource
// per https://github.com/hashicorp/terraform-provider-aws/issues/8009
func TestAccDMSEndpoint_MongoDB_update(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-mongodb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointMongoDbConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				Config: dmsEndpointMongoDbConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, "port", "27018"),
					resource.TestCheckResourceAttr(resourceName, "username", "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, "password", "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, "database_name", "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "require"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexp.MustCompile(`key=value;`)),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.auth_mechanism", "scram-sha-1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.nesting_level", "one"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.extract_doc_id", "true"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.docs_to_investigate", "1001"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDMSEndpoint_Oracle(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-oracledb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointOracleConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDMSEndpoint_Oracle_secretId(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-oracledb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointOracleConfigSecretId(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDMSEndpoint_Oracle_update(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-oracledb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointOracleConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				Config: dmsEndpointOracleConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, "port", "27018"),
					resource.TestCheckResourceAttr(resourceName, "username", "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, "password", "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, "database_name", "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexp.MustCompile(`key=value;`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDMSEndpoint_Postgres(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-oracledb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointPostgresConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDMSEndpoint_Postgres_secretId(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-oracledb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointPostgresConfigSecretId(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDMSEndpoint_Postgres_update(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-oracledb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointPostgresConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				Config: dmsEndpointPostgresConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, "port", "27018"),
					resource.TestCheckResourceAttr(resourceName, "username", "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, "password", "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, "database_name", "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "require"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexp.MustCompile(`key=value;`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDMSEndpoint_docDB(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-docdb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointDocDBConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointDocDBConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, "password", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "port", "27019"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "username", "tftestupdate"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_db2(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := sdkacctest.RandString(8) + "-db2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointDb2Config(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointDb2ConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, "password", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "port", "27019"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "username", "tftestupdate"),
				),
			},
		},
	})
}

// testAccCheckResourceAttrRegionalHostname ensures the Terraform state exactly matches a formatted DNS hostname with region and partition DNS suffix
func testAccCheckResourceAttrRegionalHostname(resourceName, attributeName, serviceName, hostnamePrefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := fmt.Sprintf("%s.%s.%s.%s", hostnamePrefix, serviceName, acctest.Region(), acctest.PartitionDNSSuffix())

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

func testAccCheckEndpointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dms_endpoint" {
			continue
		}

		_, err := tfdms.FindEndpointByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DMS Endpoint %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckEndpointExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DMS Endpoint ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn

		_, err := tfdms.FindEndpointByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func dmsEndpointBasicConfig(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "aurora"
  extra_connection_attributes = ""
  password                    = "tftest"
  port                        = 3306
  server_name                 = "tftest"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, randId)
}

func dmsEndpointBasicConfigUpdate(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "aurora"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 3303
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, randId)
}

func dmsEndpointDynamoDbConfig(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id         = "tf-test-dms-endpoint-%[1]s"
  endpoint_type       = "target"
  engine_name         = "dynamodb"
  service_access_role = aws_iam_role.iam_role.arn
  ssl_mode            = "none"

  tags = {
    Name   = "tf-test-dynamodb-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  depends_on = [aws_iam_role_policy.dms_dynamodb_access]
}

resource "aws_iam_role" "iam_role" {
  name = "tf-test-iam-dynamodb-role-%[1]s"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_dynamodb_access" {
  name = "tf-test-iam-dynamodb-role-policy-%[1]s"
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"dynamodb:PutItem",
				"dynamodb:CreateTable",
				"dynamodb:DescribeTable",
				"dynamodb:DeleteTable",
				"dynamodb:DeleteItem",
				"dynamodb:ListTables"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, randId)
}

func dmsEndpointDynamoDbConfigUpdate(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id         = "tf-test-dms-endpoint-%[1]s"
  endpoint_type       = "target"
  engine_name         = "dynamodb"
  service_access_role = aws_iam_role.iam_role.arn
  ssl_mode            = "none"

  tags = {
    Name   = "tf-test-dynamodb-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }
}

resource "aws_iam_role" "iam_role" {
  name = "tf-test-iam-dynamodb-role-%[1]s"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_dynamodb_access" {
  name = "tf-test-iam-dynamodb-role-policy-%[1]s"
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"dynamodb:PutItem",
				"dynamodb:CreateTable",
				"dynamodb:DescribeTable",
				"dynamodb:DeleteTable",
				"dynamodb:DeleteItem",
				"dynamodb:ListTables"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, randId)
}

func dmsEndpointS3Config(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "target"
  engine_name                 = "s3"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = "tf-test-s3-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  s3_settings {
    service_access_role_arn = aws_iam_role.iam_role.arn
    bucket_name             = "bucket_name"
  }

  depends_on = [aws_iam_role_policy.dms_s3_access]
}

resource "aws_iam_role" "iam_role" {
  name = "tf-test-iam-s3-role-%[1]s"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_s3_access" {
  name = "tf-test-iam-s3-role-policy-%[1]s"
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"s3:CreateBucket",
				"s3:ListBucket",
				"s3:DeleteBucket",
				"s3:GetBucketLocation",
				"s3:GetObject",
				"s3:PutObject",
				"s3:DeleteObject",
				"s3:GetObjectVersion",
				"s3:GetBucketPolicy",
				"s3:PutBucketPolicy",
				"s3:DeleteBucketPolicy"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, randId)
}

func dmsEndpointS3ExtraConnectionAttributesConfig(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "target"
  engine_name                 = "s3"
  ssl_mode                    = "none"
  extra_connection_attributes = "dataFormat=parquet;"

  s3_settings {
    service_access_role_arn = aws_iam_role.iam_role.arn
    bucket_name             = "bucket_name"
    bucket_folder           = "bucket_folder"
    compression_type        = "GZIP"
  }

  tags = {
    Name   = "tf-test-s3-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  depends_on = [aws_iam_role_policy.dms_s3_access]
}

resource "aws_iam_role" "iam_role" {
  name = "tf-test-iam-s3-role-%[1]s"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_s3_access" {
  name = "tf-test-iam-s3-role-policy-%[1]s"
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"s3:CreateBucket",
				"s3:ListBucket",
				"s3:DeleteBucket",
				"s3:GetBucketLocation",
				"s3:GetObject",
				"s3:PutObject",
				"s3:DeleteObject",
				"s3:GetObjectVersion",
				"s3:GetBucketPolicy",
				"s3:PutBucketPolicy",
				"s3:DeleteBucketPolicy"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, randId)
}

func dmsEndpointS3ConfigUpdate(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "target"
  engine_name                 = "s3"
  ssl_mode                    = "none"
  extra_connection_attributes = "key=value;"

  tags = {
    Name   = "tf-test-s3-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  s3_settings {
    service_access_role_arn   = aws_iam_role.iam_role.arn
    external_table_definition = "new-external_table_definition"
    csv_row_delimiter         = "\\r"
    csv_delimiter             = "."
    bucket_folder             = "new-bucket_folder"
    bucket_name               = "new-bucket_name"
    compression_type          = "GZIP"
  }
}

resource "aws_iam_role" "iam_role" {
  name = "tf-test-iam-s3-role-%[1]s"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_s3_access" {
  name = "tf-test-iam-s3-role-policy-%[1]s"
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"s3:CreateBucket",
				"s3:ListBucket",
				"s3:DeleteBucket",
				"s3:GetBucketLocation",
				"s3:GetObject",
				"s3:PutObject",
				"s3:DeleteObject",
				"s3:GetObjectVersion",
				"s3:GetBucketPolicy",
				"s3:PutBucketPolicy",
				"s3:DeleteBucketPolicy"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, randId)
}

func dmsEndpointElasticsearchConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				 "es:ESHttpDelete",
				 "es:ESHttpGet",
				 "es:ESHttpHead",
				 "es:ESHttpPost",
				 "es:ESHttpPut"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, rName)
}

func dmsEndpointElasticsearchConfig(rName string) string {
	return acctest.ConfigCompose(
		dmsEndpointElasticsearchConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "elasticsearch"

  elasticsearch_settings {
    endpoint_uri            = "search-estest.es.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
    service_access_role_arn = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func dmsEndpointElasticsearchExtraConnectionAttributesConfig(rName string) string {
	return acctest.ConfigCompose(
		dmsEndpointElasticsearchConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "target"
  engine_name                 = "elasticsearch"
  extra_connection_attributes = "errorRetryDuration=400;"
  elasticsearch_settings {
    endpoint_uri               = "search-estest.es.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
    service_access_role_arn    = aws_iam_role.test.arn
    full_load_error_percentage = 20
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func dmsEndpointElasticsearchConfigErrorRetryDuration(rName string, errorRetryDuration int) string {
	return acctest.ConfigCompose(
		dmsEndpointElasticsearchConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "elasticsearch"

  elasticsearch_settings {
    endpoint_uri            = "search-estest.${data.aws_region.current.name}.es.${data.aws_partition.current.dns_suffix}"
    error_retry_duration    = %[2]d
    service_access_role_arn = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, errorRetryDuration))
}

func dmsEndpointElasticsearchConfigFullLoadErrorPercentage(rName string, fullLoadErrorPercentage int) string {
	return acctest.ConfigCompose(
		dmsEndpointElasticsearchConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "elasticsearch"

  elasticsearch_settings {
    endpoint_uri               = "search-estest.${data.aws_region.current.name}.es.${data.aws_partition.current.dns_suffix}"
    full_load_error_percentage = %[2]d
    service_access_role_arn    = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, fullLoadErrorPercentage))
}

func dmsEndpointKafkaConfig(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "kafka"
  ssl_mode      = "none"

  kafka_settings {
    broker                 = "%[2]s:2345"
    include_null_and_empty = false
    security_protocol      = "plaintext"
    no_hex_prefix          = false
  }
}
`, rName, domainName)
}

func dmsEndpointKafkaConfigUpdate(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "kafka"
  ssl_mode      = "none"

  kafka_settings {
    broker                         = "%[2]s:2345"
    topic                          = "topic1"
    message_format                 = "json-unformatted"
    include_transaction_details    = true
    include_partition_value        = true
    partition_include_schema_table = true
    include_table_alter_operations = true
    include_control_details        = true
    message_max_bytes              = 500000
    include_null_and_empty         = true
    security_protocol              = "sasl-ssl"
    sasl_username                  = "tftest-new"
    sasl_password                  = "tftest-new"
    no_hex_prefix                  = true
  }
}
`, rName, domainName)
}

func dmsEndpointKinesisConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_kinesis_stream" "test1" {
  name        = "%[1]s-1"
  shard_count = 1
}

resource "aws_kinesis_stream" "test2" {
  name        = "%[1]s-2"
  shard_count = 1
}

resource "aws_iam_role" "test" {
  name_prefix = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "dms.${data.aws_partition.current.dns_suffix}"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "kinesis:DescribeStream",
      "kinesis:PutRecord",
      "kinesis:PutRecords",
    ]
    resources = [
      aws_kinesis_stream.test1.arn,
      aws_kinesis_stream.test2.arn,
    ]
  }
}
`, rName)
}

func dmsEndpointKinesisConfig(rName string) string {
	return acctest.ConfigCompose(dmsEndpointKinesisConfigBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "kinesis"

  kinesis_settings {
    include_table_alter_operations = true
    include_transaction_details    = true
    partition_include_schema_table = true

    service_access_role_arn = aws_iam_role.test.arn
    stream_arn              = aws_kinesis_stream.test1.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func dmsEndpointKinesisConfigUpdate(rName string) string {
	return acctest.ConfigCompose(dmsEndpointKinesisConfigBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "kinesis"

  kinesis_settings {
    include_control_details        = true
    include_null_and_empty         = true
    include_partition_value        = true
    include_table_alter_operations = false
    include_transaction_details    = false
    partition_include_schema_table = false

    service_access_role_arn = aws_iam_role.test.arn
    stream_arn              = aws_kinesis_stream.test2.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func dmsEndpointMongoDbConfig(randId string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dms" {
  name = "alias/aws/dms"
}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "mongodb"
  server_name                 = "tftest"
  port                        = 27017
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""
  kms_key_arn                 = data.aws_kms_alias.dms.target_key_arn

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  mongodb_settings {
    auth_type           = "password"
    auth_mechanism      = "default"
    nesting_level       = "none"
    extract_doc_id      = "false"
    docs_to_investigate = "1000"
    auth_source         = "admin"
  }
}
`, randId)
}

func dmsEndpointMongoDbConfigUpdate(randId string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dms" {
  name = "alias/aws/dms"
}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "mongodb"
  server_name                 = "tftest-new-server_name"
  port                        = 27018
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "require"
  extra_connection_attributes = "key=value;"
  kms_key_arn                 = data.aws_kms_alias.dms.target_key_arn

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  mongodb_settings {
    auth_mechanism      = "scram-sha-1"
    nesting_level       = "one"
    extract_doc_id      = "true"
    docs_to_investigate = "1001"
  }
}
`, randId)
}

func dmsEndpointOracleConfig(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "oracle"
  server_name                 = "tftest"
  port                        = 27017
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, randId)
}

func dmsEndpointOracleConfigUpdate(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "oracle"
  server_name                 = "tftest-new-server_name"
  port                        = 27018
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "none"
  extra_connection_attributes = "key=value;"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }
}
`, randId)
}

func dmsEndpointOracleConfigSecretId(randId string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dms" {
  name = "alias/aws/dms"
}

data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_secretsmanager_secret" "test" {
  name                    = %[1]q
  recovery_window_in_days = 0
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dms.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Action": "secretsmanager:*",
        "Effect": "Allow",
        "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                     = "tf-test-dms-endpoint-%[1]s"
  endpoint_type                   = "source"
  engine_name                     = "oracle"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, randId)
}

func dmsEndpointPostgresConfig(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "postgres"
  server_name                 = "tftest"
  port                        = 27017
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, randId)
}

func dmsEndpointPostgresConfigUpdate(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "postgres"
  server_name                 = "tftest-new-server_name"
  port                        = 27018
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "require"
  extra_connection_attributes = "key=value;"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }
}
`, randId)
}

func dmsEndpointPostgresConfigSecretId(randId string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dms" {
  name = "alias/aws/dms"
}

data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_secretsmanager_secret" "test" {
  name                    = %[1]q
  recovery_window_in_days = 0
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dms.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Action": "secretsmanager:*",
        "Effect": "Allow",
        "Resource": "*"
    }
  ]
}
EOF
}
resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                     = "tf-test-dms-endpoint-%[1]s"
  endpoint_type                   = "source"
  engine_name                     = "postgres"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, randId)
}

func dmsEndpointDocDBConfig(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "target"
  engine_name                 = "docdb"
  extra_connection_attributes = ""
  password                    = "tftest"
  port                        = 27017
  server_name                 = "tftest"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, randId)
}

func dmsEndpointDocDBConfigUpdate(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "target"
  engine_name                 = "docdb"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 27019
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, randId)
}

func dmsEndpointDb2Config(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "db2"
  extra_connection_attributes = ""
  password                    = "tftest"
  port                        = 27017
  server_name                 = "tftest"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, randId)
}

func dmsEndpointDb2ConfigUpdate(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "db2"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 27019
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, randId)
}
