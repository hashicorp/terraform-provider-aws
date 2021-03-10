package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEFSFileSystemBackupPolicy_basic(t *testing.T) {
	var out efs.DescribeBackupPolicyOutput
	resourceName := "aws_efs_file_system_backup_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemBackupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemBackupPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEFSFileSystemBackupPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "backup_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "backup_policy.0.status", "ENABLED"),
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

func TestAccAWSEFSFileSystemBackupPolicy_disappears(t *testing.T) {
	var out efs.DescribeBackupPolicyOutput
	resourceName := "aws_efs_file_system_backup_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemBackupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemBackupPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEFSFileSystemBackupPolicyExists(resourceName, &out),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEfsFileSystemBackupPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEFSFileSystemBackupPolicyExists(name string, bp *efs.DescribeBackupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).efsconn
		fs, err := conn.DescribeBackupPolicy(&efs.DescribeBackupPolicyInput{
			FileSystemId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*bp = *fs

		return nil
	}
}

func testAccCheckEfsFileSystemBackupPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).efsconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_file_system_backup_policy" {
			continue
		}

		resp, err := conn.DescribeBackupPolicy(&efs.DescribeBackupPolicyInput{
			FileSystemId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, efs.ErrCodeFileSystemNotFound, "") ||
				isAWSErr(err, efs.ErrCodePolicyNotFound, "") {
				return nil
			}
			return fmt.Errorf("error describing EFS file system backup policy in tests: %s", err)
		}

		if resp == nil || resp.BackupPolicy == nil || aws.StringValue(resp.BackupPolicy.Status) == efs.StatusDisabled {
			return nil
		}

		return fmt.Errorf("EFS file system backup policy %q still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSEFSFileSystemBackupPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %q
}

resource "aws_efs_file_system_backup_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  backup_policy {
    status = "ENABLED"
  }
}
`, rName)
}
