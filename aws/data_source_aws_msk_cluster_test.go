package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSMskClusterDataSource(t *testing.T) {
	var cluster kafka.ClusterInfo

	sn := fmt.Sprintf("terraform-msk-test-%d", acctest.RandInt())
	config := fmt.Sprintf(testAccCheckMskDataSourceConfig, sn, sn)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists("aws_msk_cluster.test_cluster", &cluster),
					resource.TestCheckResourceAttrSet("data.aws_msk_cluster.test_cluster", "arn"),
					resource.TestCheckResourceAttr("data.aws_msk_cluster.test_cluster", "name", sn),
					resource.TestCheckResourceAttrSet("data.aws_msk_cluster.test_cluster", "zookeeper_connect"),
					resource.TestCheckResourceAttrSet("data.aws_msk_cluster.test_cluster", "bootstrap_brokers"),
				),
			},
		},
	})
}

var testAccCheckMskDataSourceConfig = `
resource "aws_vpc" "test_vpc" {
	cidr_block = "10.1.0.0/16"
	tags {
		Name = "test_vpc-%s"
	}
}
	
resource "aws_subnet" "test_subnet_a" {
	vpc_id = "${aws_vpc.test_vpc.id}"
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-east-1a"
}

resource "aws_subnet" "test_subnet_b" {
	vpc_id = "${aws_vpc.test_vpc.id}"
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-east-1b"
}

resource "aws_subnet" "test_subnet_c" {
	vpc_id = "${aws_vpc.test_vpc.id}"
	cidr_block = "10.1.3.0/24"
	availability_zone = "us-east-1c"
}

resource "aws_msk_cluster" "test_cluster" {
	name = "%s"
	broker_count = 3
	broker_instance_type = "kafka.m5.large"
	broker_volume_size = 10
	enhanced_monitoring = "DEFAULT"
	client_subnets = ["${aws_subnet.test_subnet_a.id}", "${aws_subnet.test_subnet_b.id}", "${aws_subnet.test_subnet_c.id}"]
}

data "aws_msk_cluster" "test_cluster" {
	name = "${aws_msk_cluster.test_cluster.name}"
}
`
