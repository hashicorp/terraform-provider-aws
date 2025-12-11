// Copyright IBM Corp. 2014, 2025
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

func TestAccDynamoDBTables_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_tables.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTablesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "names.#", 1),
				),
			},
		},
	})
}

func testAccTablesDataSourceConfig_basic(tableName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "TestId"

  attribute {
    name = "TestId"
    type = "S"
  }
}

data "aws_dynamodb_tables" "test" {
  depends_on = [aws_dynamodb_table.test]
}
`, tableName)
}
