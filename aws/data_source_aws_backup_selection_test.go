package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSBackupSelectionDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_backup_selection.test"
	resourceName := "aws_backup_selection.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsBackupSelectionDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`Error getting Backup Selection`),
			},
			{
				Config: testAccAwsBackupSelectionDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionDataSourceID(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "iam_role_arn", resourceName, "iam_role_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "selection_tag", resourceName, "selection_tag"),
					resource.TestCheckResourceAttrPair(datasourceName, "resources", resourceName, "resources"),
				),
			},
		},
	})
}

func testAccCheckAwsBackupSelectionDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Backup Selection data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Backup Selection data source ID not set")
		}
		return nil
	}
}

const testAccAwsBackupSelectionDataSourceConfig_nonExistent = `
data "aws_backup_selection" "test" {
	plan_id      = "tf-acc-test-does-not-exist"
	selection_id = "tf-acc-test-dne"
}`

func testAccAwsBackupSelectionDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"
}

resource "aws_backup_plan" "test" {
  name = "tf_acc_test_backup_plan_%d"

  rule {
    rule_name         = "tf_acc_test_backup_rule_%d"
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 12 * * ? *)"
  }
}

resource "aws_backup_selection" "test" {
  plan_id      = "${aws_backup_plan.test.id}"
  name         = "tf_acc_test_backup_selection_%d"
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }

  resources = [
    "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/"
  ]
}
`, rInt, rInt, rInt, rInt) + testAccBackupSelectionDataSourceConfig
}

const testAccBackupSelectionDataSourceConfig = `
data "aws_backup_selection" "test" {
	plan_id      = aws_backup_plan.test.id
	selection_id = aws_backup_selection.test.id
}`
