// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBCreateBackupAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateBackupActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, rName),
				),
			},
		},
	})
}

func TestAccDynamoDBCreateBackupAction_customName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	backupName := acctest.RandomWithPrefix(t, "tf-test-backup")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateBackupActionConfig_customName(rName, backupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupExists(ctx, rName),
					testAccCheckBackupName(ctx, rName, backupName),
				),
			},
		},
	})
}

func TestAccDynamoDBCreateBackupAction_nonExistentTable(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCreateBackupActionConfig_nonExistentTable(rName),
				ExpectError: regexache.MustCompile(`TableNotFoundException`),
			},
		},
	})
}

func testAccCheckBackupExists(ctx context.Context, tableName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		input := &dynamodb.ListBackupsInput{
			TableName: &tableName,
		}

		output, err := conn.ListBackups(ctx, input)
		if err != nil {
			return fmt.Errorf("error listing backups for table %s: %w", tableName, err)
		}

		if len(output.BackupSummaries) == 0 {
			return fmt.Errorf("no backups found for table %s", tableName)
		}

		return nil
	}
}

func testAccCheckBackupName(ctx context.Context, tableName, expectedBackupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		input := &dynamodb.ListBackupsInput{
			TableName: &tableName,
		}

		output, err := conn.ListBackups(ctx, input)
		if err != nil {
			return fmt.Errorf("error listing backups for table %s: %w", tableName, err)
		}

		for _, backup := range output.BackupSummaries {
			if backup.BackupName != nil && *backup.BackupName == expectedBackupName {
				return nil
			}
		}

		return fmt.Errorf("backup with name %s not found for table %s", expectedBackupName, tableName)
	}
}

func testAccCreateBackupActionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }
}
`, rName)
}

func testAccCreateBackupActionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccCreateBackupActionConfig_base(rName),
		`
action "aws_dynamodb_create_backup" "test" {
  config {
    table_name = aws_dynamodb_table.test.name
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
`)
}

func testAccCreateBackupActionConfig_customName(rName, backupName string) string {
	return acctest.ConfigCompose(
		testAccCreateBackupActionConfig_base(rName),
		fmt.Sprintf(`
action "aws_dynamodb_create_backup" "test" {
  config {
    table_name  = aws_dynamodb_table.test.name
    backup_name = %[1]q
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
`, backupName))
}

func testAccCreateBackupActionConfig_nonExistentTable(rName string) string {
	return fmt.Sprintf(`
action "aws_dynamodb_create_backup" "test" {
  config {
    table_name = %[1]q
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
`, rName)
}
