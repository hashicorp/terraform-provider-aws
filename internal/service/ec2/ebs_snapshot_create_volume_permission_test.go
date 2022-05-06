package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2EBSSnapshotCreateVolumePermission_basic(t *testing.T) {
	var snapshotId string
	accountId := "111122223333"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccSnapshotCreateVolumePermissionDestroy,
		Steps: []resource.TestStep{
			// Scaffold everything
			{
				Config: testAccSnapshotCreateVolumePermissionConfig(true, accountId),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceGetAttr("aws_ebs_snapshot.test", "id", &snapshotId),
					testAccSnapshotCreateVolumePermissionExists(&accountId, &snapshotId),
				),
			},
			// Drop just create volume permission to test destruction
			{
				Config: testAccSnapshotCreateVolumePermissionConfig(false, accountId),
				Check: resource.ComposeTestCheckFunc(
					testAccSnapshotCreateVolumePermissionDestroyed(&accountId, &snapshotId),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCreateVolumePermission_disappears(t *testing.T) {
	var snapshotId string
	accountId := "111122223333"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccSnapshotCreateVolumePermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCreateVolumePermissionConfig(true, accountId),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceGetAttr("aws_ebs_snapshot.test", "id", &snapshotId),
					testAccSnapshotCreateVolumePermissionExists(&accountId, &snapshotId),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceSnapshotCreateVolumePermission(), "aws_snapshot_create_volume_permission.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckResourceGetAttr(name, key string, value *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s", name)
		}

		*value = is.Attributes[key]
		return nil
	}
}

func testAccSnapshotCreateVolumePermissionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_snapshot_create_volume_permission" {
			continue
		}

		snapshotID, accountID, err := tfec2.SnapshotCreateVolumePermissionParseID(rs.Primary.ID)
		if err != nil {
			return err
		}
		if has, err := tfec2.HasCreateVolumePermission(conn, snapshotID, accountID); err != nil {
			return err
		} else if has {
			return fmt.Errorf("create volume permission still exist for '%s' on '%s'", accountID, snapshotID)
		}
	}

	return nil
}

func testAccSnapshotCreateVolumePermissionExists(accountId, snapshotId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		if has, err := tfec2.HasCreateVolumePermission(conn, aws.StringValue(snapshotId), aws.StringValue(accountId)); err != nil {
			return err
		} else if !has {
			return fmt.Errorf("create volume permission does not exist for '%s' on '%s'", aws.StringValue(snapshotId), aws.StringValue(accountId))
		}
		return nil
	}
}

func testAccSnapshotCreateVolumePermissionDestroyed(accountId, snapshotId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		if has, err := tfec2.HasCreateVolumePermission(conn, aws.StringValue(snapshotId), aws.StringValue(accountId)); err != nil {
			return err
		} else if has {
			return fmt.Errorf("create volume permission still exists for '%s' on '%s'", aws.StringValue(snapshotId), aws.StringValue(accountId))
		}
		return nil
	}
}

func testAccSnapshotCreateVolumePermissionConfig(includeCreateVolumePermission bool, accountID string) string {
	base := `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = "ebs_snap_perm"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id
}
`

	if !includeCreateVolumePermission {
		return base
	}

	return base + fmt.Sprintf(`
resource "aws_snapshot_create_volume_permission" "test" {
  snapshot_id = aws_ebs_snapshot.test.id
  account_id  = %q
}
`, accountID)
}
