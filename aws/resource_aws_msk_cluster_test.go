package aws

import (
	"fmt"
	"strconv"
	"strings"
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
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testMskClusterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists("aws_msk_cluster.test_cluster", &cluster),
					testAccCheckAWSMskClusterAttributes(&cluster),
				),
			},
		},
	})
}

func testMskClusterConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_msk_cluster" "test_cluster" {
	name = "terraform-msk-test-%d"
	broker_count = 1
	tags = {
		Name = "tf-test"
	}
}`, rInt)
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

func testAccCheckAWSMskClusterAttributes(cluster *kafka.ClusterInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*cluster.ClusterName, "terraform-msk-test") {
			return fmt.Errorf("Bad Cluster name: %s", *cluster.ClusterName)
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_cluster" {
				continue
			}
			if *cluster.ClusterArn != rs.Primary.Attributes["arn"] {
				return fmt.Errorf("Bad Cluster ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["arn"], *cluster.ClusterArn)
			}
			broker_count := strconv.Itoa(int(aws.Int64Value(cluster.NumberOfBrokerNodes)))
			if broker_count != rs.Primary.Attributes["broker_count"] {
				return fmt.Errorf("Bad Cluster Broker Count\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["broker_count"], broker_count)
			}
		}
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

		return nil

	}

	return nil
}
