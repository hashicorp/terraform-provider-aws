// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisStreamConsumerDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_kinesis_stream_consumer.test"
	resourceName := "aws_kinesis_stream_consumer.test"
	streamName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConsumerDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStreamARN, streamName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_timestamp"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccKinesisStreamConsumerDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_kinesis_stream_consumer.test"
	resourceName := "aws_kinesis_stream_consumer.test"
	streamName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConsumerDataSourceConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStreamARN, streamName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_timestamp"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccKinesisStreamConsumerDataSource_arn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_kinesis_stream_consumer.test"
	resourceName := "aws_kinesis_stream_consumer.test"
	streamName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConsumerDataSourceConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStreamARN, streamName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_timestamp"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
				),
			},
		},
	})
}

func testAccStreamConsumerDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2
}
`, rName)
}

func testAccStreamConsumerDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStreamConsumerDataSourceConfig_base(rName), fmt.Sprintf(`
data "aws_kinesis_stream_consumer" "test" {
  stream_arn = aws_kinesis_stream_consumer.test.stream_arn
}

resource "aws_kinesis_stream_consumer" "test" {
  name       = %[1]q
  stream_arn = aws_kinesis_stream.test.arn
}
`, rName))
}

func testAccStreamConsumerDataSourceConfig_name(rName string) string {
	return acctest.ConfigCompose(testAccStreamConsumerDataSourceConfig_base(rName), fmt.Sprintf(`
data "aws_kinesis_stream_consumer" "test" {
  name       = aws_kinesis_stream_consumer.test.name
  stream_arn = aws_kinesis_stream_consumer.test.stream_arn
}

resource "aws_kinesis_stream_consumer" "test" {
  name       = %[1]q
  stream_arn = aws_kinesis_stream.test.arn
}
`, rName))
}

func testAccStreamConsumerDataSourceConfig_arn(rName string) string {
	return acctest.ConfigCompose(testAccStreamConsumerDataSourceConfig_base(rName), fmt.Sprintf(`
data "aws_kinesis_stream_consumer" "test" {
  arn        = aws_kinesis_stream_consumer.test.arn
  stream_arn = aws_kinesis_stream_consumer.test.stream_arn
}

resource "aws_kinesis_stream_consumer" "test" {
  name       = %[1]q
  stream_arn = aws_kinesis_stream.test.arn
}
`, rName))
}
