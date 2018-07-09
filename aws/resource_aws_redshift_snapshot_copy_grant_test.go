package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAWSRedshiftSnapshotCopyGrant_Basic(t *testing.T) {
	timestamp := time.Now().Format(time.RFC1123)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotCopyGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotCopyGrant_Basic("basic"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotCopyGrantExists("aws_redshift_snapshot_copy_grant.basic"),
					resource.TestCheckResourceAttr("aws_redshift_snapshot_copy_grant.basic", "snapshot_copy_grant_name", "basic"),
					resource.TestCheckResourceAttrSet("aws_redshift_snapshot_copy_grant.basic", "kms_key_id"),
				),
			},
		},
	})
}

func testAccCheckAWSRedshiftSnapshotCopyGrantDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).redshiftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_snapshot_copy_grant" {
			continue
		}

		err := waitForAwsRedshiftSnapshotCopyGrantToBeDeleted(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func testAccCheckAWSRedshiftSnapshotCopyGrantExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		return nil
	}
}

func testAccAWSRedshiftSnapshotCopyGrant_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_copy_grant" "tf-acc-test-grant" {
    snapshot_copy_grant_name = "%s"
}
`, rName)
}
