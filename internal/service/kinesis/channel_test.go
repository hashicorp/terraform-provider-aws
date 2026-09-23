// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkinesis "github.com/hashicorp/terraform-provider-aws/internal/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kinesis)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "channel_arn", "kinesis", regexache.MustCompile(`channel/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_name", rName),
					resource.TestCheckResourceAttr(resourceName, "channel_status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceName, "channel_creation_timestamp"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role_arn", "aws_iam_role.role", "arn"),
					resource.TestCheckResourceAttr(resourceName, "stream_configuration_list.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "stream_configuration_list.0.stream_arn", "aws_kinesis_stream.stream", "arn"),
					resource.TestCheckResourceAttr(resourceName, "stream_configuration_list.0.record_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stream_configuration_list.0.record_configuration.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "s3_destination_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_destination_configuration.0.data_freshness_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "s3_destination_configuration.0.storage_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_destination_configuration.0.storage_configuration.0.bucket_arn", "aws_s3_bucket.bucket", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_destination_configuration.0.storage_configuration.0.expected_bucket_owner", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "s3_destination_configuration.0.storage_configuration.0.compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "s3_destination_configuration.0.storage_configuration.0.storage_class", "STANDARD"),
					resource.TestCheckResourceAttrSet(resourceName, "s3_destination_configuration.0.storage_configuration.0.output_key_template"),
					resource.TestCheckResourceAttr(resourceName, "s3_destination_configuration.0.dead_letter_queue_s3_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "s3_destination_configuration.0.dead_letter_queue_s3_configuration.0.bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "channel_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "channel_arn",
			},
		},
	})
}

func TestAccKinesisChannel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kinesis)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfkinesis.ResourceChannel, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccKinesisChannel_updateLogging(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kinesis)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "channel_arn", "kinesis", regexache.MustCompile(`channel/.+$`)),
				),
			},
			{
				Config: testAccChannelConfig_updateLogging(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "channel_arn", "kinesis", regexache.MustCompile(`channel/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.cloudwatch_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.cloudwatch_logs.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "logging_configuration.0.cloudwatch_logs.0.log_group_name", "aws_cloudwatch_log_group.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.cloudwatch_logs.0.log_stream_name", rName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "channel_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "channel_arn",
			},
		},
	})
}

func TestAccKinesisChannel_updateFreshness(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kinesis)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "channel_arn", "kinesis", regexache.MustCompile(`channel/.+$`)),
				),
			},
			{
				Config: testAccChannelConfig_updateFreshness(rName, 420),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "channel_arn", "kinesis", regexache.MustCompile(`channel/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "s3_destination_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_destination_configuration.0.data_freshness_in_seconds", "420"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "channel_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "channel_arn",
			},
		},
	})
}

func TestAccKinesisChannel_streamingTable(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kinesis_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kinesis)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_streamingTable(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "channel_arn", "kinesis", regexache.MustCompile(`channel/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_name", rName),
					resource.TestCheckResourceAttr(resourceName, "channel_status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceName, "channel_creation_timestamp"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role_arn", "aws_iam_role.role", "arn"),
					resource.TestCheckResourceAttr(resourceName, "stream_configuration_list.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "stream_configuration_list.0.stream_arn", "aws_kinesis_stream.stream", "arn"),
					resource.TestCheckResourceAttr(resourceName, "stream_configuration_list.0.record_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stream_configuration_list.0.record_configuration.0.record_format_type", "GSR_JSON"),
					resource.TestCheckResourceAttrPair(resourceName, "stream_configuration_list.0.record_configuration.0.gsr_schema_arn", "aws_glue_schema.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "s3_tables_destination_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_tables_destination_configuration.0.dead_letter_queue_s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_tables_destination_configuration.0.dead_letter_queue_s3_configuration.0.bucket_arn", "aws_s3_bucket.bucket", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_tables_destination_configuration.0.dead_letter_queue_s3_configuration.0.expected_bucket_owner", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "s3_tables_destination_configuration.0.dead_letter_queue_s3_configuration.0.error_output_prefix", "errors/"),
					resource.TestCheckResourceAttr(resourceName, "s3_tables_destination_configuration.0.s3_tables_configuration_list.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_tables_destination_configuration.0.s3_tables_configuration_list.0.table_bucket_arn", "aws_s3tables_table_bucket.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_tables_destination_configuration.0.s3_tables_configuration_list.0.namespace", "aws_s3tables_namespace.test", "namespace"),
					resource.TestCheckResourceAttr(resourceName, "s3_tables_destination_configuration.0.s3_tables_configuration_list.0.table_name", "test_table"),
					resource.TestCheckResourceAttr(resourceName, "s3_tables_destination_configuration.0.s3_tables_configuration_list.0.compression_type", "ZSTD"),
					resource.TestCheckResourceAttr(resourceName, "s3_tables_destination_configuration.0.s3_tables_configuration_list.0.partition_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_tables_destination_configuration.0.s3_tables_configuration_list.0.partition_spec.0.partition_fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_tables_destination_configuration.0.s3_tables_configuration_list.0.partition_spec.0.partition_fields.0.source_name", "event_time"),
					resource.TestCheckResourceAttr(resourceName, "s3_tables_destination_configuration.0.s3_tables_configuration_list.0.partition_spec.0.partition_fields.0.transform", "TIME_HOUR"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "channel_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "channel_arn",
			},
		},
	})
}

func testAccCheckChannelDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KinesisClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_channel" {
				continue
			}

			arn := rs.Primary.Attributes["channel_arn"]
			_, err := tfkinesis.FindChannelByArn(ctx, conn, arn)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.Kinesis, create.ErrActionCheckingDestroyed, tfkinesis.ResNameChannel, arn, err)
			}

			return create.Error(names.Kinesis, create.ErrActionCheckingDestroyed, tfkinesis.ResNameChannel, arn, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckChannelExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Kinesis, create.ErrActionCheckingExistence, tfkinesis.ResNameChannel, name, errors.New("not found"))
		}

		arn := rs.Primary.Attributes["channel_arn"]
		if arn == "" {
			return create.Error(names.Kinesis, create.ErrActionCheckingExistence, tfkinesis.ResNameChannel, name, errors.New("channel_arn not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).KinesisClient(ctx)

		_, err := tfkinesis.FindChannelByArn(ctx, conn, arn)
		if err != nil {
			return create.Error(names.Kinesis, create.ErrActionCheckingExistence, tfkinesis.ResNameChannel, arn, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).KinesisClient(ctx)

	input := &kinesis.ListChannelsInput{}

	_, err := conn.ListChannels(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccChannelConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = [
            "kinesis.amazonaws.com",
            "s3.amazonaws.com",
			"glue.amazonaws.com"
          ]
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy" "policy" {
  name = %[1]q
  role = aws_iam_role.role.id

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kinesis:*"
        ]
        Resource = aws_kinesis_stream.stream.arn
      },
	  {
        Effect = "Allow"
        Action = [
          "glue:*"
        ]
        Resource = ["*"]
      },
	  {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogStreams"
        ]
        Resource = "${aws_cloudwatch_log_group.test.arn}:*"
      },
      {
        Effect = "Allow"
        Action = ["s3:*"]
        Resource = [
          "${aws_s3_bucket.bucket.arn}",
          "${aws_s3_bucket.bucket.arn}/*"
        ]
      },
	  {
        Effect = "Allow"
        Action = ["s3tables:*"]
        Resource = [
          "*"
        ]
      }
    ]
  })
}

resource "aws_s3tables_namespace" "test" {
  namespace        = "test_namespace"
  table_bucket_arn = aws_s3tables_table_bucket.test.arn
}

resource "aws_s3tables_table_bucket" "test" {
  name = %[1]q
}

resource "aws_glue_registry" "test" {
  registry_name = %[1]q
}

resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "JSON"
  compatibility     = "NONE"
  schema_definition = jsonencode({
    "$id": "https://example.com/record.schema.json",
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "Record",
    "type": "object",
    "properties": {
      "event_time": {
        "type": "string",
        "format": "date-time"
      },
      "data": {
        "type": "string"
      }
    },
    "required": ["event_time"]
  })
}

resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_kinesis_stream" "stream" {
  name                      = %[1]q
  encryption_type           = "NONE"
  enforce_consumer_deletion = false
  max_record_size_in_kib    = 1024
  retention_period          = 24
  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/kinesis/%[1]s"
}

data "aws_caller_identity" "current" {}
`, rName)
}

func testAccChannelConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccChannelConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_channel" "test" {
  channel_name               = %[1]q
  service_execution_role_arn = aws_iam_role.role.arn

  stream_configuration_list {
    stream_arn = aws_kinesis_stream.stream.arn

    record_configuration {
      record_format_type = "JSON"
    }
  }

  s3_destination_configuration {
    storage_configuration {
      bucket_arn            = aws_s3_bucket.bucket.arn
      expected_bucket_owner = data.aws_caller_identity.current.account_id
      compression_type      = "NONE"
    }
  }

  depends_on = [
    aws_iam_role_policy.policy,
  ]
}
`, rName))
}

func testAccChannelConfig_updateLogging(rName string) string {
	return acctest.ConfigCompose(testAccChannelConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_channel" "test" {
  channel_name               = %[1]q
  service_execution_role_arn = aws_iam_role.role.arn

  stream_configuration_list {
    stream_arn = aws_kinesis_stream.stream.arn

    record_configuration {
      record_format_type = "JSON"
    }
  }

  s3_destination_configuration {
    storage_configuration {
      bucket_arn            = aws_s3_bucket.bucket.arn
      expected_bucket_owner = data.aws_caller_identity.current.account_id
      compression_type      = "NONE"
    }
  }

  logging_configuration {
    cloudwatch_logs {
      enabled 		  = true
	  log_group_name  = aws_cloudwatch_log_group.test.name
	  log_stream_name = %[1]q
    }
  }

  depends_on = [
    aws_iam_role_policy.policy,
  ]
}
`, rName))
}

func testAccChannelConfig_updateFreshness(rName string, freshness int) string {
	return acctest.ConfigCompose(testAccChannelConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_channel" "test" {
  channel_name               = %[1]q
  service_execution_role_arn = aws_iam_role.role.arn

  stream_configuration_list {
    stream_arn = aws_kinesis_stream.stream.arn

    record_configuration {
      record_format_type = "JSON"
    }
  }

  s3_destination_configuration {
    data_freshness_in_seconds = %[2]d
    storage_configuration {
      bucket_arn            = aws_s3_bucket.bucket.arn
      expected_bucket_owner = data.aws_caller_identity.current.account_id
      compression_type      = "NONE"
    }
  }

  depends_on = [
    aws_iam_role_policy.policy,
  ]
}
`, rName, freshness))
}

func testAccChannelConfig_streamingTable(rName string) string {
	return acctest.ConfigCompose(testAccChannelConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_channel" "test" {
  channel_name               = %[1]q
  service_execution_role_arn = aws_iam_role.role.arn

  stream_configuration_list {
    stream_arn = aws_kinesis_stream.stream.arn

    record_configuration {
      record_format_type = "GSR_JSON"
      gsr_schema_arn     = aws_glue_schema.test.arn
    }
  }

  s3_tables_destination_configuration {
    dead_letter_queue_s3_configuration {
      bucket_arn            = aws_s3_bucket.bucket.arn
      expected_bucket_owner = data.aws_caller_identity.current.account_id
      error_output_prefix   = "errors/"
    }

    s3_tables_configuration_list {
      table_bucket_arn = aws_s3tables_table_bucket.test.arn
      namespace        = aws_s3tables_namespace.test.namespace
      table_name       = "test_table"
      compression_type = "ZSTD"
	  partition_spec {
		partition_fields {
		  source_name = "event_time"
		  transform   = "TIME_HOUR" 
		}
	  }
    }
  }

  depends_on = [
    aws_iam_role_policy.policy,
  ]
}
	`, rName))
}
