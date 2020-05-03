package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsBackupVaultNotifications_basic(t *testing.T) {
	var notifications backup.GetBackupVaultNotificationsOutput

	rInt := acctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultNotificationsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultNotificationsConfigBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultNotificationsExists(resourceName, &notifications),
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

func testAccCheckAwsBackupVaultNotificationsExists(name string, notifications *backup.GetBackupVaultNotificationsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: #{name}, #{s.RootModule(}.Resources")
		}

		conn := testAccProvider.Meta().(*AWSClient).backupconn

		input := &backup.GetBackupVaultNotificationsInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetBackupVaultNotifications(input)

		if err != nil {
			return err
		}

		*notifications = *output
		return nil
	}
}

func testAccCheckAwsBackupVaultNotificationsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).backupconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault_notifications" {
			continue
		}

		input := &backup.GetBackupVaultNotificationsInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetBackupVaultNotifications(input)

		if err == nil {
			if *resp.BackupVaultName == rs.Primary.ID {
				return fmt.Errorf("VaultNotifications '#{rs.Primary.ID}' was not deleted properly")
			}
		}
	}

	return nil
}

func testAccBackupVaultNotificationsConfigBase(rInt int) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_sns_topic" "test" {
  name = "terraform-test-topic-%d"
}
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"
}
`, rInt, rInt)
}

func testAccBackupVaultNotificationsConfigBasic(rInt int) string {
	return testAccBackupVaultNotificationsConfigBase(rInt) + fmt.Sprintf(`
resource "aws_backup_vault_notifications" "test" {
  vault_name    = "tf_acc_test_backup_vault_%d"
  sns_topic_arn = aws_sns_topic.test.arn

  events = [
    "BACKUP_JOB_STARTED",
    "BACKUP_JOB_COMPLETED",
    "BACKUP_JOB_SUCCESSFUL",
  ]
}
`, rInt)
}
