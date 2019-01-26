package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsBackupPlan_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPlanConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists("aws_backup_plan.test"),
				),
			},
		},
	})
}

func testAccCheckAwsBackupPlanDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).backupconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault" {
			continue
		}

		input := &backup.DescribeBackupPlan\Input{
			BackupPlanId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetBackupPlan(input)
		if err != nil {
			return err
		}

		if !isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
			return fmt.Errorf("Plan '%s' was not deleted properly", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsBackupPlanExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s, %v", name, s.RootModule().Resources)
		}
		return nil
	}
}

func testAccBackupPlanConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_plan" "test" {
	name = "tf_acc_test_backup_plan_%d"

	rule {
		rule_name 			= "tf_acc_test_backup_rule_%d"
		target_backup_vault = "${aws_backup_vault.test.name}"

	}
}
`, randInt, randInt)
}
