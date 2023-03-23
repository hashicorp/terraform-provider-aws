package rds_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRDSSnapshotCopy_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBSnapshot
	resourceName := "aws_db_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(resourceName, &v),
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

func TestAccRDSSnapshotCopy_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBSnapshot
	resourceName := "aws_db_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyTagsConfig1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotCopyTagsConfig2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSnapshotCopyTagsConfig1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRDSSnapshotCopy_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBSnapshot
	resourceName := "aws_db_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfrds.ResourceSnapshotCopy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotCopyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_snapshot_copy" {
			continue
		}

		log.Printf("[DEBUG] Checking if RDS DB Snapshot %s exists", rs.Primary.ID)

		_, err := tfrds.FindSnapshot(context.Background(), conn, rs.Primary.ID)

		// verify error is what we want
		if tfresource.NotFound(err) {
			continue
		}

		return err
	}

	return nil
}

func testAccCheckSnapshotCopyExists(n string, ci *rds.DBSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no RDS DB Snapshot ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		out, err := tfrds.FindSnapshot(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		ci = out

		return nil
	}
}

func testAccSnapshotCopyBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]
}

resource "aws_db_instance" "test" {
  allocated_storage       = 10
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  name                    = "baz"
  identifier              = %[1]q
  password                = "barbarbarbar"
  username                = "foo"
  maintenance_window      = "Fri:09:00-Fri:09:30"
  backup_retention_period = 0
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.id
  db_snapshot_identifier = "%[1]s-source"
}`, rName)
}

func testAccSnapshotCopyConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotCopyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_db_snapshot_copy" "test" {
  source_db_snapshot_identifier = aws_db_snapshot.test.db_snapshot_arn
  target_db_snapshot_identifier = "%[1]s-target"
}`, rName))
}

func testAccSnapshotCopyTagsConfig1(rName, tagKey, tagValue string) string {
	return acctest.ConfigCompose(
		testAccSnapshotCopyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_db_snapshot_copy" "test" {
  source_db_snapshot_identifier = aws_db_snapshot.test.db_snapshot_arn
  target_db_snapshot_identifier = "%[1]s-target"

  tags = {
    %[2]q = %[3]q
  }
}`, rName, tagKey, tagValue))
}

func testAccSnapshotCopyTagsConfig2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccSnapshotCopyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_db_snapshot_copy" "test" {
  source_db_snapshot_identifier = aws_db_snapshot.test.db_snapshot_arn
  target_db_snapshot_identifier = "%[1]s-target"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
