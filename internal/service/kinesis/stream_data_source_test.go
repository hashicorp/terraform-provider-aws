// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkinesis "github.com/hashicorp/terraform-provider-aws/internal/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisStreamDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_kinesis_stream.test"
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamDataSourceConfig_basic(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					// resource.TestCheckResourceAttrPair(dataSourceName, "creation_timestamp", resourceName, "creation_timestamp"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_timestamp"),
					resource.TestCheckResourceAttrPair(dataSourceName, "closed_shards.#", resourceName, "closed_shards.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encryption_type", resourceName, "encryption_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "open_shards.#", resourceName, "shard_count"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrRetentionPeriod, resourceName, names.AttrRetentionPeriod),
					resource.TestCheckResourceAttrPair(dataSourceName, "shard_level_metrics.#", resourceName, "shard_level_metrics.#"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttrPair(dataSourceName, "stream_mode_details.0.stream_mode", resourceName, "stream_mode_details.0.stream_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
			{
				Config: testAccStreamDataSourceConfig_basic(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "closed_shards.#", "4"),
					resource.TestCheckResourceAttrPair(dataSourceName, "open_shards.#", resourceName, "shard_count"),
				),
			},
		},
	})
}

func TestAccKinesisStreamDataSource_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_kinesis_stream.test"
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamDataSourceConfig_encryption(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					// resource.TestCheckResourceAttrPair(dataSourceName, "creation_timestamp", resourceName, "creation_timestamp"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_timestamp"),
					resource.TestCheckResourceAttr(dataSourceName, "closed_shards.#", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encryption_type", resourceName, "encryption_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "open_shards.#", resourceName, "shard_count"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttrPair(dataSourceName, "stream_mode_details.0.stream_mode", resourceName, "stream_mode_details.0.stream_mode"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/40494
func TestAccKinesisStreamDataSource_pagedShards(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_kinesis_stream.test"
	const shardCount = 1100

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckShardLimitGreaterThanOrEqual(ctx, t, shardCount)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamDataSourceConfig_basic(rName, 1100),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "open_shards.#", strconv.Itoa(shardCount)),
				),
			},
		},
	})
}

func testAccPreCheckShardLimitGreaterThanOrEqual(ctx context.Context, t *testing.T, n int) {
	t.Helper()

	conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisClient(ctx)
	output, err := tfkinesis.FindLimits(ctx, conn)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if shardLimit := int(aws.ToInt32(output.ShardLimit)); n > shardLimit {
		t.Skipf("skipping tests; shard count (%d) > shard limit quota (%d)", n, shardLimit)
	}
}

func testAccStreamDataSourceConfig_basic(rName string, shardCount int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name             = %[1]q
  shard_count      = %[2]d
  retention_period = 72

  tags = {
    Name = %[1]q
  }

  shard_level_metrics = [
    "IncomingBytes",
    "OutgoingBytes"
  ]
}

data "aws_kinesis_stream" "test" {
  name = aws_kinesis_stream.test.name
}
`, rName, shardCount)
}
func testAccStreamDataSourceConfig_encryption(rName string, shardCount int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name            = %[1]q
  shard_count     = %[2]d
  encryption_type = "KMS"
  kms_key_id      = aws_kms_key.test.id
}

data "aws_kinesis_stream" "test" {
  name = aws_kinesis_stream.test.name
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}
`, rName, shardCount)
}
