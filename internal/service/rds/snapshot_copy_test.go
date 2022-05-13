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

func TestAccSnapshotCopy_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBSnapshot
	resourceName := "aws_db_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdsDbSnapshotCopyExists(resourceName, &v),
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

func TestAccAWSRdsDbSnapshotCopy_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBSnapshot
	resourceName := "aws_db_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdsDbSnapshotCopyExists(resourceName, &v),
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
		if rs.Type != "aws_rds_db_snapshot_copy" {
			continue
		}

		log.Printf("[DEBUG] Checking if RDS DB Snapshot %s exists", rs.Primary.ID)

		_, err := tfrds.FindDBSnapshot(context.Background(), conn, rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("the RDS DB Snapshot %s still exists. Failing", rs.Primary.ID)
		}

		// verify error is what we want
		if tfresource.NotFound(err) {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckRdsDbSnapshotCopyExists(n string, ci *rds.DBSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no RDS DB Snapshot ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		out, err := tfrds.FindDBSnapshot(context.Background(), conn, rs.Primary.ID)
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

//func testAccAwsRdsDbSnapshotCopyConfigWithRegions(rInt int) string {
//	return fmt.Sprintf(`provider "aws" {
//	region = "us-west-2"
//	alias  = "uswest2"
//}
//
//provider "aws" {
//	region = "us-west-1"
//	alias  = "uswest1"
//}
//
//resource "aws_db_instance" "bar" {
//	provider          = "aws.uswest2"
//	allocated_storage = 10
//	engine = "MySQL"
//	engine_version = "5.6.35"
//	instance_class = "db.t2.micro"
//	name = "baz"
//	password = "barbarbarbar"
//	username = "foo"
//
//	maintenance_window = "Fri:09:00-Fri:09:30"
//
//	backup_retention_period = 0
//
//	parameter_group_name = "default.mysql5.6"
//
//	skip_final_snapshot = true
//}
//
//resource "aws_db_snapshot" "test" {
//  provider = "aws.uswest2"
//  db_instance_identifier = aws_db_instance.bar.id
//  db_snapshot_identifier = "testsnapshot%d"
//}
//
//resource "aws_db_snapshot_copy" "test" {
//	provider           = "aws.uswest1"
//	source_db_snapshot_identifier = aws_db_snapshot.test.db_snapshot_arn
//	target_db_snapshot_identifier = "testsnapshot%d"
//	source_region = "us-west-2"
//}`, rInt, rInt)
//}
