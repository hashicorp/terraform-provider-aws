package backup_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
)

func TestAccBackupVaultNotification_basic(t *testing.T) {
	var vault backup.GetBackupVaultNotificationsOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_notifications.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultNotificationsConfig_notification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultNotificationExists(resourceName, &vault),
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

func TestAccBackupVaultNotification_disappears(t *testing.T) {
	var vault backup.GetBackupVaultNotificationsOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_notifications.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultNotificationsConfig_notification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultNotificationExists(resourceName, &vault),
					acctest.CheckResourceDisappears(acctest.Provider, tfbackup.ResourceVaultNotifications(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVaultNotificationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault_notifications" {
			continue
		}

		input := &backup.GetBackupVaultNotificationsInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetBackupVaultNotifications(input)

		if err == nil {
			if aws.StringValue(resp.BackupVaultName) == rs.Primary.ID {
				return fmt.Errorf("Backup Plan notifications '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckVaultNotificationExists(name string, vault *backup.GetBackupVaultNotificationsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn
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

func testAccVaultNotificationsConfig_notification(rName string) string {
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
  arn    = aws_sns_topic.test.arn
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_backup_vault_notifications" "test" {
  backup_vault_name   = aws_backup_vault.test.name
  sns_topic_arn       = aws_sns_topic.test.arn
  backup_vault_events = ["BACKUP_JOB_STARTED", "RESTORE_JOB_COMPLETED"]
}
`, rName)
}
