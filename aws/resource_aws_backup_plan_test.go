package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsBackupPlan_basic(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPlanConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					testAccMatchResourceAttrRegionalARN("aws_backup_plan.test", "arn", "backup", regexp.MustCompile(`backup-plan:.+`)),
					resource.TestCheckResourceAttrSet("aws_backup_plan.test", "version"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_withTags(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPlanWithTag(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "tags.env", "test"),
				),
			},
			{
				Config: testAccBackupPlanWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "tags.env", "test"),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "tags.app", "widget"),
				),
			},
			{
				Config: testAccBackupPlanWithTag(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "tags.env", "test"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_withRules(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPlanWithRules(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "rule.#", "2"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_withRuleRemove(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPlanWithRules(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "rule.#", "2"),
				),
			},
			{
				Config: testAccBackupPlanConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "rule.#", "1"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_withRuleAdd(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPlanConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "rule.#", "1"),
				),
			},
			{
				Config: testAccBackupPlanWithRules(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "rule.#", "2"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_withLifecycle(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	rStr := "lifecycle_policy"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPlanWithLifecycle(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					resource.TestCheckResourceAttr("aws_backup_plan.test", "rule.1028372010.lifecycle.#", "1"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_disappears(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPlanConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test", &plan),
					testAccCheckAwsBackupPlanDisappears(&plan),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsBackupPlanDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).backupconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_plan" {
			continue
		}

		input := &backup.GetBackupPlanInput{
			BackupPlanId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetBackupPlan(input)

		if err == nil {
			if *resp.BackupPlanId == rs.Primary.ID {
				return fmt.Errorf("Plane '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsBackupPlanDisappears(backupPlan *backup.GetBackupPlanOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).backupconn

		input := &backup.DeleteBackupPlanInput{
			BackupPlanId: backupPlan.BackupPlanId,
		}

		_, err := conn.DeleteBackupPlan(input)

		return err
	}
}

func testAccCheckAwsBackupPlanExists(name string, plan *backup.GetBackupPlanOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID is not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).backupconn

		input := &backup.GetBackupPlanInput{
			BackupPlanId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetBackupPlan(input)

		if err != nil {
			return err
		}

		*plan = *output

		return nil
	}
}

func testAccBackupPlanConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"
}

resource "aws_backup_plan" "test" {
  name = "tf_acc_test_backup_plan_%d"

  rule {
    rule_name          = "tf_acc_test_backup_rule_%d"
    target_vault_name  = "${aws_backup_vault.test.name}"
    schedule           = "cron(0 12 * * ? *)"
  }
}
`, randInt, randInt, randInt)
}

func testAccBackupPlanWithTag(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"
}

resource "aws_backup_plan" "test" {
  name = "tf_acc_test_backup_plan_%d"

  rule {
    rule_name          = "tf_acc_test_backup_rule_%d"
    target_vault_name  = "${aws_backup_vault.test.name}"
    schedule           = "cron(0 12 * * ? *)"
  }

  tags {
	  env = "test"
  }
}
`, randInt, randInt, randInt)
}

func testAccBackupPlanWithTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"
}

resource "aws_backup_plan" "test" {
  name = "tf_acc_test_backup_plan_%d"

  rule {
    rule_name          = "tf_acc_test_backup_rule_%d"
    target_vault_name  = "${aws_backup_vault.test.name}"
    schedule           = "cron(0 12 * * ? *)"
  }

  tags {
	  env = "test"
	  app = "widget"
  }
}
`, randInt, randInt, randInt)
}

func testAccBackupPlanWithLifecycle(stringID string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%s"
}

resource "aws_backup_plan" "test" {
  name = "tf_acc_test_backup_plan_%s"

  rule {
    rule_name          = "tf_acc_test_backup_rule_%s"
    target_vault_name  = "${aws_backup_vault.test.name}"
    schedule           = "cron(0 12 * * ? *)"
    lifecycle {
        cold_storage_after = 30
        delete_after       = 160
    }
  }
}
`, stringID, stringID, stringID)
}

func testAccBackupPlanWithRules(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"
}

resource "aws_backup_plan" "test" {
  name = "tf_acc_test_backup_plan_%d"

  rule {
    rule_name          = "tf_acc_test_backup_rule_%d"
    target_vault_name  = "${aws_backup_vault.test.name}"
    schedule           = "cron(0 12 * * ? *)"
  }

  rule {
    rule_name          = "tf_acc_test_backup_rule_%d_2"
    target_vault_name  = "${aws_backup_vault.test.name}"
    schedule           = "cron(0 6 * * ? *)"
  }
}
`, randInt, randInt, randInt, randInt)
}
