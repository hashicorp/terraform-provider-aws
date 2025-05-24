// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfdynamo "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
)

func TestAccDynamoDBTableQueryDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_query.test"
	hashKey := "ID"
	hashKeyValue := "0000A"
	itemContent := `{
		"ID": {"S": "0000A"},
		"one": {"N": "11111"},
		"two": {"N": "22222"}
	}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableQueryDataSourceConfig_basic(rName, itemContent, hashKey, hashKeyValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "items.#", "1"),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "items.0", itemContent),
					resource.TestCheckResourceAttr(dataSourceName, "table_name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "scanned_count", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "item_count", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "query_count", "1"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableQueryDataSource_projectionExpression(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_query.test"
	hashKey := "ID"
	hashKeyValue := "0000A"
	itemContent := `{
		"ID": {"S": "0000A"},
		"one": {"N": "11111"},
		"two": {"N": "22222"},
		"three": {"N": "33333"},
		"four": {"N": "44444"}
	}`

	projectionExpression := "one,two"

	expected := `{
		"one": {"N": "11111"},
		"two": {"N": "22222"}
	}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableQueryDataSourceConfig_projectionExpression(rName, itemContent, projectionExpression, hashKey, hashKeyValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "items.#", "1"),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "items.0", expected),
					resource.TestCheckResourceAttr(dataSourceName, "table_name", rName),
				),
			},
		},
	})
}

func TestAccDynamoDBTableQueryDataSource_expressionAttributeNames(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_query.test"
	hashKey := "ID"
	hashKeyValue := "0000A"
	itemContent := `{
		"ID": {"S": "0000A"},
		"one": {"N": "11111"},
		"two": {"N": "22222"},
		"Percentile": {"N": "33333"}
	}`

	projectionExpression := `#P`
	expressionAttributeNames := `{
    "#P" = "Percentile"
  }`

	expected := `{
		"Percentile": {"N": "33333"}
	}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableQueryDataSourceConfig_expressionAttributeNames(rName, itemContent, hashKey, hashKeyValue, expressionAttributeNames, projectionExpression),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "items.#", "1"),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "items.0", expected),
					resource.TestCheckResourceAttr(dataSourceName, "table_name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "projection_expression", "#P"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableQueryDataSource_select(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_query.test"
	hashKey := "ID"
	hashKeyValue := "0000A"
	itemContent := `{
		"ID": {"S": "0000A"},
		"one": {"N": "11111"},
		"two": {"N": "22222"},
		"three": {"N": "33333"},
		"four": {"N": "44444"}
	}`

	selectValue := "COUNT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableQueryDataSourceConfig_select(rName, itemContent, hashKey, hashKeyValue, selectValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "items.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "item_count", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "table_name", rName),
				),
			},
		},
	})
}

func TestAccDynamoDBTableQueryDataSource_filterExpression(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_query.test"
	item1 := `{"ID": {"S": "0000A"}, "Category": {"S": "A"}}`
	item2 := `{"ID": {"S": "0000B"}, "Category": {"S": "B"}}`
	expressionAttributeValues := `{
		":value" = jsonencode({"S": "0000A"})
		":category" = jsonencode({"S": "A"})
  }`

	filterExpression := "Category = :category"

	expected := `{
  "ID": {"S": "0000A"},
  "Category": {"S": "A"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableQueryDataSourceConfig_filterExpression(rName, item1, item2, expressionAttributeValues, filterExpression),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "items.#", "1"),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "items.0", expected),
				),
			},
		},
	})
}

func TestAccDynamoDBTableQueryDataSource_scanIndexForward(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_query.test"
	sortKeyValue1 := "1"
	sortKeyValue2 := "2"
	sortKeyValue3 := "3"

	scanIndexForward := "false"

	expected := []string{
		`{"ID":{"S":"0000A"},"sortKey":{"N":"3"}}`,
		`{"ID":{"S":"0000A"},"sortKey":{"N":"2"}}`,
		`{"ID":{"S":"0000A"},"sortKey":{"N":"1"}}`,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableQueryDataSourceConfig_scanIndexForward(rName, sortKeyValue1, sortKeyValue2, sortKeyValue3, scanIndexForward),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "items.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "scan_index_forward", scanIndexForward),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "items.0", expected[0]),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "items.1", expected[1]),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "items.2", expected[2]),
					resource.TestCheckResourceAttr(dataSourceName, "table_name", rName),
				),
			},
		},
	})
}

func TestAccDynamoDBTableQueryDataSource_index(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_query.test"
	itemContent := `{
		"ID": {"S": "0000A"},
		"value": {"N": "1111"},
		"extraAttribute": {"S": "additionalValue"}
	}`

	projectionType := "INCLUDE"
	indexName := "exampleIndex"

	expected := []string{
		`{"ID":{"S":"0000A"},"extraAttribute":{"S":"additionalValue"}}`,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableQueryDataSourceConfig_index(rName, itemContent, projectionType, indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "items.#", "1"),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "items.0", expected[0]),
					resource.TestCheckResourceAttr(dataSourceName, "table_name", rName),
				),
			},
		},
	})
}

func TestAccDynamoDBTableQueryDataSource_handlesPagination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_query.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableQueryDataSourceConfig_handlesPagination(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "items.#", "10"),
					resource.TestCheckResourceAttr(dataSourceName, "table_name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "scanned_count", "10"),
					resource.TestCheckResourceAttr(dataSourceName, "item_count", "10"),
					resource.TestCheckResourceAttr(dataSourceName, "query_count", "3"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableQueryDataSource_outputLimit(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_table_query.test"

	outputLimit := 5

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableQueryDataSourceConfig_outputLimit(rName, outputLimit),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "items.#", "5"),
					resource.TestCheckResourceAttr(dataSourceName, "output_limit", "5"),
					resource.TestCheckResourceAttr(dataSourceName, "table_name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "item_count", "5"),
					resource.TestCheckResourceAttr(dataSourceName, "query_count", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "scanned_count", "8"),
				),
			},
		},
	})
}

func testAccTableQueryDataSourceConfig_basic(tableName, item, hashKey, hashKeyValue string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = %q

  attribute {
    name = %q
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  item = <<ITEM
%s
ITEM
}

data "aws_dynamodb_table_query" "test" {
  table_name                  = aws_dynamodb_table.test.name
	key_condition_expression    = "%s = :value"
	expression_attribute_values = {
		":value"= jsonencode({"S" = %q})
	}
  depends_on                  = [aws_dynamodb_table_item.test]
}
`, tableName, hashKey, hashKey, item, hashKey, hashKeyValue)
}

func testAccTableQueryDataSourceConfig_projectionExpression(tableName, item, projectionExpression, hashKey, hashKeyValue string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = %q

  attribute {
    name = %q
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  item = <<ITEM
%s
ITEM
}

data "aws_dynamodb_table_query" "test" {
  projection_expression = %q
  table_name                  = aws_dynamodb_table.test.name
	key_condition_expression    = "%s = :value"
	expression_attribute_values = {
		":value"= jsonencode({"S" = %q})
	}
  depends_on                  = [aws_dynamodb_table_item.test]
}
`, tableName, hashKey, hashKey, item, projectionExpression, hashKey, hashKeyValue)
}

func testAccTableQueryDataSourceConfig_expressionAttributeNames(tableName, item, hashKey, hashKeyValue, expressionAttributeNames, projectionExpression string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = %q

  attribute {
    name = %q
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  item = <<ITEM
%s
ITEM
}

data "aws_dynamodb_table_query" "test" {
	expression_attribute_names = %s
  projection_expression = %q
  table_name                  = aws_dynamodb_table.test.name
	key_condition_expression    = "%s = :value"
	expression_attribute_values = {
		":value"= jsonencode({"S" = %q})
	}
  depends_on                  = [aws_dynamodb_table_item.test]
}
`, tableName, hashKey, hashKey, item, expressionAttributeNames, projectionExpression, hashKey, hashKeyValue)
}

func testAccTableQueryDataSourceConfig_select(tableName, item, hashKey, hashKeyValue, selectValue string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = %q

  attribute {
    name = %q
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  item = <<ITEM
%s
ITEM
}

	data "aws_dynamodb_table_query" "test" {
		select = %q
	  table_name                  = aws_dynamodb_table.test.name
		key_condition_expression    = "%s = :value"
		expression_attribute_values = {
			":value"= jsonencode({"S" = %q})
		}
	  depends_on                  = [aws_dynamodb_table_item.test]
	}

`, tableName, hashKey, hashKey, item, selectValue, hashKey, hashKeyValue)
}

func testAccTableQueryDataSourceConfig_filterExpression(tableName, item1, item2, expressionAttributeValues, filterExpression string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "ID"

  attribute {
    name = "ID"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "item1" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  item       = %q
}

resource "aws_dynamodb_table_item" "item2" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  item       = %q
}

data "aws_dynamodb_table_query" "test" {
  table_name                  = aws_dynamodb_table.test.name
	key_condition_expression    = "ID = :value"
  expression_attribute_values = %s
  filter_expression = %q
  depends_on        = [aws_dynamodb_table_item.item1, aws_dynamodb_table_item.item2]
}
`, tableName, item1, item2, expressionAttributeValues, filterExpression)
}

func testAccTableQueryDataSourceConfig_scanIndexForward(tableName, sortKeyValue1, sortKeyValue2, sortKeyValue3, scanIndexForward string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "ID"
  range_key      = "sortKey"

  attribute {
    name = "ID"
    type = "S"
  }

  attribute {
    name = "sortKey"
    type = "N"
  }
}

resource "aws_dynamodb_table_item" "item1" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  range_key  = aws_dynamodb_table.test.range_key
  item = <<ITEM
{
  "ID": {"S": "0000A"},
  "sortKey": {"N": %q}
}
ITEM
}

resource "aws_dynamodb_table_item" "item2" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  range_key  = aws_dynamodb_table.test.range_key
  item = <<ITEM
{
  "ID": {"S": "0000A"},
  "sortKey": {"N": %q}
}
ITEM
}

resource "aws_dynamodb_table_item" "item3" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  range_key  = aws_dynamodb_table.test.range_key
  item = <<ITEM
{
  "ID": {"S": "0000A"},
  "sortKey": {"N": %q}
}
ITEM
}

data "aws_dynamodb_table_query" "test" {
  table_name                = aws_dynamodb_table.test.name
  key_condition_expression  = "ID = :value"
  expression_attribute_values = {
    ":value" = "{\"S\": \"0000A\"}"
  }
  scan_index_forward = %s
	consistent_read = true
  depends_on         = [aws_dynamodb_table_item.item1, aws_dynamodb_table_item.item2, aws_dynamodb_table_item.item3]
}
`, tableName, sortKeyValue1, sortKeyValue2, sortKeyValue3, scanIndexForward)
}

func testAccTableQueryDataSourceConfig_index(tableName, item, projectionType, GSIName string) string {
	return fmt.Sprintf(`
	resource "aws_dynamodb_table" "test" {
		name           = %q
		read_capacity  = 10
		write_capacity = 10
		hash_key       = "ID"
	
		attribute {
			name = "ID"
			type = "S"
		}

		global_secondary_index {
			name            = %q
			hash_key        = "ID"
			read_capacity   = 5
			write_capacity  = 5
			projection_type = %q
			non_key_attributes = ["extraAttribute"]
		}
	}
	
	resource "aws_dynamodb_table_item" "item" {
		table_name = aws_dynamodb_table.test.name
		hash_key   = aws_dynamodb_table.test.hash_key
		item = <<ITEM
%s
	ITEM
	}

	data "aws_dynamodb_table_query" "test" {
		table_name               = %q
		key_condition_expression  = "ID = :value"
		expression_attribute_values = {
			":value" = "{\"S\": \"0000A\"}"
		}
		index_name              = %q
		depends_on              = [aws_dynamodb_table_item.item]
	}
`, tableName, GSIName, projectionType, item, tableName, GSIName)
}

func generateRandomPayload(sizeKB int) string {
	payload := make([]byte, sizeKB*1024)
	for i := range payload {
		payload[i] = byte('A' + i%26)
	}
	return string(payload)
}

func testAccTableQueryDataSourceConfig_handlesPagination(tableName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "ID"
	range_key      = "sortKey"

  attribute {
    name = "ID"
    type = "S"
  }
	
	attribute {
    name = "sortKey"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
	count = 10
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
	range_key  = aws_dynamodb_table.test.range_key
  item = <<ITEM
	{
		"ID": {"S": "0000A"},
		"sortKey": {"S": "sortValue${count.index + 1}"},
		"payload": {"S": %q}
	}
ITEM
}

data "aws_dynamodb_table_query" "test" {
  table_name                  = aws_dynamodb_table.test.name
	key_condition_expression    = "ID = :value"
	expression_attribute_values = {
		":value"= jsonencode({"S" = "0000A"})
	}
  depends_on                  = [aws_dynamodb_table_item.test[0], aws_dynamodb_table_item.test[1], aws_dynamodb_table_item.test[2], aws_dynamodb_table_item.test[3], aws_dynamodb_table_item.test[4], aws_dynamodb_table_item.test[5], aws_dynamodb_table_item.test[6], aws_dynamodb_table_item.test[7], aws_dynamodb_table_item.test[8], aws_dynamodb_table_item.test[9]]
}
`, tableName, generateRandomPayload(300))
}

func testAccTableQueryDataSourceConfig_outputLimit(tableName string, outputLimit int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "ID"
	range_key      = "sortKey"

  attribute {
    name = "ID"
    type = "S"
  }
	
	attribute {
    name = "sortKey"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
	count = 10
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
	range_key  = aws_dynamodb_table.test.range_key
  item = <<ITEM
	{
		"ID": {"S": "0000A"},
		"sortKey": {"S": "sortValue${count.index + 1}"},
		"payload": {"S": %q}
	}
ITEM
}

data "aws_dynamodb_table_query" "test" {
	output_limit 							  = %d
  table_name                  = aws_dynamodb_table.test.name
	key_condition_expression    = "ID = :value"
	expression_attribute_values = {
		":value"= jsonencode({"S" = "0000A"})
	}
  depends_on                  = [aws_dynamodb_table_item.test[0], aws_dynamodb_table_item.test[1], aws_dynamodb_table_item.test[2], aws_dynamodb_table_item.test[3], aws_dynamodb_table_item.test[4], aws_dynamodb_table_item.test[5], aws_dynamodb_table_item.test[6], aws_dynamodb_table_item.test[7], aws_dynamodb_table_item.test[8], aws_dynamodb_table_item.test[9]]
}
`, tableName, generateRandomPayload(300), outputLimit)
}

func TestConvertJSONToAttributeValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		jsonStr     string
		expectedAV  *tfdynamo.AttributeValue
		expectedErr bool
	}{
		{
			jsonStr: "{\"S\":\"example\"}",
			expectedAV: &tfdynamo.AttributeValue{
				S: strPtr("example"),
			},
			expectedErr: false,
		},
		{
			jsonStr: `{"B":"YmFzZTY0IGVuY29kaW5nIGVuY3J5cHRpb24="}`,
			expectedAV: &tfdynamo.AttributeValue{
				B: []byte("base64 encoding encryption"),
			},
			expectedErr: false,
		},
		{
			jsonStr: `{"BOOL":true}`,
			expectedAV: &tfdynamo.AttributeValue{
				BOOL: boolPtr(true),
			},
			expectedErr: false,
		},
		{
			jsonStr: `{"BS":["YmFzZTY0","ZW5jb2Rpbmc="]}`,
			expectedAV: &tfdynamo.AttributeValue{
				BS: [][]byte{[]byte("base64"), []byte("encoding")},
			},
			expectedErr: false,
		},
		{
			jsonStr: `{"L":[{"S":"value1"},{"S":"value2"}]}`,
			expectedAV: &tfdynamo.AttributeValue{
				L: []*tfdynamo.AttributeValue{
					{S: strPtr("value1")},
					{S: strPtr("value2")},
				},
			},
			expectedErr: false,
		},
		{
			jsonStr: `{"M":{"key1":{"S":"value1"},"key2":{"S":"value2"}}}`,
			expectedAV: &tfdynamo.AttributeValue{
				M: map[string]*tfdynamo.AttributeValue{
					"key1": {S: strPtr("value1")},
					"key2": {S: strPtr("value2")},
				},
			},
			expectedErr: false,
		},
		{
			jsonStr: `{"N":"12345"}`,
			expectedAV: &tfdynamo.AttributeValue{
				N: strPtr("12345"),
			},
			expectedErr: false,
		},
		{
			jsonStr: `{"NS":["123","456"]}`,
			expectedAV: &tfdynamo.AttributeValue{
				NS: []*string{strPtr("123"), strPtr("456")},
			},
			expectedErr: false,
		},
		{
			jsonStr: `{"NULL":true}`,
			expectedAV: &tfdynamo.AttributeValue{
				NULL: boolPtr(true),
			},
			expectedErr: false,
		},
		{
			jsonStr: `{"SS":["value1","value2"]}`,
			expectedAV: &tfdynamo.AttributeValue{
				SS: []*string{strPtr("value1"), strPtr("value2")},
			},
			expectedErr: false,
		},
		{
			jsonStr:     `invalidJSON`,
			expectedAV:  nil,
			expectedErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.jsonStr, func(t *testing.T) {
			result, err := tfdynamo.ConvertJSONToAttributeValue(c.jsonStr)
			log.Println("[ERROR] result: ", result)

			if c.expectedErr && err == nil {
				t.Errorf("Expected error, but got nil")
			}

			if !c.expectedErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, c.expectedAV) {
				t.Errorf("Unexpected result. Expected %+v, got %+v", c.expectedAV, result)
			}
		})
	}
}

func TestConvertToDynamoAttributeValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		av          *tfdynamo.AttributeValue
		expectedAV  *dynamodb.AttributeValue
		expectedErr bool
	}{
		{
			av: &tfdynamo.AttributeValue{
				S: strPtr("example"),
			},
			expectedAV: &dynamodb.AttributeValue{
				S: strPtr("example"),
			},
			expectedErr: false,
		},
		{
			av: &tfdynamo.AttributeValue{
				B: []byte("base64 encoding encryption"),
			},
			expectedAV: &dynamodb.AttributeValue{
				B: []byte("base64 encoding encryption"),
			},
			expectedErr: false,
		},
		{
			av: &tfdynamo.AttributeValue{
				BOOL: boolPtr(true),
			},
			expectedAV: &dynamodb.AttributeValue{
				BOOL: boolPtr(true),
			},
			expectedErr: false,
		},
		{
			av: &tfdynamo.AttributeValue{
				BS: [][]byte{[]byte("base64"), []byte("encoding")},
			},
			expectedAV: &dynamodb.AttributeValue{
				BS: [][]byte{[]byte("base64"), []byte("encoding")},
			},
			expectedErr: false,
		},
		{
			av: &tfdynamo.AttributeValue{
				L: []*tfdynamo.AttributeValue{
					{S: strPtr("value1")},
					{S: strPtr("value2")},
				},
			},
			expectedAV: &dynamodb.AttributeValue{
				L: []*dynamodb.AttributeValue{
					{S: strPtr("value1")},
					{S: strPtr("value2")},
				},
			},
			expectedErr: false,
		},
		{
			av: &tfdynamo.AttributeValue{
				M: map[string]*tfdynamo.AttributeValue{
					"key1": {S: strPtr("value1")},
					"key2": {S: strPtr("value2")},
				},
			},
			expectedAV: &dynamodb.AttributeValue{
				M: map[string]*dynamodb.AttributeValue{
					"key1": {S: strPtr("value1")},
					"key2": {S: strPtr("value2")},
				},
			},
			expectedErr: false,
		},
		{
			av: &tfdynamo.AttributeValue{
				N: strPtr("12345"),
			},
			expectedAV: &dynamodb.AttributeValue{
				N: strPtr("12345"),
			},
			expectedErr: false,
		},
		{
			av: &tfdynamo.AttributeValue{
				NS: []*string{strPtr("123"), strPtr("456")},
			},
			expectedAV: &dynamodb.AttributeValue{
				NS: []*string{strPtr("123"), strPtr("456")},
			},
			expectedErr: false,
		},
		{
			av: &tfdynamo.AttributeValue{
				NULL: boolPtr(true),
			},
			expectedAV: &dynamodb.AttributeValue{
				NULL: boolPtr(true),
			},
			expectedErr: false,
		},
		{
			av: &tfdynamo.AttributeValue{
				SS: []*string{strPtr("value1"), strPtr("value2")},
			},
			expectedAV: &dynamodb.AttributeValue{
				SS: []*string{strPtr("value1"), strPtr("value2")},
			},
			expectedErr: false,
		},
		{
			av:          nil,
			expectedAV:  nil,
			expectedErr: false,
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			result, err := tfdynamo.ConvertToDynamoAttributeValue(c.av)

			if c.expectedErr && err == nil {
				t.Errorf("Expected error, but got nil")
			}

			if !c.expectedErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, c.expectedAV) {
				t.Errorf("Unexpected result. Expected %+v, got %+v", c.expectedAV, result)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
