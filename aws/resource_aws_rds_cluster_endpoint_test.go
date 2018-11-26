package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSRDSClusterEndpoint_basic(t *testing.T) {
	rInt := acctest.RandInt()
	readerResourceName := "aws_rds_cluster_endpoint.reader"
	defaultResourceName := "aws_rds_cluster_endpoint.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterEndpointConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(readerResourceName, "arn", regexp.MustCompile(`^arn:[^:]+:rds:[^:]+:\d{12}:cluster-endpoint:.+`)),
					resource.TestCheckResourceAttrSet(readerResourceName, "endpoint"),
					resource.TestMatchResourceAttr(defaultResourceName, "arn", regexp.MustCompile(`^arn:[^:]+:rds:[^:]+:\d{12}:cluster-endpoint:.+`)),
					resource.TestCheckResourceAttrSet(defaultResourceName, "endpoint"),
				),
			},
			{
				ResourceName:      "aws_rds_cluster_endpoint.reader",
				ImportState:       true,
				ImportStateVerify: true,
			},

			{
				ResourceName:      "aws_rds_cluster_endpoint.default",
				ImportState:       true,
				ImportStateVerify: true,
			},

		},
	})
}

func testAccAWSClusterEndpointConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test1" {
  apply_immediately = true
  cluster_identifier = "${aws_rds_cluster.default.id}"
  identifier = "tf-aurora-cluster-instance-test1-%d"
  instance_class = "db.t2.small"
}

resource "aws_rds_cluster_instance" "test2" {
  apply_immediately = true
  cluster_identifier = "${aws_rds_cluster.default.id}"
  identifier = "tf-aurora-cluster-instance-test2-%d"
  instance_class = "db.t2.small"
}

resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier = "${aws_rds_cluster.default.id}"
  cluster_endpoint_identifier = "reader-%d"
  custom_endpoint_type = "READER"

  static_members = ["${aws_rds_cluster_instance.test2.id}"]
}

resource "aws_rds_cluster_endpoint" "default" {
  cluster_identifier = "${aws_rds_cluster.default.id}"
  cluster_endpoint_identifier = "default-%d"
  custom_endpoint_type = "ANY"

  excluded_members = ["${aws_rds_cluster_instance.test2.id}"]
}
`, n, n, n, n, n)
}
