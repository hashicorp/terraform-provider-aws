package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
)

func TestAccRedshiftSnapshotCopyGrant_basic(t *testing.T) {
	resourceName := "aws_redshift_snapshot_copy_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotCopyGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyGrantConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_copy_grant_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
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

func TestAccRedshiftSnapshotCopyGrant_update(t *testing.T) {
	resourceName := "aws_redshift_snapshot_copy_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotCopyGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyGrantConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Env", "Test"),
				),
			},
			{
				Config: testAccSnapshotCopyGrantConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccSnapshotCopyGrantConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Env", "Test"),
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

func TestAccRedshiftSnapshotCopyGrant_disappears(t *testing.T) {
	resourceName := "aws_redshift_snapshot_copy_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotCopyGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyGrantConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyGrantExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceSnapshotCopyGrant(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotCopyGrantDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_snapshot_copy_grant" {
			continue
		}

		err := tfredshift.WaitForSnapshotCopyGrantToBeDeleted(conn, rs.Primary.ID)
		return err
	}

	return nil
}

func testAccCheckSnapshotCopyGrantExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot Copy Grant ID (SnapshotCopyGrantName) is not set")
		}

		// retrieve the client from the test provider
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		input := redshift.DescribeSnapshotCopyGrantsInput{
			MaxRecords:            aws.Int64(100),
			SnapshotCopyGrantName: aws.String(rs.Primary.ID),
		}

		response, err := conn.DescribeSnapshotCopyGrants(&input)

		if err != nil {
			return err
		}

		// we expect only a single snapshot copy grant by this ID. If we find zero, or many,
		// then we consider this an error
		if len(response.SnapshotCopyGrants) != 1 ||
			*response.SnapshotCopyGrants[0].SnapshotCopyGrantName != rs.Primary.ID {
			return fmt.Errorf("Snapshot copy grant not found")
		}

		return nil
	}
}

func testAccSnapshotCopyGrantConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_copy_grant" "test" {
  snapshot_copy_grant_name = %[1]q
}
`, rName)
}

func testAccSnapshotCopyGrantConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_copy_grant" "test" {
  snapshot_copy_grant_name = %[1]q

  tags = {
    Name = %[1]q
    Env  = "Test"
  }
}
`, rName)
}
