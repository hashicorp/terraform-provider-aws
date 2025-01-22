// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBTableDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "read_capacity", resourceName, "read_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "write_capacity", resourceName, "write_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "hash_key", resourceName, "hash_key"),
					resource.TestCheckResourceAttrPair(datasourceName, "range_key", resourceName, "range_key"),
					resource.TestCheckResourceAttrPair(datasourceName, "attribute.#", resourceName, "attribute.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "global_secondary_index.#", resourceName, "global_secondary_index.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ttl.#", resourceName, "ttl.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Environment", resourceName, "tags.Environment"),
					resource.TestCheckResourceAttrPair(datasourceName, "server_side_encryption.#", resourceName, "server_side_encryption.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "billing_mode", resourceName, "billing_mode"),
					resource.TestCheckResourceAttrPair(datasourceName, "point_in_time_recovery.#", resourceName, "point_in_time_recovery.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "point_in_time_recovery.0.enabled", resourceName, "point_in_time_recovery.0.enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "table_class", resourceName, "table_class"),
				),
			},
		},
	})
}

func testAccTableDataSourceConfig_basic(tableName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "UserId"
  range_key      = "GameTitle"

  attribute {
    name = "UserId"
    type = "S"
  }

  attribute {
    name = "GameTitle"
    type = "S"
  }

  attribute {
    name = "TopScore"
    type = "N"
  }

  global_secondary_index {
    name               = "GameTitleIndex"
    hash_key           = "GameTitle"
    range_key          = "TopScore"
    write_capacity     = 10
    read_capacity      = 10
    projection_type    = "INCLUDE"
    non_key_attributes = ["UserId"]
  }

  tags = {
    Name        = "dynamodb-table-1"
    Environment = "test"
  }
}

data "aws_dynamodb_table" "test" {
  name = aws_dynamodb_table.test.name
}
`, tableName)
}
