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
					resource.TestCheckResourceAttr("data.aws_msk_cluster.test_cluster", "name", sn),
					resource.TestCheckResourceAttrPair("data.aws_msk_cluster.test_cluster", "arn", "aws_msk_cluster.test_cluster", "arn"),
					resource.TestCheckResourceAttrPair("data.aws_msk_cluster.test_cluster", "zookeeper_connect", "aws_msk_cluster.test_cluster", "zookeeper_connect"),
					resource.TestCheckResourceAttrPair("data.aws_msk_cluster.test_cluster", "bootstrap_brokers", "aws_msk_cluster.test_cluster", "bootstrap_brokers"),
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

data "aws_availability_zones" "available" {}

resource "aws_subnet" "test_subnet_a" {
	vpc_id = "${aws_vpc.test_vpc.id}"
	cidr_block = "10.1.1.0/24"
	availability_zone = "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_subnet" "test_subnet_b" {
	vpc_id = "${aws_vpc.test_vpc.id}"
	cidr_block = "10.1.2.0/24"
	availability_zone = "${data.aws_availability_zones.available.names[1]}"
}

resource "aws_subnet" "test_subnet_c" {
	vpc_id = "${aws_vpc.test_vpc.id}"
	cidr_block = "10.1.3.0/24"
	availability_zone = "${data.aws_availability_zones.available.names[2]}"
}

resource "aws_security_group" "test_sg_a" {
	name        = "allow_all"
	description = "Allow all inbound traffic"
	vpc_id      = "${aws_vpc.test_vpc.id}"

	ingress {
		from_port   = 0
		to_port     = 0
		protocol    = "-1"
		cidr_blocks = ["0.0.0.0/0"]
	}
}

resource "aws_msk_cluster" "test_cluster" {
	name = "%s"
	broker_count = 3
	broker_instance_type = "kafka.m5.large"
	broker_volume_size = 10
	broker_security_groups =["${aws_security_group.test_sg_a.id}"]
	kafka_version = "1.1.1"
	client_subnets = ["${aws_subnet.test_subnet_a.id}", "${aws_subnet.test_subnet_b.id}", "${aws_subnet.test_subnet_c.id}"]
}

data "aws_msk_cluster" "test_cluster" {
	name = "${aws_msk_cluster.test_cluster.name}"
}
`
