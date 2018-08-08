package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDBClusterSnapshot_basic(t *testing.T) {
	var v rds.DBClusterSnapshot
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDbClusterSnapshotConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbClusterSnapshotExists("aws_db_cluster_snapshot.test", &v),
				),
			},
		},
	})
}

func testAccCheckDbClusterSnapshotExists(n string, v *rds.DBClusterSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		request := &rds.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(rs.Primary.ID),
		}

		response, err := conn.DescribeDBClusterSnapshots(request)
		if err == nil {
			if response.DBClusterSnapshots != nil && len(response.DBClusterSnapshots) > 0 {
				*v = *response.DBClusterSnapshots[0]
				return nil
			}
		}
		return fmt.Errorf("Error finding RDS DB Cluster Snapshot %s", rs.Primary.ID)
	}
}

func testAccAwsDbClusterSnapshotConfig(rInt int) string {
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
        Name = "resource_aws_db_cluster_snapshot_test"
    }
}

resource "aws_subnet" "aurora1" {
    vpc_id = "${aws_vpc.aurora.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
        Name = "resource_aws_db_cluster_snapshot_test"
    }
}

resource "aws_subnet" "aurora2" {
    vpc_id = "${aws_vpc.aurora.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
    tags {
        Name = "resource_aws_db_cluster_snapshot_test"
    }
}

resource "aws_db_subnet_group" "aurora" {
  subnet_ids = [
    "${aws_subnet.aurora1.id}",
    "${aws_subnet.aurora2.id}"
  ]
}

resource "aws_db_cluster_snapshot" "test" {
	db_cluster_identifier = "${aws_rds_cluster.aurora.id}"
	db_cluster_snapshot_identifier = "resourcetestsnapshot%d"
}`, rInt)
}
