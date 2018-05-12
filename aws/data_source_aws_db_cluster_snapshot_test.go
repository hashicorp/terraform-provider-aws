package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDbClusterSnapshotDataSource_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDbClusterSnapshotDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDbClusterSnapshotDataSourceID("data.aws_db_cluster_snapshot.snapshot"),
				),
			},
		},
	})
}

func testAccCheckAwsDbClusterSnapshotDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot data source ID not set")
		}
		return nil
	}
}

func testAccCheckAwsDbClusterSnapshotDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "aurora" {
  master_username         = "foo"
  master_password         = "barbarbarbar"
  db_subnet_group_name = "${aws_db_subnet_group.aurora.name}"
  backup_retention_period = 1
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "aurora" {
  count=1
  cluster_identifier = "${aws_rds_cluster.aurora.id}"
  instance_class = "db.t2.small"
  db_subnet_group_name = "${aws_db_subnet_group.aurora.name}"
}

resource "aws_vpc" "aurora" {
    cidr_block = "192.168.0.0/16"
    tags {
        Name = "data_source_aws_db_cluster_snapshot_test"
    }
}

resource "aws_subnet" "aurora1" {
    vpc_id = "${aws_vpc.aurora.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
        Name = "data_source_aws_db_cluster_snapshot_test"
    }
}

resource "aws_subnet" "aurora2" {
    vpc_id = "${aws_vpc.aurora.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
    tags {
        Name = "data_source_aws_db_cluster_snapshot_test"
    }
}

resource "aws_db_subnet_group" "aurora" {
  subnet_ids = [
    "${aws_subnet.aurora1.id}",
    "${aws_subnet.aurora2.id}"
  ]
}

data "aws_db_cluster_snapshot" "snapshot" {
	most_recent = "true"
	db_cluster_snapshot_identifier = "${aws_db_cluster_snapshot.test.id}"
}


resource "aws_db_cluster_snapshot" "test" {
	db_cluster_identifier = "${aws_rds_cluster.aurora.id}"
	db_cluster_snapshot_identifier = "testsnapshot%d"
}`, rInt)
}
