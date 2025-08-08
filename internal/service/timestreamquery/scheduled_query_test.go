// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamquery_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamquery"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreamquery/types"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	awswritetypes "github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftimestreamquery "github.com/hashicorp/terraform-provider-aws/internal/service/timestreamquery"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamQueryScheduledQuery_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var scheduledquery awstypes.ScheduledQueryDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamquery_scheduled_query.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamQueryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledQueryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Must be done in 2 steps because the scheduled query requires data to be ingested first
				// which creates the columns. Otherwise, the SQL will always be invalid because no columns exist.
				Config: testAccScheduledQueryConfig_base(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccWriteRecords(ctx, "aws_timestreamwrite_table.test", rName, rName),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccScheduledQueryConfig_base(rName),
					testAccScheduledQueryConfig_basic(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledQueryExists(ctx, resourceName, &scheduledquery),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "timestream", regexache.MustCompile(`scheduled-query/.+$`)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, "error_report_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_report_configuration.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_report_configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_report_configuration.0.s3_configuration.0.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "error_report_configuration.0.s3_configuration.0.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "error_report_configuration.0.s3_configuration.0.encryption_option", "SSE_S3"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "last_run_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "next_invocation_time"),
					resource.TestCheckResourceAttr(resourceName, "notification_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_configuration.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_configuration.0.sns_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_configuration.0.sns_configuration.0.%", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "notification_configuration.0.sns_configuration.0.topic_arn", "aws_sns_topic.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "query_string", fmt.Sprintf("SELECT region, az, hostname, BIN(time, 15s) AS binned_timestamp,\n\tROUND(AVG(cpu_utilization), 2) AS avg_cpu_utilization, \n\tROUND(APPROX_PERCENTILE(cpu_utilization, 0.9), 2) AS p90_cpu_utilization, \n\tROUND(APPROX_PERCENTILE(cpu_utilization, 0.95), 2) AS p95_cpu_utilization, \n\tROUND(APPROX_PERCENTILE(cpu_utilization, 0.99), 2) AS p99_cpu_utilization \nFROM \"%[1]s\".\"%[1]s\"\nWHERE measure_name = 'metrics' AND time > ago(2h) \nGROUP BY region, hostname, az, BIN(time, 15s) \nORDER BY binned_timestamp ASC \nLIMIT 5\n", rName)),
					resource.TestCheckResourceAttr(resourceName, "recently_failed_runs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule_configuration.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule_configuration.0.schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.%", "7"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.database_name", fmt.Sprintf("%s-results", rName)),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.dimension_mapping.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.dimension_mapping.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.dimension_mapping.0.dimension_value_type", "VARCHAR"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.dimension_mapping.0.name", "az"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.dimension_mapping.1.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.dimension_mapping.1.dimension_value_type", "VARCHAR"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.dimension_mapping.1.name", names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.dimension_mapping.2.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.dimension_mapping.2.dimension_value_type", "VARCHAR"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.dimension_mapping.2.name", "hostname"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.mixed_measure_mapping.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.0.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.0.measure_value_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.0.source_column", "avg_cpu_utilization"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.1.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.1.measure_value_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.1.source_column", "p90_cpu_utilization"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.2.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.2.measure_value_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.2.source_column", "p95_cpu_utilization"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.3.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.3.measure_value_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.multi_measure_attribute_mapping.3.source_column", "p99_cpu_utilization"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.multi_measure_mappings.0.target_multi_measure_name", "multi-metrics"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.table_name", rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.timestream_configuration.0.time_column", "binned_timestamp"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccScheduledQueryImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccTimestreamQueryScheduledQuery_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var scheduledquery awstypes.ScheduledQueryDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamquery_scheduled_query.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamQueryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledQueryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Must be done in 2 steps because the scheduled query requires data to be ingested first
				// which creates the columns. Otherwise, the SQL will always be invalid because no columns exist.
				Config: testAccScheduledQueryConfig_base(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccWriteRecords(ctx, "aws_timestreamwrite_table.test", rName, rName),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccScheduledQueryConfig_base(rName),
					testAccScheduledQueryConfig_basic(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledQueryExists(ctx, resourceName, &scheduledquery),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tftimestreamquery.ResourceScheduledQuery, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccScheduledQueryImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrARN], nil
	}
}

func testAccWriteRecords(ctx context.Context, name, database, table string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.TimestreamQuery, create.ErrActionCheckingExistence, tftimestreamquery.ResNameScheduledQuery, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrARN] == "" {
			return create.Error(names.TimestreamQuery, create.ErrActionCheckingExistence, tftimestreamquery.ResNameScheduledQuery, name, errors.New("not set"))
		}

		// From an AWS example: https://github.com/awslabs/amazon-timestream-tools/blob/mainline/sample_apps/goV2/utils/timestream-helper.go
		//
		// 0      1         2  3          4        5          6                             7            8               9                 10     11                 12                13
		// region,us-east-2,az,us-east-2a,hostname,host-WgAuL,2020-03-18 02:56:02.342000000,MILLISECONDS,cpu_utilization,59.16598729806647,DOUBLE,memory_utilization,57.18926269056821,DOUBLE
		// region,us-east-1,az,us-east-1b,hostname,host-EOh5a,2020-03-18 01:28:46.608000000,MILLISECONDS,memory_utilization,66.99426972454896,DOUBLE,cpu_utilization,90.75165411419735,DOUBLE

		dimensions := []awswritetypes.Dimension{{
			Name:  aws.String(names.AttrRegion),
			Value: aws.String("us-east-2"), //lintignore:AWSAT003
		}, {
			Name:  aws.String("az"),
			Value: aws.String("us-east-2a"), //lintignore:AWSAT003
		}, {
			Name:  aws.String("hostname"),
			Value: aws.String("host-WgAuL"),
		}}

		multiMeasures := []awswritetypes.MeasureValue{{
			Name:  aws.String("cpu_utilization"),
			Value: aws.String("59.16598729806647"),
			Type:  awswritetypes.MeasureValueTypeDouble,
		}, {
			Name:  aws.String("memory_utilization"),
			Value: aws.String("57.18926269056821"),
			Type:  awswritetypes.MeasureValueTypeDouble,
		}}

		currentTimeInMilliSeconds := time.Now().UnixNano() / int64(time.Millisecond)
		records := make([]awswritetypes.Record, 0)
		records = append(records, awswritetypes.Record{
			Dimensions:       dimensions,
			MeasureName:      aws.String("metrics"),
			MeasureValueType: awswritetypes.MeasureValueTypeMulti,
			MeasureValues:    multiMeasures,
			Time:             aws.String(strconv.FormatInt(currentTimeInMilliSeconds-10*int64(50), 10)),
			TimeUnit:         awswritetypes.TimeUnitMilliseconds,
		})

		input := &timestreamwrite.WriteRecordsInput{
			DatabaseName: aws.String(database),
			TableName:    aws.String(table),
			Records:      records,
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamWriteClient(ctx)
		_, err := conn.WriteRecords(ctx, input)
		if err != nil {
			return create.Error(names.TimestreamQuery, create.ErrActionChecking, tftimestreamquery.ResNameScheduledQuery, rs.Primary.Attributes[names.AttrARN], err)
		}

		return nil
	}
}

func testAccCheckScheduledQueryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamQueryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_timestreamquery_scheduled_query" {
				continue
			}

			_, err := tftimestreamquery.FindScheduledQueryByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.TimestreamQuery, create.ErrActionCheckingDestroyed, tftimestreamquery.ResNameScheduledQuery, rs.Primary.Attributes[names.AttrARN], err)
			}

			return create.Error(names.TimestreamQuery, create.ErrActionCheckingDestroyed, tftimestreamquery.ResNameScheduledQuery, rs.Primary.Attributes[names.AttrARN], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckScheduledQueryExists(ctx context.Context, name string, scheduledquery *awstypes.ScheduledQueryDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.TimestreamQuery, create.ErrActionCheckingExistence, tftimestreamquery.ResNameScheduledQuery, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrARN] == "" {
			return create.Error(names.TimestreamQuery, create.ErrActionCheckingExistence, tftimestreamquery.ResNameScheduledQuery, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamQueryClient(ctx)

		resp, err := tftimestreamquery.FindScheduledQueryByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return create.Error(names.TimestreamQuery, create.ErrActionCheckingExistence, tftimestreamquery.ResNameScheduledQuery, rs.Primary.Attributes[names.AttrARN], err)
		}

		*scheduledquery = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamQueryClient(ctx)

	input := &timestreamquery.ListScheduledQueriesInput{}

	_, err := conn.ListScheduledQueries(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccScheduledQueryConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = ["sqs:SendMessage"]
      Resource = aws_sqs_queue.test.arn
      Condition = {
        ArnEquals = {
          "aws:SourceArn" = aws_sns_topic.test.arn
        }
      }
    }]
  })
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "timestream.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = [
        "kms:Decrypt",
        "sns:Publish",
        "timestream:describeEndpoints",
        "timestream:Select",
        "timestream:SelectValues",
        "timestream:WriteRecords",
        "s3:PutObject",
      ]
      Resource = "*"
      Effect   = "Allow"
    }]
  })
}

resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}

resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  magnetic_store_write_properties {
    enable_magnetic_store_writes = true
  }

  retention_properties {
    magnetic_store_retention_period_in_days = 1
    memory_store_retention_period_in_hours  = 1
  }
}

resource "aws_timestreamwrite_database" "results" {
  database_name = "%[1]s-results"
}

resource "aws_timestreamwrite_table" "results" {
  database_name = aws_timestreamwrite_database.results.database_name
  table_name    = %[1]q

  magnetic_store_write_properties {
    enable_magnetic_store_writes = true
  }

  retention_properties {
    magnetic_store_retention_period_in_days = 1
    memory_store_retention_period_in_hours  = 1
  }
}
`, rName)
}

func testAccScheduledQueryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamquery_scheduled_query" "test" {
  execution_role_arn = aws_iam_role.test.arn
  name               = aws_timestreamwrite_table.test.table_name
  query_string       = <<EOF
SELECT region, az, hostname, BIN(time, 15s) AS binned_timestamp,
	ROUND(AVG(cpu_utilization), 2) AS avg_cpu_utilization, 
	ROUND(APPROX_PERCENTILE(cpu_utilization, 0.9), 2) AS p90_cpu_utilization, 
	ROUND(APPROX_PERCENTILE(cpu_utilization, 0.95), 2) AS p95_cpu_utilization, 
	ROUND(APPROX_PERCENTILE(cpu_utilization, 0.99), 2) AS p99_cpu_utilization 
FROM %[1]q.%[1]q
WHERE measure_name = 'metrics' AND time > ago(2h) 
GROUP BY region, hostname, az, BIN(time, 15s) 
ORDER BY binned_timestamp ASC 
LIMIT 5
EOF

  error_report_configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.bucket
    }
  }

  notification_configuration {
    sns_configuration {
      topic_arn = aws_sns_topic.test.arn
    }
  }

  schedule_configuration {
    schedule_expression = "rate(1 hour)"
  }

  target_configuration {
    timestream_configuration {
      database_name = aws_timestreamwrite_database.results.database_name
      table_name    = aws_timestreamwrite_table.results.table_name
      time_column   = "binned_timestamp"

      dimension_mapping {
        dimension_value_type = "VARCHAR"
        name                 = "az"
      }

      dimension_mapping {
        dimension_value_type = "VARCHAR"
        name                 = "region"
      }

      dimension_mapping {
        dimension_value_type = "VARCHAR"
        name                 = "hostname"
      }

      multi_measure_mappings {
        target_multi_measure_name = "multi-metrics"

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "avg_cpu_utilization"
        }

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "p90_cpu_utilization"
        }

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "p95_cpu_utilization"
        }

        multi_measure_attribute_mapping {
          measure_value_type = "DOUBLE"
          source_column      = "p99_cpu_utilization"
        }
      }
    }
  }

  tags = {
    Name        = %[1]q
    Environment = "test"
  }
}
`, rName)
}
