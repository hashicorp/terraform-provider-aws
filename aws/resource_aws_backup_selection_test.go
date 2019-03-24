package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsBackupSelection_basic(t *testing.T) {
	var selection backup.Selection
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupSelectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsBackupSelectionConfig_create(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists("aws_backup_selection.tftest", &selection),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "name", fmt.Sprintf("tftest-%d", rInt)),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "resources.#", "2"),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "tag_condition.#", "2"),
				),
			},
			{
				Config: testAccAwsBackupSelectionConfig_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists("aws_backup_selection.tftest", &selection),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "name", fmt.Sprintf("tftest-%d", rInt)),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "resources.#", "3"),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "tag_condition.#", "3"),
				),
			},
			{
				Config: testAccAwsBackupSelectionConfig_tagsonly(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists("aws_backup_selection.tftest", &selection),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "name", fmt.Sprintf("tftest-%d", rInt)),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "resources.#", "0"),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "tag_condition.#", "3"),
				),
			},
			{
				Config: testAccAwsBackupSelectionConfig_resourcesonly(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists("aws_backup_selection.tftest", &selection),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "name", fmt.Sprintf("tftest-%d", rInt)),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "resources.#", "3"),
					resource.TestCheckResourceAttr("aws_backup_selection.tftest", "tag_condition.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAwsBackupSelectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).backupconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_selection" {
			continue
		}

		ids := strings.Split(rs.Primary.ID, "/")
		backupPlanId, selectionId := ids[0], ids[1]

		input := &backup.GetBackupSelectionInput{
			BackupPlanId: aws.String(backupPlanId),
			SelectionId:  aws.String(selectionId),
		}

		resp, err := conn.GetBackupSelection(input)

		if err == nil {
			if *resp.SelectionId == rs.Primary.ID {
				return fmt.Errorf("Backup selection '%s' from plan '%s' was not deleted properly", rs.Primary.ID, rs.Primary.Attributes["backup_plan_id"])
			}
		}
	}

	return nil
}

func testAccCheckAwsBackupSelectionExists(n string, selection *backup.Selection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Backup selection not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).backupconn

		ids := strings.Split(rs.Primary.ID, "/")
		backupPlanId, selectionId := ids[0], ids[1]

		input := &backup.GetBackupSelectionInput{
			BackupPlanId: aws.String(backupPlanId),
			SelectionId:  aws.String(selectionId),
		}

		resp, err := conn.GetBackupSelection(input)
		if err != nil {
			return err
		}

		*selection = *resp.BackupSelection

		return nil
	}
}

func testAccAwsBackupSelectionConfig_deps(rInt int) string {
	return fmt.Sprintf(`
locals {
  rand_int = "%d"
}

resource "aws_backup_vault" "tftest" {
  name = "tftest-${local.rand_int}"
}

resource "aws_backup_plan" "tftest" {
  name = "tftest-${local.rand_int}"

  rule {
    rule_name         = "tftest-${local.rand_int}"
    target_vault_name = "${aws_backup_vault.tftest.name}"
    schedule          = "cron(0 0 31 12 ? 2099)"
  }
}

data "aws_iam_policy_document" "tftest_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["backup.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "tftest" {
  name               = "tftest-${local.rand_int}"
  assume_role_policy = "${data.aws_iam_policy_document.tftest_assume.json}"
}

data "aws_iam_policy_document" "tftest" {
  statement {
    actions   = ["tag:GetResources"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "tftest" {
  role   = "${aws_iam_role.tftest.name}"
  policy = "${data.aws_iam_policy_document.tftest.json}"
}

data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
`, rInt)
}

func testAccAwsBackupSelectionConfig_create(rInt int) string {
	return testAccAwsBackupSelectionConfig_deps(rInt) + `
resource "aws_backup_selection" "tftest" {
  name           = "tftest-${local.rand_int}"
  backup_plan_id = "${aws_backup_plan.tftest.id}"
  iam_role_arn   = "${aws_iam_role.tftest.arn}"

  resources = [
    "arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/vol-1234",
    "arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/vol-2345",
  ]

  tag_condition {
    test     = "STRINGEQUALS"
    variable = "ec2:ResourceTag/tftest"
    value    = "tftest"
  }

  tag_condition {
    variable = "ec2:ResourceTag/tftest2"
    value    = "tftest2"
  }
}
`
}

func testAccAwsBackupSelectionConfig_update(rInt int) string {
	return testAccAwsBackupSelectionConfig_deps(rInt) + `
resource "aws_backup_selection" "tftest" {
  name           = "tftest-${local.rand_int}"
  backup_plan_id = "${aws_backup_plan.tftest.id}"
  iam_role_arn   = "${aws_iam_role.tftest.arn}"

  resources = [
    "arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/vol-1234",
    "arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/vol-2345",
	"arn:aws:rds:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:db:tftest",
  ]

  tag_condition {
    test     = "STRINGEQUALS"
    variable = "ec2:ResourceTag/tftest"
    value    = "tftest"
  }

  tag_condition {
    variable = "ec2:ResourceTag/tftest2"
    value    = "tftest2-update"
  }

  tag_condition {
	variable = "ec2:ResourceTag/tftest2"
	value    = "tftest2"
  }
}
`
}

func testAccAwsBackupSelectionConfig_tagsonly(rInt int) string {
	return testAccAwsBackupSelectionConfig_deps(rInt) + `
resource "aws_backup_selection" "tftest" {
  name           = "tftest-${local.rand_int}"
  backup_plan_id = "${aws_backup_plan.tftest.id}"
  iam_role_arn   = "${aws_iam_role.tftest.arn}"

  tag_condition {
    test     = "STRINGEQUALS"
    variable = "ec2:ResourceTag/tftest"
    value    = "tftest"
  }

  tag_condition {
    variable = "ec2:ResourceTag/tftest2"
    value    = "tftest2-update"
  }

  tag_condition {
	variable = "ec2:ResourceTag/tftest2"
	value    = "tftest2"
  }
}
`
}

func testAccAwsBackupSelectionConfig_resourcesonly(rInt int) string {
	return testAccAwsBackupSelectionConfig_deps(rInt) + `
resource "aws_backup_selection" "tftest" {
  name           = "tftest-${local.rand_int}"
  backup_plan_id = "${aws_backup_plan.tftest.id}"
  iam_role_arn   = "${aws_iam_role.tftest.arn}"

  resources = [
    "arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/vol-1234",
    "arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/vol-2345",
	"arn:aws:rds:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:db:tftest",
  ]
}
`
}
