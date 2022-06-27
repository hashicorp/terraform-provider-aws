package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRDSSnapshotDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rInt := sdkacctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotIDDataSource("data.aws_db_snapshot.snapshot"),
				),
			},
		},
	})
}

func testAccCheckSnapshotIDDataSource(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Volume data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot data source ID not set")
		}
		return nil
	}
}

func testAccSnapshotDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = [%[1]s]
}

resource "aws_db_instance" "bar" {
  allocated_storage   = 10
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  name                = "baz"
  password            = "barbarbarbar"
  username            = "foo"
  skip_final_snapshot = true

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"

  backup_retention_period = 0

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
}

data "aws_db_snapshot" "snapshot" {
  most_recent            = "true"
  db_snapshot_identifier = aws_db_snapshot.test.id
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.bar.id
  db_snapshot_identifier = "testsnapshot%[2]d"
}
`, mySQLPreferredInstanceClasses, rInt)
}
