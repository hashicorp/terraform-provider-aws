// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qldb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/qldb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfqldb "github.com/hashicorp/terraform-provider-aws/internal/service/qldb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQLDBStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.JournalKinesisStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "qldb", regexache.MustCompile(`stream/.+`)),
					resource.TestCheckResourceAttr(resourceName, "exclusive_end_time", ""),
					resource.TestCheckResourceAttrSet(resourceName, "inclusive_start_time"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kinesis_configuration.0.aggregation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "kinesis_configuration.0.stream_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "ledger_name"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "stream_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccQLDBStream_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.JournalKinesisStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfqldb.ResourceStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQLDBStream_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.JournalKinesisStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccStreamConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStreamConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccQLDBStream_withEndTime(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.JournalKinesisStreamDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_endTime(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "exclusive_end_time"),
					resource.TestCheckResourceAttrSet(resourceName, "inclusive_start_time"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kinesis_configuration.0.aggregation_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "kinesis_configuration.0.stream_arn"),
				),
			},
		},
	})
}

func testAccCheckStreamExists(ctx context.Context, n string, v *types.JournalKinesisStreamDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No QLDB Stream ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QLDBClient(ctx)

		output, err := tfqldb.FindStreamByTwoPartKey(ctx, conn, rs.Primary.Attributes["ledger_name"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStreamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QLDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_qldb_stream" {
				continue
			}

			_, err := tfqldb.FindStreamByTwoPartKey(ctx, conn, rs.Primary.Attributes["ledger_name"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QLDB Stream %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStreamBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  permissions_mode    = "ALLOW_ALL"
  deletion_protection = false
}

resource "aws_kinesis_stream" "test" {
  name             = %[1]q
  shard_count      = 1
  retention_period = 24
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Sid    = ""
      Principal = {
        Service = "qldb.amazonaws.com"
      }
    }]
  })

  inline_policy {
    name = "test-qldb-policy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Action = [
          "kinesis:PutRecord*",
          "kinesis:DescribeStream",
          "kinesis:ListShards",
        ]
        Effect   = "Allow"
        Resource = aws_kinesis_stream.test.arn
      }]
    })
  }
}
`, rName)
}

func testAccStreamConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStreamBaseConfig(rName), fmt.Sprintf(`
resource "aws_qldb_stream" "test" {
  stream_name          = %[1]q
  ledger_name          = aws_qldb_ledger.test.id
  inclusive_start_time = "2021-01-01T00:00:00Z"
  role_arn             = aws_iam_role.test.arn

  kinesis_configuration {
    stream_arn = aws_kinesis_stream.test.arn
  }
}
`, rName))
}

func testAccStreamConfig_endTime(rName string) string {
	return acctest.ConfigCompose(testAccStreamBaseConfig(rName), fmt.Sprintf(`
resource "aws_qldb_stream" "test" {
  stream_name          = %[1]q
  ledger_name          = aws_qldb_ledger.test.id
  exclusive_end_time   = "2021-12-31T23:59:59Z"
  inclusive_start_time = "2021-01-01T00:00:00Z"
  role_arn             = aws_iam_role.test.arn

  kinesis_configuration {
    aggregation_enabled = false
    stream_arn          = aws_kinesis_stream.test.arn
  }
}
`, rName))
}

func testAccStreamConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccStreamBaseConfig(rName), fmt.Sprintf(`
resource "aws_qldb_stream" "test" {
  stream_name          = %[1]q
  ledger_name          = aws_qldb_ledger.test.id
  inclusive_start_time = "2021-01-01T00:00:00Z"
  role_arn             = aws_iam_role.test.arn

  kinesis_configuration {
    stream_arn = aws_kinesis_stream.test.arn
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccStreamConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccStreamBaseConfig(rName), fmt.Sprintf(`
resource "aws_qldb_stream" "test" {
  stream_name          = %[1]q
  ledger_name          = aws_qldb_ledger.test.id
  inclusive_start_time = "2021-01-01T00:00:00Z"
  role_arn             = aws_iam_role.test.arn

  kinesis_configuration {
    stream_arn = aws_kinesis_stream.test.arn
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
