// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBKinesisStreamingDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dynamodb_kinesis_streaming_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKinesisStreamingDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamingDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamingDestinationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrStreamARN, "kinesis", regexache.MustCompile(fmt.Sprintf("stream/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
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

func TestAccDynamoDBKinesisStreamingDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dynamodb_kinesis_streaming_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKinesisStreamingDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamingDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamingDestinationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdynamodb.ResourceKinesisStreamingDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDynamoDBKinesisStreamingDestination_Disappears_dynamoDBTable(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_dynamodb_kinesis_streaming_destination.test"
	tableResourceName := "aws_dynamodb_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKinesisStreamingDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamingDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamingDestinationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdynamodb.ResourceTable(), tableResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccKinesisStreamingDestinationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "hk"

  attribute {
    name = "hk"
    type = "S"
  }
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2
}

resource "aws_dynamodb_kinesis_streaming_destination" "test" {
  table_name = aws_dynamodb_table.test.name
  stream_arn = aws_kinesis_stream.test.arn
}
`, rName)
}

func testAccCheckKinesisStreamingDestinationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		_, err := tfdynamodb.FindKinesisDataStreamDestinationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrStreamARN], rs.Primary.Attributes[names.AttrTableName])

		return err
	}
}

func testAccCheckKinesisStreamingDestinationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dynamodb_kinesis_streaming_destination" {
				continue
			}

			_, err := tfdynamodb.FindKinesisDataStreamDestinationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrStreamARN], rs.Primary.Attributes[names.AttrTableName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DynamoDB Kinesis Streaming Destination %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
