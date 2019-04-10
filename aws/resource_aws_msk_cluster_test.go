package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSMskCluster_basic(t *testing.T) {
	var cluster kafka.ClusterInfo

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testMskClusterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists("aws_msk_cluster.test_cluster", &cluster),
					testAccMatchResourceAttrRegionalARN("aws_msk_cluster.test_cluster", "arn", "kafka", regexp.MustCompile(`cluster/.+`)),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "broker_count", "3"),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "broker_volume_size", "10"),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "broker_security_groups.#", "1"),
					resource.TestCheckResourceAttrSet("aws_msk_cluster.test_cluster", "zookeeper_connect"),
					resource.TestCheckResourceAttrSet("aws_msk_cluster.test_cluster", "bootstrap_brokers"),
				),
			},

			{
				ResourceName:      "aws_msk_cluster.test_cluster",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSMskCluster_encryptAtRest(t *testing.T) {
	var cluster kafka.ClusterInfo

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testMskClusterConfigWithEncryption(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists("aws_msk_cluster.test_cluster", &cluster),
					testAccMatchResourceAttrRegionalARN("aws_msk_cluster.test_cluster", "arn", "kafka", regexp.MustCompile(`cluster/.+`)),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "broker_count", "3"),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "broker_volume_size", "10"),
					resource.TestCheckResourceAttrSet("aws_msk_cluster.test_cluster", "zookeeper_connect"),
					resource.TestCheckResourceAttrSet("aws_msk_cluster.test_cluster", "bootstrap_brokers"),
				),
			},

			{
				ResourceName:      "aws_msk_cluster.test_cluster",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSMskCluster_brokerMonitoring(t *testing.T) {
	var cluster kafka.ClusterInfo

	rInt := acctest.RandInt()

	monitoring_type := "PER_BROKER"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testMskClusterConfigBrokerMonitoring(rInt, monitoring_type),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists("aws_msk_cluster.test_cluster", &cluster),
					testAccMatchResourceAttrRegionalARN("aws_msk_cluster.test_cluster", "arn", "kafka", regexp.MustCompile(`cluster/.+`)),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "broker_count", "3"),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "broker_volume_size", "10"),
					resource.TestCheckResourceAttrSet("aws_msk_cluster.test_cluster", "zookeeper_connect"),
					resource.TestCheckResourceAttrSet("aws_msk_cluster.test_cluster", "bootstrap_brokers"),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "enhanced_monitoring", "PER_BROKER"),
				),
			},

			{
				ResourceName:      "aws_msk_cluster.test_cluster",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSMskCluster_topicMonitoring(t *testing.T) {
	var cluster kafka.ClusterInfo

	rInt := acctest.RandInt()

	monitoring_type := "PER_TOPIC_PER_BROKER"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testMskClusterConfigBrokerMonitoring(rInt, monitoring_type),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists("aws_msk_cluster.test_cluster", &cluster),
					testAccMatchResourceAttrRegionalARN("aws_msk_cluster.test_cluster", "arn", "kafka", regexp.MustCompile(`cluster/.+`)),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "broker_count", "3"),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "broker_volume_size", "10"),
					resource.TestCheckResourceAttrSet("aws_msk_cluster.test_cluster", "zookeeper_connect"),
					resource.TestCheckResourceAttrSet("aws_msk_cluster.test_cluster", "bootstrap_brokers"),
					resource.TestCheckResourceAttr("aws_msk_cluster.test_cluster", "enhanced_monitoring", monitoring_type),
				),
			},

			{
				ResourceName:      "aws_msk_cluster.test_cluster",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testMskClusterCommonConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test_vpc" {
	cidr_block = "10.1.0.0/16"
	tags {
		Name = "test_vpc-%d"
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

`, rInt)
}

func testMskClusterConfig(rInt int) string {
	return testMskClusterCommonConfig(rInt) + fmt.Sprintf(`

resource "aws_msk_cluster" "test_cluster" {
	name = "terraform-msk-test-%d"
	broker_count = 3
	broker_instance_type = "kafka.m5.large"
	broker_volume_size = 10
	kafka_version = "1.1.1"
	broker_security_groups =["${aws_security_group.test_sg_a.id}"]
	client_subnets = ["${aws_subnet.test_subnet_a.id}", "${aws_subnet.test_subnet_b.id}", "${aws_subnet.test_subnet_c.id}"]
}`, rInt)
}

func testMskClusterConfigWithEncryption(rInt int) string {
	return testMskClusterCommonConfig(rInt) + fmt.Sprintf(`

resource "aws_kms_key" "test_key" {
	description             = "KMS key 1"
	deletion_window_in_days = 10
}

resource "aws_msk_cluster" "test_cluster" {
	name = "terraform-msk-test-%d"
	broker_count = 3
	broker_instance_type = "kafka.m5.large"
	broker_volume_size = 10
	broker_security_groups =["${aws_security_group.test_sg_a.id}"]
	kafka_version = "1.1.1"
	client_subnets = ["${aws_subnet.test_subnet_a.id}", "${aws_subnet.test_subnet_b.id}", "${aws_subnet.test_subnet_c.id}"]
	encryption_key = "${aws_kms_key.test_key.arn}"
}`, rInt)
}

func testMskClusterConfigBrokerMonitoring(rInt int, monitoring_type string) string {
	return testMskClusterCommonConfig(rInt) + fmt.Sprintf(`

resource "aws_msk_cluster" "test_cluster" {
	name = "terraform-msk-test-%d"
	broker_count = 3
	broker_instance_type = "kafka.m5.large"
	broker_security_groups =["${aws_security_group.test_sg_a.id}"]
	broker_volume_size = 10
	kafka_version = "1.1.1"
	client_subnets = ["${aws_subnet.test_subnet_a.id}", "${aws_subnet.test_subnet_b.id}", "${aws_subnet.test_subnet_c.id}"]
	enhanced_monitoring = "%s"
}`, rInt, monitoring_type)
}

func testAccCheckMskClusterExists(n string, cluster *kafka.ClusterInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kafkaconn
		describeOpts := &kafka.DescribeClusterInput{
			ClusterArn: aws.String(rs.Primary.Attributes["arn"]),
		}
		resp, err := conn.DescribeCluster(describeOpts)
		if err != nil {
			return err
		}

		*cluster = *resp.ClusterInfo

		return nil
	}
}

func testAccCheckMskClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_msk_cluster" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).kafkaconn
		describeOpts := &kafka.DescribeClusterInput{
			ClusterArn: aws.String(rs.Primary.Attributes["arn"]),
		}
		resp, err := conn.DescribeCluster(describeOpts)
		if err == nil {
			if resp.ClusterInfo != nil && *resp.ClusterInfo.State != "DELETING" {
				return fmt.Errorf("Error: Cluster still exists")
			}
		}
	}

	return nil
}
