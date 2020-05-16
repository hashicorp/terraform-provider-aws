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

func TestAccAwsBackupVaultNotification_basic(t *testing.T) {
	var vault backup.GetBackupVaultNotificationsOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_backup_vault_notifications.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultNotificationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultNotificationExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "backup_vault_events.#", "2"),
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

func TestAccAwsBackupVaultNotification_disappears(t *testing.T) {
	var vault backup.GetBackupVaultNotificationsOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_backup_vault_notifications.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultNotificationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultNotificationExists(resourceName, &vault),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBackupVaultNotifications(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsBackupVaultNotificationDestroy(s *terraform.State) error {
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
				return fmt.Errorf("Backup Plan notifications '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsBackupVaultNotificationExists(name string, vault *backup.GetBackupVaultNotificationsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).backupconn
		params := &backup.GetBackupVaultNotificationsInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetBackupVaultNotifications(params)
		if err != nil {
			return err
		}

		*vault = *resp

		return nil
	}
}

func testAccBackupVaultNotificationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

data "aws_iam_policy_document" "test" {
  policy_id = "__default_policy_ID"

  statement {
    actions = [
      "SNS:Publish",
    ]

    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["backup.amazonaws.com"]
    }

    resources = [
      "${aws_sns_topic.test.arn}",
    ]

    sid = "__default_statement_ID"
  }
}

resource "aws_sns_topic_policy" "test" {
  arn = "${aws_sns_topic.test.arn}"
  policy = "${data.aws_iam_policy_document.test.json}"
}

resource "aws_backup_vault_notifications" "test" {
  backup_vault_name   = "${aws_backup_vault.test.name}"
  sns_topic_arn       = "${aws_sns_topic.test.arn}"
  backup_vault_events = ["BACKUP_JOB_STARTED", "RESTORE_JOB_COMPLETED"] 
}
`, rName)
}
