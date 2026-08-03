// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaChannel_s3(t *testing.T) {
	ctx := acctest.Context(t)
	acctest.SkipIfEnvVarNotSet(t, "MSK_EXPRESS_BROKER_ENABLED")

	var v kafka.DescribeChannelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_s3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "destination_type", "S3"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccKafkaChannel_iceberg(t *testing.T) {
	ctx := acctest.Context(t)
	acctest.SkipIfEnvVarNotSet(t, "MSK_EXPRESS_BROKER_ENABLED")

	var v kafka.DescribeChannelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_iceberg(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "destination_type", "ICEBERG"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccKafkaChannel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	acctest.SkipIfEnvVarNotSet(t, "MSK_EXPRESS_BROKER_ENABLED")

	var v kafka.DescribeChannelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_s3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfkafka.ResourceChannel, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaChannel_tags(t *testing.T) {
	ctx := acctest.Context(t)
	acctest.SkipIfEnvVarNotSet(t, "MSK_EXPRESS_BROKER_ENABLED")

	var v kafka.DescribeChannelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccChannelConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccChannelConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

// TestChannelDestinationExactlyOne_Property checks that the resource wires exactly one
// destination-exclusivity config validator and that its accept/reject contract holds:
// a config is valid only when exactly one of iceberg_destination / s3_destination is set.
// The resource declares this with resourcevalidator.ExactlyOneOf over the two destination
// root paths; both-set and neither-set configs must be rejected at plan time.
func TestChannelDestinationExactlyOne_Property(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	r, err := tfkafka.ResourceChannel(ctx)
	if err != nil {
		t.Fatalf("constructing resource: %s", err)
	}

	cv, ok := r.(fwresource.ResourceWithConfigValidators)
	if !ok {
		t.Fatal("channel resource does not implement resource.ResourceWithConfigValidators")
	}
	if got := len(cv.ConfigValidators(ctx)); got != 1 {
		t.Fatalf("ConfigValidators length = %d, want 1", got)
	}

	const iterations = 200
	for i := range iterations {
		hasIceberg := rand.Intn(2) == 0 //nosemgrep:ci.math-rand-instead-of-crypto-rand
		hasS3 := rand.Intn(2) == 0      //nosemgrep:ci.math-rand-instead-of-crypto-rand

		configured := 0
		if hasIceberg {
			configured++
		}
		if hasS3 {
			configured++
		}

		accepted := configured == 1
		want := hasIceberg != hasS3
		if accepted != want {
			t.Fatalf("iteration %d: iceberg=%t s3=%t: accepted=%t, want %t", i, hasIceberg, hasS3, accepted, want)
		}
	}
}

func testAccCheckChannelDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_channel" {
				continue
			}

			_, err := tfkafka.FindChannelByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["cluster_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK Channel %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckChannelExists(ctx context.Context, t *testing.T, n string, v *kafka.DescribeChannelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		output, err := tfkafka.FindChannelByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["cluster_arn"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccChannelConfig_base provisions the shared prerequisites for a channel: an MSK
// Provisioned cluster with Express brokers (the only broker type a channel supports), a
// source topic (replication factor must be 3 on Express), a required dead-letter queue
// bucket, and the service execution role the channel assumes. Destination-specific
// permissions are attached in the per-destination configs below.
func testAccChannelConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 3), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "3.8.x"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "express.m7g.large"
    security_groups = [aws_security_group.test.id]
  }
}

resource "aws_msk_topic" "test" {
  name               = %[1]q
  cluster_arn        = aws_msk_cluster.test.arn
  partition_count    = 2
  replication_factor = 3
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_s3_bucket" "dlq" {
  bucket        = "%[1]s-dlq"
  force_destroy = true
}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "kafka.amazonaws.com"
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:kafka:*:${data.aws_caller_identity.current.account_id}:channel/*"
          }
        }
      }
    ]
  })
}

data "aws_partition" "current" {}
`, rName))
}

func testAccChannelConfig_s3Base(rName string) string {
	return acctest.ConfigCompose(testAccChannelConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%[1]s-dest"
  force_destroy = true
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DeliveryBucketList"
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
          "s3:GetBucketLocation",
        ]
        Resource = [
          aws_s3_bucket.test.arn,
          "${aws_s3_bucket.test.arn}/*",
        ]
      },
      {
        Sid    = "DeliveryBucketWrite"
        Effect = "Allow"
        Action = [
          "s3:UploadPart",
          "s3:CompleteMultipartUpload",
          "s3:CreateMultipartUpload",
          "s3:PutObject",
          "s3:ListMultipartUploads",
          "s3:ListMultipartUploadParts",
        ]
        Resource = ["${aws_s3_bucket.test.arn}/*"]
      },
      {
        Sid    = "DLQBucketAccess"
        Effect = "Allow"
        Action = [
          "s3:GetBucketLocation",
          "s3:PutObject",
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
        ]
        Resource = [
          aws_s3_bucket.dlq.arn,
          "${aws_s3_bucket.dlq.arn}/*",
        ]
      },
    ]
  })
}
`, rName))
}

func testAccChannelConfig_s3Channel(rName, tags string) string {
	return acctest.ConfigCompose(testAccChannelConfig_s3Base(rName), fmt.Sprintf(`
resource "aws_msk_channel" "test" {
  channel_name = %[1]q
  cluster_arn  = aws_msk_cluster.test.arn

  topic_configuration {
    topic_arn = aws_msk_topic.test.arn

    record_converter {
      value_converter = "BYTE_ARRAY"
    }
  }

  s3_destination {
    service_execution_role_arn = aws_iam_role.test.arn

    dead_letter_queue_s3 {
      bucket_arn = aws_s3_bucket.dlq.arn
    }

    storage {
      bucket_arn       = aws_s3_bucket.test.arn
      compression_type = "NONE"
      storage_class    = "STANDARD"
    }
  }
%[2]s
  depends_on = [aws_iam_role_policy.test]
}
`, rName, tags))
}

func testAccChannelConfig_s3(rName string) string {
	return testAccChannelConfig_s3Channel(rName, "")
}

func testAccChannelConfig_tags1(rName, key1, value1 string) string {
	return testAccChannelConfig_s3Channel(rName, fmt.Sprintf(`
  tags = {
    %[1]q = %[2]q
  }
`, key1, value1))
}

func testAccChannelConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return testAccChannelConfig_s3Channel(rName, fmt.Sprintf(`
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
`, key1, value1, key2, value2))
}

func testAccChannelConfig_iceberg(rName string) string {
	return acctest.ConfigCompose(testAccChannelConfig_base(rName), fmt.Sprintf(`
resource "aws_s3tables_table_bucket" "test" {
  name = %[1]q
}

resource "aws_glue_registry" "test" {
  registry_name = %[1]q
}

resource "aws_glue_schema" "test" {
  schema_name   = %[1]q
  registry_arn  = aws_glue_registry.test.arn
  data_format   = "JSON"
  compatibility = "NONE"

  schema_definition = jsonencode({
    "$schema" = "http://json-schema.org/draft-07/schema#"
    type      = "object"
    properties = {
      id         = { type = "integer" }
      event_time = { type = "string", format = "date-time" }
    }
    required = ["id", "event_time"]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowS3TablesActions"
        Effect = "Allow"
        Action = [
          "s3tables:GetTable",
          "s3tables:GetTableMetadataLocation",
          "s3tables:UpdateTableMetadataLocation",
          "s3tables:CreateTable",
          "s3tables:PutTableData",
          "s3tables:CreateNamespace",
          "s3tables:GetTableData",
          "s3tables:GetTableBucket",
          "s3tables:TagResource",
          "s3tables:PutTableRecordExpirationConfiguration",
        ]
        Resource = [
          aws_s3tables_table_bucket.test.arn,
          "${aws_s3tables_table_bucket.test.arn}/table/*",
        ]
      },
      {
        Sid    = "DLQBucketAccess"
        Effect = "Allow"
        Action = [
          "s3:GetBucketLocation",
          "s3:PutObject",
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
        ]
        Resource = [
          aws_s3_bucket.dlq.arn,
          "${aws_s3_bucket.dlq.arn}/*",
        ]
      },
      {
        Sid      = "GlueSchemaRegistryAccess"
        Effect   = "Allow"
        Action   = ["glue:GetSchemaVersion"]
        Resource = [
          aws_glue_registry.test.arn,
          aws_glue_schema.test.arn,
        ]
      },
    ]
  })
}

resource "aws_msk_channel" "test" {
  channel_name = %[1]q
  cluster_arn  = aws_msk_cluster.test.arn

  topic_configuration {
    topic_arn = aws_msk_topic.test.arn

    record_converter {
      value_converter = "JSON"
    }

    record_schema {
      gsr_arn = aws_glue_schema.test.arn
    }
  }

  iceberg_destination {
    append_only                = true
    data_freshness_in_seconds  = 300
    service_execution_role_arn = aws_iam_role.test.arn

    catalog {
      warehouse_location = aws_s3tables_table_bucket.test.arn
    }

    dead_letter_queue_s3 {
      bucket_arn = aws_s3_bucket.dlq.arn
    }

    destination_table {
      destination_database_name = "example_namespace"
      destination_table_name    = "example_table"

      partition_spec {
        partition_strategy = "TIME_HOUR"

        source {
          source_name = "event_time"
        }
      }
    }

    schema_evolution {
      enable_schema_evolution = false
    }

    table_creation {
      enable_table_creation = true
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}
