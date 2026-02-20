// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccEndpointConfig_basicUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "3303"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftestupdate"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_Aurora_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_aurora(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_Aurora_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_auroraSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
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

func TestAccDMSEndpoint_Aurora_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_aurora(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				Config: testAccEndpointConfig_auroraUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "3307"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexache.MustCompile(`EventsPollInterval=40;`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_AuroraPostgreSQL_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_auroraPostgreSQL(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_AuroraPostgreSQL_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_auroraPostgreSQLSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
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

func TestAccDMSEndpoint_AuroraPostgreSQL_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_auroraPostgreSQL(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				Config: testAccEndpointConfig_auroraPostgreSQLUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27018"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "require"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexache.MustCompile(`ExecuteTimeout=1000;`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_dynamoDB(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_dynamoDB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccEndpointConfig_dynamoDBUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_OpenSearch_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_openSearch(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
					testAccCheckResourceAttrRegionalHostname(resourceName, "elasticsearch_settings.0.endpoint_uri", "es", "search-estest"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.full_load_error_percentage", "10"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.error_retry_duration", "300"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.use_new_mapping_type", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

// TestAccDMSEndpoint_OpenSearch_extraConnectionAttributes validates
// extra_connection_attributes handling for "elasticsearch" engine not affected
// by changes made specific to suppressing diffs in the case of "s3"/"mongodb" engine
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/8009
func TestAccDMSEndpoint_OpenSearch_extraConnectionAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_openSearchExtraConnectionAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "errorRetryDuration=400;"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_OpenSearch_errorRetryDuration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_openSearchErrorRetryDuration(rName, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.error_retry_duration", "60"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			// Resource needs additional creation retry handling for the following:
			// Error creating DMS endpoint: ResourceAlreadyExistsFault: ReplicationEndpoint "xxx" already in use
			// {
			// 	Config: testAccEndpointConfig_openSearchErrorRetryDuration(rName, 120),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		testAccCheckEndpointExists(resourceName),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.error_retry_duration", "120"),
			// 	),
			// },
		},
	})
}

func TestAccDMSEndpoint_OpenSearch_UseNewMappingType(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_openSearchUseNewMappingType(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.use_new_mapping_type", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_OpenSearch_fullLoadErrorPercentage(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_openSearchFullLoadErrorPercentage(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.full_load_error_percentage", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			// Resource needs additional creation retry handling for the following:
			// Error creating DMS endpoint: ResourceAlreadyExistsFault: ReplicationEndpoint "xxx" already in use
			// {
			// 	Config: testAccEndpointConfig_openSearchFullLoadErrorPercentage(rName, 2),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		testAccCheckEndpointExists(resourceName),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.full_load_error_percentage", "2"),
			// 	),
			// },
		},
	})
}

func TestAccDMSEndpoint_kafka(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.RandomSubdomain()
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dms_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_kafka(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_control_details", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_null_and_empty", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_partition_value", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_table_alter_operations", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_transaction_details", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.message_format", names.AttrJSON),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.message_max_bytes", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.no_hex_prefix", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.partition_include_schema_table", acctest.CtFalse),
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
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccEndpointConfig_kafkaUpdate(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_control_details", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_null_and_empty", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_partition_value", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_table_alter_operations", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.include_transaction_details", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.message_format", "json-unformatted"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.message_max_bytes", "500000"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.no_hex_prefix", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.partition_include_schema_table", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.sasl_password", "tftest-new"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.sasl_username", "tftest-new"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.sasl_mechanism", "plain"),
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
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	iamRoleResourceName := "aws_iam_role.test"
	stream1ResourceName := "aws_kinesis_stream.test1"
	stream2ResourceName := "aws_kinesis_stream.test2"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_kinesis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_control_details", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_null_and_empty", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_partition_value", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_table_alter_operations", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_transaction_details", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.message_format", names.AttrJSON),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.partition_include_schema_table", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.use_large_integer_value", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_settings.0.service_access_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_settings.0.stream_arn", stream1ResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccEndpointConfig_kinesisUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_control_details", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_null_and_empty", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_partition_value", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_table_alter_operations", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.include_transaction_details", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.message_format", names.AttrJSON),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.partition_include_schema_table", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.use_large_integer_value", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_settings.0.service_access_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_settings.0.stream_arn", stream2ResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_MongoDB_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mongoDB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "data.aws_kms_alias.dms", "target_key_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_MongoDB_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mongoDBSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

// TestAccDMSEndpoint_MongoDB_update validates engine-specific
// configured fields and extra_connection_attributes now set in the resource
// per https://github.com/hashicorp/terraform-provider-aws/issues/8009
func TestAccDMSEndpoint_MongoDB_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mongoDB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "data.aws_kms_alias.dms", "target_key_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27017"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
				),
			},
			{
				Config: testAccEndpointConfig_mongoDBUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27018"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "require"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.auth_mechanism", "scram-sha-1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.nesting_level", "one"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.extract_doc_id", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.docs_to_investigate", "1001"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.use_update_lookup", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_MariaDB_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mariaDB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_MariaDB_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mariaDBSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
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

func TestAccDMSEndpoint_MariaDB_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mariaDB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				Config: testAccEndpointConfig_mariaDBUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "3307"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexache.MustCompile(`EventsPollInterval=30;`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_MySQL_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mySQL(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_MySQL_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mySQLSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
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

func TestAccDMSEndpoint_MySQL_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mySQL(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				Config: testAccEndpointConfig_mySQLUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "3307"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexache.MustCompile(`CleanSrcMetadataOnMismatch=false;`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_Oracle_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_oracle(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_Oracle_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_oracleSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
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

func TestAccDMSEndpoint_Oracle_kerberos(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_kerberos(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "oracle_settings.0.authentication_method", "kerberos"),
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
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_oracle(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27017"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", ""),
				),
			},
			{
				Config: testAccEndpointConfig_oracleUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27018"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexache.MustCompile(`charLengthSemantics=CHAR;`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_Oracle_settings(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_oracleSettings(rName, 2000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oracle_settings"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"add_supplemental_logging": knownvalue.Bool(false),
						"read_ahead_blocks":        knownvalue.Int64Exact(2000),
					})})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccEndpointConfig_oracleSettings(rName, 20000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oracle_settings"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"add_supplemental_logging": knownvalue.Bool(false),
						"read_ahead_blocks":        knownvalue.Int64Exact(20000),
					})})),
				},
			},
		},
	})
}

func TestAccDMSEndpoint_PostgreSQL_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_postgreSQL(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_PostgreSQL_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_postgreSQLSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
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

func TestAccDMSEndpoint_PostgreSQL_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_postgreSQL(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				Config: testAccEndpointConfig_postgreSQLUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27018"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "require"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexache.MustCompile(`HeartbeatFrequency=180;`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/23143
func TestAccDMSEndpoint_PostgreSQL_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_postgresKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_MySQL_settings_source(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mySQLSourceSettings(rName, true, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.after_connect_script", "SELECT NOW();"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.authentication_method", "iam"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.clean_source_metadata_on_mismatch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.events_poll_interval", "5"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.execute_timeout", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "mysql_settings.0.service_access_role_arn", "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				// Change events_poll_interval from 5 to 10
				Config: testAccEndpointConfig_mySQLSourceSettings(rName, true, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.after_connect_script", "SELECT NOW();"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.authentication_method", "iam"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.clean_source_metadata_on_mismatch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.events_poll_interval", "10"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.execute_timeout", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "mysql_settings.0.service_access_role_arn", "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				// Remove mysql_settings block (inherited the previous values)
				Config: testAccEndpointConfig_mySQLSourceSettings(rName, false, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.after_connect_script", "SELECT NOW();"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.authentication_method", "iam"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.clean_source_metadata_on_mismatch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.events_poll_interval", "10"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.execute_timeout", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "mysql_settings.0.service_access_role_arn", "aws_iam_role.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_MySQL_settings_target(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_mySQLTargetSettings(rName, true, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.after_connect_script", "SELECT NOW();"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.authentication_method", string(awstypes.MySQLAuthenticationMethodPassword)),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.execute_timeout", "100"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.max_file_size", "1024"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.parallel_load_threads", "1"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.target_db_type", string(awstypes.TargetDbTypeMultipleDatabases)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				// Change execute_timeout from 100 to 60
				Config: testAccEndpointConfig_mySQLTargetSettings(rName, true, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.after_connect_script", "SELECT NOW();"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.authentication_method", string(awstypes.MySQLAuthenticationMethodPassword)),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.execute_timeout", "60"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.max_file_size", "1024"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.parallel_load_threads", "1"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.target_db_type", string(awstypes.TargetDbTypeMultipleDatabases)),
				),
			},
			{
				// Remove mysql_settings block (inherited the previous values)
				Config: testAccEndpointConfig_mySQLTargetSettings(rName, false, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.after_connect_script", "SELECT NOW();"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.authentication_method", string(awstypes.MySQLAuthenticationMethodPassword)),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.execute_timeout", "60"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.max_file_size", "1024"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.parallel_load_threads", "1"),
					resource.TestCheckResourceAttr(resourceName, "mysql_settings.0.target_db_type", string(awstypes.TargetDbTypeMultipleDatabases)),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_PostgreSQL_settings_source(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_postgreSQLSourceSettings(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.after_connect_script", "SET search_path TO pg_catalog,public;"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.authentication_method", "iam"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.capture_ddls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.ddl_artifacts_schema", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.execute_timeout", "100"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.fail_tasks_on_lob_truncation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.heartbeat_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.heartbeat_frequency", "5"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.heartbeat_schema", "test"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.map_boolean_as_boolean", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.map_jsonb_as_clob", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.map_long_varchar_as", "wstring"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.max_file_size", "1024"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.plugin_name", "pglogical"),
					resource.TestCheckResourceAttrSet(resourceName, "postgres_settings.0.service_access_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.slot_name", "test"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_PostgreSQL_settings_target(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_postgreSQLTargetSettings(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.after_connect_script", "SET search_path TO pg_catalog,public;"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.authentication_method", names.AttrPassword),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.babelfish_database_name", "babelfish"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.database_mode", "babelfish"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.execute_timeout", "100"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.max_file_size", "1024"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_PostgreSQL_settings_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Create with heartbeat disabled
				Config: testAccEndpointConfig_postgreSQLHeartbeat(rName, false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "engine_name", "postgres"),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.heartbeat_enable", acctest.CtFalse),
				),
			},
			{
				// Update only nested postgres_settings: enable heartbeat + set schema
				Config: testAccEndpointConfig_postgreSQLHeartbeat(rName, true, "dms_heartbeat"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.heartbeat_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "postgres_settings.0.heartbeat_schema", "dms_heartbeat"),
				),
			},
		},
	})
}

func testAccEndpointConfig_postgreSQLHeartbeat(id string, heartbeat bool, schema string) string {
	schemaLine := ""
	if schema != "" {
		schemaLine = fmt.Sprintf(`heartbeat_schema = %q`, schema)
	}

	// DMS ModifyEndpoint accepts metadata changes without validating connectivity,
	// so placeholder connection values are sufficient for the test
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %q
  endpoint_type = "source"
  engine_name   = "postgres"

  username      = "user"
  password      = "pass"
  server_name   = "example.com"
  database_name = "postgres"
  port          = 5432

  postgres_settings {
    heartbeat_enable = %t
    %s
  }
}
`, id, heartbeat, schemaLine)
}

func TestAccDMSEndpoint_SQLServer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_sqlServer(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_SQLServer_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_sqlServerSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
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

func TestAccDMSEndpoint_SQLServer_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_sqlServer(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				Config: testAccEndpointConfig_sqlServerUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27018"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "require"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexache.MustCompile(`TlogAccessMode=PreferTlog;`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_babelfish(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_babelfish(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				Config: testAccEndpointConfig_babelfishUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27018"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "require"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/23143
func TestAccDMSEndpoint_SQLServer_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_sqlserverKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_Sybase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_sybase(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_Sybase_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_sybaseSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
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

func TestAccDMSEndpoint_Sybase_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_sybase(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27017"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", ""),
				),
			},
			{
				Config: testAccEndpointConfig_sybaseUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27018"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/23143
func TestAccDMSEndpoint_Sybase_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_sybaseKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_docDB(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_docDB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccEndpointConfig_docDBUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27019"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftestupdate"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_db2_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_db2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccEndpointConfig_db2Update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27019"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftestupdate"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_db2zOS_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_db2zOS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccEndpointConfig_db2zOSUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27019"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftestupdate"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_azureSQLManagedInstance(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_azureSQLManagedInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccEndpointConfig_azureSQLManagedInstanceUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "3342"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftestupdate"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_db2_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_db2SecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
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

func TestAccDMSEndpoint_db2zOS_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_db2zOSSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
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

func TestAccDMSEndpoint_redis(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_redis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.auth_password", ""),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.auth_type", "none"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.auth_user_name", ""),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.server_name", "redis1.test"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.ssl_ca_certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.ssl_security_protocol", "plaintext"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEndpointConfig_redisUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.auth_password", "avoid-plaintext-passwords"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.auth_type", "auth-role"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.auth_user_name", "tfacctest"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.server_name", "redis2.test"),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.ssl_ca_certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "redis_settings.0.ssl_security_protocol", "ssl-encryption"),
				),
			},
		},
	})
}

func TestAccDMSEndpoint_Redshift_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_redshift(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", ""),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.bucket_name", ""),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.bucket_folder", ""),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.encryption_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.server_side_encryption_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.service_access_role_arn", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_Redshift_secretID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_redshiftSecretID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "secrets_manager_access_role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "secrets_manager_arn"),
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

func TestAccDMSEndpoint_Redshift_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	iamRoleResourceName := "aws_iam_role.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_redshift(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27017"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest"),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.bucket_name", ""),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.bucket_folder", ""),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.encryption_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.server_side_encryption_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.service_access_role_arn", ""),
				),
			},
			{
				Config: testAccEndpointConfig_redshiftUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "tftest-new-database_name"),
					resource.TestMatchResourceAttr(resourceName, "extra_connection_attributes", regexache.MustCompile(`acceptanydate=true`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "27018"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.bucket_name", names.AttrBucketName),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.bucket_folder", "bucket_folder"),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.encryption_mode", "SSE_S3"),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.server_side_encryption_kms_key_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "redshift_settings.0.service_access_role_arn", iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_Redshift_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_redshiftKMSKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_Redshift_SSEKMSKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_redshiftConnSSEKMSKeyARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.encryption_mode", "SSE_KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "redshift_settings.0.server_side_encryption_kms_key_id", "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_Redshift_SSEKMSKeyId(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_redshiftConnSSEKMSKeyId(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redshift_settings.0.encryption_mode", "SSE_KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "redshift_settings.0.server_side_encryption_kms_key_id", "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDMSEndpoint_pauseReplicationTasks(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	endpointNameSource := "aws_dms_endpoint.source"
	endpointNameTarget := "aws_dms_endpoint.target"
	replicationTaskName := "aws_dms_replication_task.test"
	var task awstypes.ReplicationTask

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_pauseReplicationTasks(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, endpointNameSource),
					testAccCheckEndpointExists(ctx, t, endpointNameTarget),
					testAccCheckReplicationTaskExists(ctx, t, replicationTaskName, &task),
					resource.TestCheckResourceAttr(replicationTaskName, names.AttrStatus, "running"),
				),
			},
			{
				Config: testAccEndpointConfig_pauseReplicationTasks(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, endpointNameSource),
					testAccCheckEndpointExists(ctx, t, endpointNameTarget),
					testAccCheckReplicationTaskExists(ctx, t, replicationTaskName, &task),
					resource.TestCheckResourceAttr(replicationTaskName, names.AttrStatus, "running"),
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

func testAccCheckEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_endpoint" {
				continue
			}

			_, err := tfdms.FindEndpointByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEndpointExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		_, err := tfdms.FindEndpointByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccEndpointConfig_secretBase(rName string) string {
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
        "Service": "dms.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
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
`, rName)
}

func testAccEndpointConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name = "tf-test-dms-db"
  endpoint_id   = %[1]q
  endpoint_type = "source"
  engine_name   = "aurora"
  password      = "tftest"
  port          = 3306
  server_name   = "tftest"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, rName)
}

func testAccEndpointConfig_basicUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "aurora"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 3303
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, rName)
}

func testAccEndpointConfig_aurora(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "aurora"
  server_name                 = "tftest"
  port                        = 3306
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName)
}

func testAccEndpointConfig_auroraSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "source"
  engine_name                     = "aurora"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName))
}

func testAccEndpointConfig_auroraUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "aurora"
  server_name                 = "tftest-new-server_name"
  port                        = 3307
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "none"
  extra_connection_attributes = "EventsPollInterval=40;"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}
`, rName)
}

func testAccEndpointConfig_auroraPostgreSQL(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "aurora-postgresql"
  server_name                 = "tftest"
  port                        = 27017
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName)
}

func testAccEndpointConfig_auroraPostgreSQLSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "source"
  engine_name                     = "aurora-postgresql"
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
`, rName))
}

func testAccEndpointConfig_auroraPostgreSQLUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "aurora-postgresql"
  server_name                 = "tftest-new-server_name"
  port                        = 27018
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "require"
  extra_connection_attributes = "ExecuteTimeout=1000;"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}
`, rName)
}

func testAccEndpointConfig_babelfish(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "babelfish"
  server_name   = "tftest"
  port          = 27017
  username      = "tftest"
  password      = "tftest"
  database_name = "tftest"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName)
}

func testAccEndpointConfig_babelfishUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "babelfish"
  server_name   = "tftest-new-server_name"
  port          = 27018
  username      = "tftest-new-username"
  password      = "tftest-new-password"
  database_name = "tftest-new-database_name"
  ssl_mode      = "require"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}
`, rName)
}

func testAccEndpointConfig_dynamoDB(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "test" {
  endpoint_id         = %[1]q
  endpoint_type       = "target"
  engine_name         = "dynamodb"
  service_access_role = aws_iam_role.iam_role.arn
  ssl_mode            = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  depends_on = [aws_iam_role_policy.dms_dynamodb_access]
}

resource "aws_iam_role" "iam_role" {
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

resource "aws_iam_role_policy" "dms_dynamodb_access" {
  name = %[1]q
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
`, rName)
}

func testAccEndpointConfig_dynamoDBUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "test" {
  endpoint_id         = %[1]q
  endpoint_type       = "target"
  engine_name         = "dynamodb"
  service_access_role = aws_iam_role.iam_role.arn
  ssl_mode            = "none"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}

resource "aws_iam_role" "iam_role" {
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

resource "aws_iam_role_policy" "dms_dynamodb_access" {
  name = %[1]q
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
`, rName)
}

func testAccEndpointConfig_openSearchBase(rName string) string {
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

func testAccEndpointConfig_openSearch(rName string) string {
	return acctest.ConfigCompose(
		testAccEndpointConfig_openSearchBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "opensearch"

  elasticsearch_settings {
    endpoint_uri            = "search-estest.es.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
    service_access_role_arn = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccEndpointConfig_openSearchExtraConnectionAttributes(rName string) string {
	return acctest.ConfigCompose(
		testAccEndpointConfig_openSearchBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "target"
  engine_name                 = "elasticsearch"
  extra_connection_attributes = "errorRetryDuration=400;"
  elasticsearch_settings {
    endpoint_uri               = "search-estest.es.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
    service_access_role_arn    = aws_iam_role.test.arn
    full_load_error_percentage = 20
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccEndpointConfig_openSearchErrorRetryDuration(rName string, errorRetryDuration int) string {
	return acctest.ConfigCompose(
		testAccEndpointConfig_openSearchBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "elasticsearch"

  elasticsearch_settings {
    endpoint_uri            = "search-estest.${data.aws_region.current.region}.es.${data.aws_partition.current.dns_suffix}"
    error_retry_duration    = %[2]d
    service_access_role_arn = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, errorRetryDuration))
}

func testAccEndpointConfig_openSearchUseNewMappingType(rName string, useNewMappingType bool) string {
	return acctest.ConfigCompose(
		testAccEndpointConfig_openSearchBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "elasticsearch"

  elasticsearch_settings {
    endpoint_uri            = "search-estest.${data.aws_region.current.region}.es.${data.aws_partition.current.dns_suffix}"
    use_new_mapping_type    = %[2]t
    service_access_role_arn = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, useNewMappingType))
}

func testAccEndpointConfig_openSearchFullLoadErrorPercentage(rName string, fullLoadErrorPercentage int) string {
	return acctest.ConfigCompose(
		testAccEndpointConfig_openSearchBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "elasticsearch"

  elasticsearch_settings {
    endpoint_uri               = "search-estest.${data.aws_region.current.region}.es.${data.aws_partition.current.dns_suffix}"
    full_load_error_percentage = %[2]d
    service_access_role_arn    = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, fullLoadErrorPercentage))
}

func testAccEndpointConfig_kafka(rName, domainName string) string {
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

func testAccEndpointConfig_kafkaUpdate(rName, domainName string) string {
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
    sasl_mechanism                 = "plain"
    sasl_username                  = "tftest-new"
    sasl_password                  = "tftest-new"
    no_hex_prefix                  = true
  }
}
`, rName, domainName)
}

func testAccEndpointConfig_kinesisBase(rName string) string {
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

func testAccEndpointConfig_kinesis(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_kinesisBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "kinesis"

  kinesis_settings {
    include_table_alter_operations = true
    include_transaction_details    = true
    partition_include_schema_table = true
    use_large_integer_value        = false

    service_access_role_arn = aws_iam_role.test.arn
    stream_arn              = aws_kinesis_stream.test1.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccEndpointConfig_kinesisUpdate(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_kinesisBase(rName), fmt.Sprintf(`
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
    use_large_integer_value        = true

    service_access_role_arn = aws_iam_role.test.arn
    stream_arn              = aws_kinesis_stream.test2.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccEndpointConfig_mongoDB(rName string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dms" {
  name = "alias/aws/dms"
}

resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
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
    Name   = %[1]q
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
`, rName)
}

func testAccEndpointConfig_mongoDBSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "source"
  engine_name                     = "mongodb"
  database_name                   = "tftest"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  tags = {
    Name   = %[1]q
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
`, rName))
}

func testAccEndpointConfig_mongoDBUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dms" {
  name = "alias/aws/dms"
}

resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "source"
  engine_name   = "mongodb"
  server_name   = "tftest-new-server_name"
  port          = 27018
  username      = "tftest-new-username"
  password      = "tftest-new-password"
  database_name = "tftest-new-database_name"
  ssl_mode      = "require"
  kms_key_arn   = data.aws_kms_alias.dms.target_key_arn

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }

  mongodb_settings {
    auth_mechanism      = "scram-sha-1"
    nesting_level       = "one"
    extract_doc_id      = "true"
    docs_to_investigate = "1001"
    use_update_lookup   = true
  }
}
`, rName)
}

func testAccEndpointConfig_mariaDB(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "mariadb"
  server_name                 = "tftest"
  port                        = 3306
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName)
}

func testAccEndpointConfig_mariaDBSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "source"
  engine_name                     = "mariadb"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName))
}

func testAccEndpointConfig_mariaDBUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "mariadb"
  server_name                 = "tftest-new-server_name"
  port                        = 3307
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "none"
  extra_connection_attributes = "EventsPollInterval=30;"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}
`, rName)
}

func testAccEndpointConfig_mySQL(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "mysql"
  server_name                 = "tftest"
  port                        = 3306
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName)
}

func testAccEndpointConfig_mySQLSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "source"
  engine_name                     = "mysql"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName))
}

func testAccEndpointConfig_mySQLUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "mysql"
  server_name                 = "tftest-new-server_name"
  port                        = 3307
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "none"
  extra_connection_attributes = "CleanSrcMetadataOnMismatch=false;"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}
`, rName)
}

func testAccEndpointConfig_oracle(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "source"
  engine_name   = "oracle"
  server_name   = "tftest"
  port          = 27017
  username      = "tftest"
  password      = "tftest"
  database_name = "tftest"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName)
}

func testAccEndpointConfig_oracleUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "oracle"
  server_name                 = "tftest-new-server_name"
  port                        = 27018
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "none"
  extra_connection_attributes = "charLengthSemantics=CHAR;"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}
`, rName)
}

func testAccEndpointConfig_oracleSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "source"
  engine_name                     = "oracle"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName))
}

func testAccEndpointConfig_kerberos(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "source"
  engine_name   = "oracle"

  server_name                 = "tftest"
  port                        = 27017
  username                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  oracle_settings {
    authentication_method = "kerberos"
  }

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName)
}

func testAccEndpointConfig_oracleSettings(rName string, readAheadBlocks int) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "source"
  engine_name   = "oracle"
  server_name   = "tftest"
  port          = 27017
  username      = "tftest"
  password      = "tftest"
  database_name = "tftest"
  ssl_mode      = "none"

  oracle_settings {
    add_supplemental_logging = false
    read_ahead_blocks        = %[2]d
  }
}
`, rName, readAheadBlocks)
}

func testAccEndpointConfig_postgreSQL(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
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
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName)
}

func testAccEndpointConfig_postgreSQLSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
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
`, rName))
}

func testAccEndpointConfig_postgreSQLUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "postgres"
  server_name                 = "tftest-new-server_name"
  port                        = 27018
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "require"
  extra_connection_attributes = "HeartbeatFrequency=180;"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}
`, rName)
}

func testAccEndpointConfig_mySQLSourceSettingsBlock(eventsPollInterval int) string {
	return fmt.Sprintf(`
  mysql_settings {
    after_connect_script              = "SELECT NOW();"
    authentication_method             = "iam"
    clean_source_metadata_on_mismatch = true
    events_poll_interval              = %[1]d
    execute_timeout                   = 100
    service_access_role_arn           = aws_iam_role.test.arn
	server_timezone                   = "UTC"
  }
`, eventsPollInterval)
}

func testAccEndpointConfig_certificateBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_certificate" "dms_certificate" {
  certificate_id  = %[1]q
  certificate_pem = "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIJAJ58TJVjU7G1MA0GCSqGSIb3DQEBBQUAMFExCzAJBgNV\nBAYTAlVTMREwDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYD\nVQQKEwdDaGFydGVyMQwwCgYDVQQLEwNDU0UwHhcNMTcwMTMwMTkyMDA4WhcNMjYx\nMjA5MTkyMDA4WjBRMQswCQYDVQQGEwJVUzERMA8GA1UECBMIQ29sb3JhZG8xDzAN\nBgNVBAcTBkRlbnZlcjEQMA4GA1UEChMHQ2hhcnRlcjEMMAoGA1UECxMDQ1NFMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv6dq6VLIImlAaTrckb5w3X6J\nWP7EGz2ChGAXlkEYto6dPCba0v5+f+8UlMOpeB25XGoai7gdItqNWVFpYsgmndx3\nvTad3ukO1zeElKtw5oHPH2plOaiv/gVJaDa9NTeINj0EtGZs74fCOclAzGFX5vBc\nb08ESWBceRgGjGv3nlij4JzHfqTkCKQz6P6pBivQBfk62rcOkkH5rKoaGltRHROS\nMbkwOhu2hN0KmSYTXRvts0LXnZU4N0l2ms39gmr7UNNNlKYINL2JoTs9dNBc7APD\ndZvlEHd+/FjcLCI8hC3t4g4AbfW0okIBCNG0+oVjqGb2DeONSJKsThahXt89MQID\nAQABo4G0MIGxMB0GA1UdDgQWBBQKq8JxjY1GmeZXJjfOMfW0kBIzPDCBgQYDVR0j\nBHoweIAUCqvCcY2NRpnmVyY3zjH1tJASMzyhVaRTMFExCzAJBgNVBAYTAlVTMREw\nDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYDVQQKEwdDaGFy\ndGVyMQwwCgYDVQQLEwNDU0WCCQCefEyVY1OxtTAMBgNVHRMEBTADAQH/MA0GCSqG\nSIb3DQEBBQUAA4IBAQAWifoMk5kbv+yuWXvFwHiB4dWUUmMlUlPU/E300yVTRl58\np6DfOgJs7MMftd1KeWqTO+uW134QlTt7+jwI8Jq0uyKCu/O2kJhVtH/Ryog14tGl\n+wLcuIPLbwJI9CwZX4WMBrq4DnYss+6F47i8NCc+Z3MAiG4vtq9ytBmaod0dj2bI\ng4/Lac0e00dql9RnqENh1+dF0V+QgTJCoPkMqDNAlSB8vOodBW81UAb2z12t+IFi\n3X9J3WtCK2+T5brXL6itzewWJ2ALvX3QpmZx7fMHJ3tE+SjjyivE1BbOlzYHx83t\nTeYnm7pS9un7A/UzTDHbs7hPUezLek+H3xTPAnnq\n-----END CERTIFICATE-----\n"
}
`, rName)
}

func testAccEndpointConfig_mySQLSourceSettings(rName string, outputMySQLSettings bool, eventsPollInterval int) string {
	mysqlSettings := ""
	if outputMySQLSettings {
		mysqlSettings = testAccEndpointConfig_mySQLSourceSettingsBlock(eventsPollInterval)
	}
	return acctest.ConfigCompose(testAccEndpointConfig_certificateBase(rName), fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dms.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_dms_endpoint" "test" {
  certificate_arn             = aws_dms_certificate.dms_certificate.certificate_arn
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "mysql"
  server_name                 = "tftest"
  port                        = 5432
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "verify-full"
  extra_connection_attributes = ""

  %[2]s
}
`, rName, mysqlSettings))
}

func testAccEndpointConfig_mySQLTargetSettingsBlock(executeTimeout int) string {
	return fmt.Sprintf(`
  mysql_settings {
    after_connect_script    = "SELECT NOW();"
    authentication_method   = "password"
    execute_timeout         = %[1]d
    max_file_size           = 1024
    parallel_load_threads   = 1
    target_db_type          = "multiple-databases"
  }
`, executeTimeout)
}

func testAccEndpointConfig_mySQLTargetSettings(rName string, outputMySQLSettings bool, executeTimeout int) string {
	mysqlSettings := ""
	if outputMySQLSettings {
		mysqlSettings = testAccEndpointConfig_mySQLTargetSettingsBlock(executeTimeout)
	}
	return acctest.ConfigCompose(testAccEndpointConfig_certificateBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  certificate_arn             = aws_dms_certificate.dms_certificate.certificate_arn
  endpoint_id                 = %[1]q
  endpoint_type               = "target"
  engine_name                 = "mysql"
  server_name                 = "tftest"
  port                        = 5432
  username                    = "tftest"
  password                    = "tftest"
  ssl_mode                    = "verify-full"
  extra_connection_attributes = ""

  %[2]s
}
`, rName, mysqlSettings))
}

func testAccEndpointConfig_postgreSQLSourceSettings(rName string) string {
	return fmt.Sprintf(`

data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dms.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "postgres"
  server_name                 = "tftest"
  port                        = 5432
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "require"
  extra_connection_attributes = ""

  postgres_settings {
    after_connect_script         = "SET search_path TO pg_catalog,public;"
    authentication_method        = "iam"
    capture_ddls                 = true
    ddl_artifacts_schema         = true
    execute_timeout              = 100
    fail_tasks_on_lob_truncation = false
    heartbeat_enable             = true
    heartbeat_frequency          = 5
    heartbeat_schema             = "test"
    map_boolean_as_boolean       = true
    map_jsonb_as_clob            = true
    map_long_varchar_as          = "wstring"
    max_file_size                = 1024
    plugin_name                  = "pglogical"
    service_access_role_arn      = aws_iam_role.test.arn
    slot_name                    = "test"
  }
}
`, rName)
}

func testAccEndpointConfig_postgreSQLTargetSettings(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "target"
  engine_name                 = "postgres"
  server_name                 = "tftest"
  port                        = 5432
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "require"
  extra_connection_attributes = ""

  postgres_settings {
    after_connect_script    = "SET search_path TO pg_catalog,public;"
    authentication_method   = "password"
    babelfish_database_name = "babelfish"
    database_mode           = "babelfish"
    execute_timeout         = 100
    max_file_size           = 1024
  }
}
`, rName)
}

func testAccEndpointConfig_sqlServer(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "sqlserver"
  server_name                 = "tftest"
  port                        = 27017
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName)
}

func testAccEndpointConfig_sqlServerUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "sqlserver"
  server_name                 = "tftest-new-server_name"
  port                        = 27018
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "require"
  extra_connection_attributes = "TlogAccessMode=PreferTlog;"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}
`, rName)
}

func testAccEndpointConfig_sqlServerSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "source"
  engine_name                     = "sqlserver"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName))
}

func testAccEndpointConfig_sybase(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "source"
  engine_name   = "sybase"
  server_name   = "tftest"
  port          = 27017
  username      = "tftest"
  password      = "tftest"
  database_name = "tftest"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName)
}

func testAccEndpointConfig_sybaseUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "source"
  engine_name   = "sybase"
  server_name   = "tftest-new-server_name"
  port          = 27018
  username      = "tftest-new-username"
  password      = "tftest-new-password"
  database_name = "tftest-new-database_name"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}
`, rName)
}

func testAccEndpointConfig_sybaseSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "source"
  engine_name                     = "sybase"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName))
}

func testAccEndpointConfig_docDB(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name               = "tf-test-dms-db"
  endpoint_id                 = %[1]q
  endpoint_type               = "target"
  engine_name                 = "docdb"
  extra_connection_attributes = ""
  password                    = "tftest"
  port                        = 27017
  server_name                 = "tftest"
  ssl_mode                    = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, rName)
}

func testAccEndpointConfig_docDBUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = %[1]q
  endpoint_type               = "target"
  engine_name                 = "docdb"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 27019
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, rName)
}

func testAccEndpointConfig_db2(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name               = "tf-test-dms-db"
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "db2"
  extra_connection_attributes = ""
  password                    = "tftest"
  port                        = 27017
  server_name                 = "tftest"
  ssl_mode                    = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, rName)
}

func testAccEndpointConfig_db2SecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "source"
  engine_name                     = "db2"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName))
}

func testAccEndpointConfig_db2Update(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "db2"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 27019
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, rName)
}

func testAccEndpointConfig_db2zOS(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name               = "tf-test-dms-db"
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "db2-zos"
  extra_connection_attributes = ""
  password                    = "tftest"
  port                        = 27017
  server_name                 = "tftest"
  ssl_mode                    = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, rName)
}

func testAccEndpointConfig_db2zOSSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "source"
  engine_name                     = "db2-zos"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id

  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName))
}

func testAccEndpointConfig_db2zOSUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "db2-zos"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 27019
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, rName)
}

func testAccEndpointConfig_azureSQLManagedInstance(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name               = "tf-test-dms-db"
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "azure-sql-managed-instance"
  extra_connection_attributes = ""
  password                    = "tftest"
  port                        = 3342
  server_name                 = "tftest"
  ssl_mode                    = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, rName)
}

func testAccEndpointConfig_azureSQLManagedInstanceUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "azure-sql-managed-instance"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 3342
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, rName)
}

func testAccEndpointConfig_postgresKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "postgres"
  server_name                 = "tftest"
  port                        = 27018
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "require"
  extra_connection_attributes = ""
  kms_key_arn                 = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccEndpointConfig_sqlserverKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "sqlserver"
  server_name                 = "tftest"
  port                        = 27018
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "require"
  extra_connection_attributes = ""
  kms_key_arn                 = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccEndpointConfig_sybaseKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "source"
  engine_name                 = "sybase"
  server_name                 = "tftest"
  port                        = 27018
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""
  kms_key_arn                 = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccEndpointConfig_redis(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "redis"

  redis_settings {
    auth_type             = "none"
    port                  = 6379
    server_name           = "redis1.test"
    ssl_security_protocol = "plaintext"
  }
}
`, rName)
}

func testAccEndpointConfig_redisUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "redis"

  redis_settings {
    auth_password  = "avoid-plaintext-passwords"
    auth_type      = "auth-role"
    auth_user_name = "tfacctest"
    port           = 6379
    server_name    = "redis2.test"
  }
}
`, rName)
}

func testAccEndpointConfig_redshiftBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier = %[1]q
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "Mustbe8characters"
  node_type          = "ra3.large"
  cluster_type       = "single-node"
  encrypted          = true

  skip_final_snapshot = true
}

data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {"Service": "dms.${data.aws_partition.current.dns_suffix}"},
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
    ]
    resources = ["*"]
  }
}
`, rName)
}

func testAccEndpointConfig_redshift(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_redshiftBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "redshift"
  server_name   = aws_redshift_cluster.test.dns_name
  port          = 27017
  username      = "tftest"
  password      = "tftest"
  database_name = "tftest"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName))
}

func testAccEndpointConfig_redshiftSecretID(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_secretBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                     = %[1]q
  endpoint_type                   = "target"
  engine_name                     = "redshift"
  secrets_manager_access_role_arn = aws_iam_role.test.arn
  secrets_manager_arn             = aws_secretsmanager_secret.test.id
  database_name                   = "tftest"
  ssl_mode                        = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}
`, rName))
}

func testAccEndpointConfig_redshiftUpdate(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_redshiftBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id                 = %[1]q
  endpoint_type               = "target"
  engine_name                 = "redshift"
  server_name                 = aws_redshift_cluster.test.dns_name
  port                        = 27018
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  extra_connection_attributes = "acceptanydate=true"

  redshift_settings {
    service_access_role_arn = aws_iam_role.test.arn
    bucket_name             = "bucket_name"
    bucket_folder           = "bucket_folder"
    encryption_mode         = "SSE_S3"
  }

  tags = {
    Name   = %[1]q
    Update = "updated"
    Add    = "added"
  }
}
`, rName))
}

func testAccEndpointConfig_redshiftKMSKey(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_redshiftBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "redshift"
  server_name   = aws_redshift_cluster.test.dns_name
  port          = 27017
  username      = "tftest"
  password      = "tftest"
  database_name = "tftest"
  ssl_mode      = "none"
  kms_key_arn   = aws_kms_key.test.arn

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`, rName))
}

func testAccEndpointConfig_redshiftConnSSEKMSKeyARN(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_redshiftBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "redshift"
  server_name   = aws_redshift_cluster.test.dns_name
  port          = 27017
  username      = "tftest"
  password      = "tftest"
  database_name = "tftest"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  redshift_settings {
    encryption_mode                   = "SSE_KMS"
    server_side_encryption_kms_key_id = aws_kms_key.test.arn
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`, rName))
}

func testAccEndpointConfig_redshiftConnSSEKMSKeyId(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_redshiftBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "redshift"
  server_name   = aws_redshift_cluster.test.dns_name
  port          = 27017
  username      = "tftest"
  password      = "tftest"
  database_name = "tftest"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  redshift_settings {
    encryption_mode                   = "SSE_KMS"
    server_side_encryption_kms_key_id = aws_kms_key.test.key_id
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`, rName))
}

func testAccEndpointConfig_pauseReplicationTasks(rName string, pause bool) string {
	return acctest.ConfigCompose(testAccEndpointConfig_rdsClusterBase(rName), fmt.Sprintf(`
resource "aws_dms_endpoint" "source" {
  database_name           = "tftest"
  endpoint_id             = "%[1]s-source"
  endpoint_type           = "source"
  engine_name             = "aurora"
  password                = "mustbeeightcharaters"
  pause_replication_tasks = %[2]t
  port                    = 3306
  server_name             = aws_rds_cluster.source.endpoint
  username                = "tftest"
}

resource "aws_dms_endpoint" "target" {
  database_name           = "tftest"
  endpoint_id             = "%[1]s-target"
  endpoint_type           = "target"
  engine_name             = "aurora"
  password                = "mustbeeightcharaters"
  pause_replication_tasks = %[2]t
  port                    = 3306
  server_name             = aws_rds_cluster.target.endpoint
  username                = "tftest"
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_replication_instance" "test" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.c5.large"
  replication_instance_id      = %[1]q
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
  vpc_security_group_ids       = [aws_security_group.test.id]
}

resource "aws_dms_replication_task" "test" {
  migration_type           = "full-load-and-cdc"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  replication_task_id      = %[1]q
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "testrule",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )

  start_replication_task = true

  tags = {
    Name = %[1]q
  }

  target_endpoint_arn = aws_dms_endpoint.target.endpoint_arn

  depends_on = [aws_rds_cluster_instance.source, aws_rds_cluster_instance.target]
}
`, rName, pause))
}

// testAccEndpointConfig_rdsClusterBase configures a pair of Aurora RDS clusters (and instances) ready for replication.
func testAccEndpointConfig_rdsClusterBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = -1
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_rds_engine_version" "default" {
  engine = "aurora-mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.t3.small", "db.t3.medium", "db.t3.large"]
}

resource "aws_rds_cluster_parameter_group" "test" {
  name        = "%[1]s-pg-cluster"
  family      = data.aws_rds_engine_version.default.parameter_group_family
  description = "DMS cluster parameter group"

  parameter {
    name         = "binlog_format"
    value        = "ROW"
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "binlog_row_image"
    value        = "Full"
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "binlog_checksum"
    value        = "NONE"
    apply_method = "pending-reboot"
  }
}

resource "aws_rds_cluster" "source" {
  cluster_identifier              = "%[1]s-aurora-cluster-source"
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  engine_version                  = data.aws_rds_orderable_db_instance.test.engine_version
  database_name                   = "tftest"
  master_username                 = "tftest"
  master_password                 = "mustbeeightcharaters"
  skip_final_snapshot             = true
  vpc_security_group_ids          = [aws_security_group.test.id]
  db_subnet_group_name            = aws_db_subnet_group.test.name
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.test.name
}

resource "aws_rds_cluster_instance" "source" {
  identifier           = "%[1]s-source-primary"
  cluster_identifier   = aws_rds_cluster.source.id
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_subnet_group_name = aws_db_subnet_group.test.name
}

resource "aws_rds_cluster" "target" {
  cluster_identifier     = "%[1]s-aurora-cluster-target"
  engine                 = data.aws_rds_orderable_db_instance.test.engine
  engine_version         = data.aws_rds_orderable_db_instance.test.engine_version
  database_name          = "tftest"
  master_username        = "tftest"
  master_password        = "mustbeeightcharaters"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
  db_subnet_group_name   = aws_db_subnet_group.test.name
}

resource "aws_rds_cluster_instance" "target" {
  identifier           = "%[1]s-target-primary"
  cluster_identifier   = aws_rds_cluster.target.id
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_subnet_group_name = aws_db_subnet_group.test.name
}
`, rName))
}
