// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBTag_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dynamodb_tag.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1),
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

func TestAccDynamoDBTag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dynamodb_tag.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdynamodb.ResourceTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13725
func TestAccDynamoDBTag_ResourceARN_tableReplica(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dynamodb_tag.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_resourceARNTableReplica(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
				),
			},
			{
				Config:            testAccTagConfig_resourceARNTableReplica(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDynamoDBTag_value(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dynamodb_tag.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1Updated),
				),
			},
		},
	})
}

func TestAccDynamoDBTag_multipleTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dynamodb_tag.test"
	resourceName2 := "aws_dynamodb_tag.test2"
	resourceName3 := "aws_dynamodb_tag.test3"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_multipleTags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					testAccCheckTagExists(ctx, t, resourceName2),
					testAccCheckTagExists(ctx, t, resourceName3),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName2, names.AttrKey, fmt.Sprintf("%s-2", acctest.CtKey1)),
					resource.TestCheckResourceAttr(resourceName2, names.AttrValue, fmt.Sprintf("%s-2", acctest.CtValue1)),
					resource.TestCheckResourceAttr(resourceName3, names.AttrKey, fmt.Sprintf("%s-3", acctest.CtKey1)),
					resource.TestCheckResourceAttr(resourceName3, names.AttrValue, fmt.Sprintf("%s-3", acctest.CtValue1)),
				),
			},
			{
				Config: testAccTagConfig_multipleTagsUpdate(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					testAccCheckTagExists(ctx, t, resourceName2),
					testAccCheckTagExists(ctx, t, resourceName3),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName2, names.AttrKey, fmt.Sprintf("%s-2", acctest.CtKey1)),
					resource.TestCheckResourceAttr(resourceName2, names.AttrValue, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName3, names.AttrKey, fmt.Sprintf("%s-3", acctest.CtKey1)),
					resource.TestCheckResourceAttr(resourceName3, names.AttrValue, fmt.Sprintf("%s-3", acctest.CtValue1)),
				),
			},
			{
				Config: testAccTagConfig_multipleTagsRemove(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					testAccCheckTagExists(ctx, t, resourceName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName2, names.AttrKey, fmt.Sprintf("%s-2", acctest.CtKey1)),
					resource.TestCheckResourceAttr(resourceName2, names.AttrValue, fmt.Sprintf("%s-2", acctest.CtValue1)),
				),
			},
		},
	})
}

func testAccTagConfig_basic(rName string, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_dynamodb_tag" "test" {
  resource_arn = aws_dynamodb_table.test.arn
  key          = %[2]q
  value        = %[3]q
}
`, rName, key, value)
}

func testAccTagConfig_resourceARNTableReplica(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  region = %[2]q
}

data "aws_region" "current" {}

resource "aws_dynamodb_table" "test" {
  region = %[2]q

  billing_mode     = "PAY_PER_REQUEST"
  hash_key         = "TestTableHashKey"
  name             = %[1]q
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.current.region
  }
}

resource "aws_dynamodb_tag" "test" {
  resource_arn = replace(aws_dynamodb_table.test.arn, data.aws_region.alternate.name, data.aws_region.current.region)
  key          = "testkey"
  value        = "testvalue"
}
`, rName, acctest.AlternateRegion()))
}

func testAccTagConfig_multipleTags(rName string, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_dynamodb_tag" "test" {
  resource_arn = aws_dynamodb_table.test.arn
  key          = %[2]q
  value        = %[3]q
}

resource "aws_dynamodb_tag" "test2" {
  resource_arn = aws_dynamodb_table.test.arn
  key          = "%[2]s-2"
  value        = "%[3]s-2"
}

resource "aws_dynamodb_tag" "test3" {
  resource_arn = aws_dynamodb_table.test.arn
  key          = "%[2]s-3"
  value        = "%[3]s-3"
}
`, rName, key, value)
}

func testAccTagConfig_multipleTagsUpdate(rName string, key string, value, value2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_dynamodb_tag" "test" {
  resource_arn = aws_dynamodb_table.test.arn
  key          = %[2]q
  value        = %[3]q
}

resource "aws_dynamodb_tag" "test2" {
  resource_arn = aws_dynamodb_table.test.arn
  key          = "%[2]s-2"
  value        = %[4]q
}

resource "aws_dynamodb_tag" "test3" {
  resource_arn = aws_dynamodb_table.test.arn
  key          = "%[2]s-3"
  value        = "%[3]s-3"
}
`, rName, key, value, value2)
}

func testAccTagConfig_multipleTagsRemove(rName string, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_dynamodb_tag" "test" {
  resource_arn = aws_dynamodb_table.test.arn
  key          = %[2]q
  value        = %[3]q
}

resource "aws_dynamodb_tag" "test2" {
  resource_arn = aws_dynamodb_table.test.arn
  key          = "%[2]s-2"
  value        = "%[3]s-2"
}
`, rName, key, value)
}
