// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBTableItemDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_item.test"
	hashKey := "hashKey"
	itemContent := `{
	"hashKey": {"S": "something"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"three": {"N": "33333"},
	"four": {"N": "44444"}
}`
	key := `{
	"hashKey": {"S": "something"}
}`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableItemDataSourceConfig_basic(rName, hashKey, itemContent, key),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "item", itemContent),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrTableName, rName),
				),
			},
		},
	})
}

func TestAccDynamoDBTableItemDataSource_projectionExpression(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_item.test"
	hashKey := "hashKey"
	projectionExpression := "one,two"
	itemContent := `{
	"hashKey": {"S": "something"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"three": {"N": "33333"},
	"four": {"N": "44444"}
}`
	key := `{
	"hashKey": {"S": "something"}
}`

	expected := `{
	"one": {"N": "11111"},
	"two": {"N": "22222"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableItemDataSourceConfig_projectionExpression(rName, hashKey, itemContent, projectionExpression, key),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "item", expected),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(dataSourceName, "projection_expression", projectionExpression),
				),
			},
		},
	})
}

func TestAccDynamoDBTableItemDataSource_expressionAttributeNames(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_item.test"
	hashKey := "hashKey"
	itemContent := `{
	"hashKey": {"S": "something"},
	"one": {"N": "11111"},
	"Percentile": {"N": "22222"}
}`
	key := `{
	"hashKey": {"S": "something"}
}`

	expected := `{
	"Percentile": {"N": "22222"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableItemDataSourceConfig_expressionAttributeNames(rName, hashKey, itemContent, key),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "item", expected),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(dataSourceName, "projection_expression", "#P"),
				),
			},
		},
	})
}

func testAccTableItemDataSourceConfig_basic(tableName, hashKey, item string, key string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = %[2]q

  attribute {
    name = %[3]q
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key

  item = <<ITEM
%[4]s
ITEM
}

data "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name

  key        = <<KEY
%[5]s
KEY
  depends_on = [aws_dynamodb_table_item.test]
}
`, tableName, hashKey, hashKey, item, key)
}

func testAccTableItemDataSourceConfig_projectionExpression(tableName, hashKey, item, projectionExpression, key string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = %[2]q

  attribute {
    name = %[3]q
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key

  item = <<ITEM
%[4]s
ITEM
}

data "aws_dynamodb_table_item" "test" {
  table_name            = aws_dynamodb_table.test.name
  projection_expression = %[5]q
  key                   = <<KEY
%[6]s
KEY
  depends_on            = [aws_dynamodb_table_item.test]
}
`, tableName, hashKey, hashKey, item, projectionExpression, key)
}

func testAccTableItemDataSourceConfig_expressionAttributeNames(tableName, hashKey, item string, key string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = %[2]q

  attribute {
    name = %[3]q
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key

  item = <<ITEM
%[4]s
ITEM
}

data "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  expression_attribute_names = {
    "#P" = "Percentile"
  }
  projection_expression = "#P"
  key                   = <<KEY
%[5]s
KEY
  depends_on            = [aws_dynamodb_table_item.test]
}
`, tableName, hashKey, hashKey, item, key)
}
