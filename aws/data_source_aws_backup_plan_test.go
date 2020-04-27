package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSBackupPlanDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_backup_plan.test"
	resourceName := "aws_backup_plan.test"
	planName := fmt.Sprintf("tf-testacc-backup-%s", acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsBackupPlanDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`Error getting Backup Plan`),
			},
			{
				Config: testAccAwsBackupPlanDataSourceConfig_basic(planName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanDataSourceID(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "version", resourceName, "version"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "rule.rule_name", resourceName, "rule.rule_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "rule.target_vault_name", resourceName, "rule.target_value_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "rule.schedule", resourceName, "rule.schedule"),
				),
			},
		},
	})
}

func TestAccAWSBackupPlanDataSource_withTags(t *testing.T) {
	datasourceName := "data.aws_backup_plan.test"
	resourceName := "aws_backup_plan.test"
	planName := fmt.Sprintf("tf-testacc-backup-%s", acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsBackupPlanDataSourceConfig_tags(planName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanDataSourceID(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key2", resourceName, "tags.Key2"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key3", resourceName, "tags.Key3")),
			},
		},
	})
}

func testAccCheckAwsBackupPlanDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Backup Plan data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Backup Plan data source ID not set")
		}
		return nil
	}
}

const testAccAwsBackupPlanDataSourceConfig_nonExistent = `
data "aws_backup_plan" "test" {
	plan_id = "tf-acc-test-does-not-exist"
}`

func testAccAwsBackupPlanDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
	resource "aws_backup_vault" "test" {
		name = %[1]q
	}

	resource "aws_backup_plan" "test" {
		name = %[1]q

		rule {
			rule_name         = %[1]q
			target_vault_name = "${aws_backup_vault.test.name}"
			schedule          = "cron(0 12 * * ? *)"
		}
	}
	`, name) + testAccBackupPlanDataSourceConfig
}

func testAccAwsBackupPlanDataSourceConfig_tags(name string) string {
	return fmt.Sprintf(`
	resource "aws_backup_vault" "test" {
		name = %[1]q
	}

	resource "aws_backup_plan" "test" {
		name = %[1]q

		rule {
			rule_name         = %[1]q
			target_vault_name = "${aws_backup_vault.test.name}"
			schedule          = "cron(0 12 * * ? *)"
		}

		tags = {
			Name = %[1]q
			Key2 = "Value2b"
			Key3 = "Value3"
		  }
	}
	`, name) + testAccBackupPlanDataSourceConfig
}

const testAccBackupPlanDataSourceConfig = `
data "aws_backup_plan" "test" {
	plan_id = aws_backup_plan.test.id
}`
