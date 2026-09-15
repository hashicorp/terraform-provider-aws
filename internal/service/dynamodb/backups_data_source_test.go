// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBBackupsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_dynamodb_backups.test"
	tableResourceName := "aws_dynamodb_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupsDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("backup_summaries"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"backup_arn":                knownvalue.NotNull(),
							"backup_creation_date_time": knownvalue.NotNull(),
							"backup_expiry_date_time":   knownvalue.Null(),
							"backup_name":               knownvalue.StringExact(rName + "-backup"),
							"backup_size_bytes":         knownvalue.NotNull(),
							"backup_status":             knownvalue.NotNull(),
							"backup_type":               tfknownvalue.StringExact(awstypes.BackupTypeUser),
							"table_arn":                 knownvalue.NotNull(),
							"table_id":                  knownvalue.NotNull(),
							names.AttrTableName:         knownvalue.NotNull(),
						}),
					})),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New("backup_summaries").AtSliceIndex(0).AtMapKey("table_arn"),
						tableResourceName, tfjsonpath.New(names.AttrARN),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New("backup_summaries").AtSliceIndex(0).AtMapKey(names.AttrTableName),
						tableResourceName, tfjsonpath.New(names.AttrName),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func testAccBackupsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
}

action "aws_dynamodb_create_backup" "test" {
  config {
    table_name  = aws_dynamodb_table.test.name
    backup_name = "%[1]s-backup"
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_dynamodb_create_backup.test]
    }
  }
}

data "aws_dynamodb_backups" "test" {
  table_name = aws_dynamodb_table.test.name

  depends_on = [terraform_data.trigger]
}
`, rName)
}
