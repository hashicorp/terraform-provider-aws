package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDaxCluster_basic(t *testing.T) {
	var dc dax.Cluster
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDaxClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDaxClusterConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxClusterExists("aws_dax_cluster.test", &dc),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "nodes.0.id", "tf-test-dax-cluster-a"),
					resource.TestCheckResourceAttrSet(
						"aws_dax_cluster.test", "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(
						"aws_dax_cluster.test", "cluster_address"),
				),
			},
		},
	})
}

func TestAccAwsDaxCluster_resize(t *testing.T) {
	var dc dax.Cluster
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDaxClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDaxClusterConfigResize_singleNode,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxClusterExists("aws_dax_cluster.test", &dc),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "replication_factor", "1"),
				),
			},
			{
				Config: testAccAwsDaxClusterConfigResize_multiNode,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxClusterExists("aws_dax_cluster.test", &dc),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "replication_factor", "2"),
				),
			},
			{
				Config: testAccAwsDaxClusterConfigResize_singleNode,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxClusterExists("aws_dax_cluster.test", &dc),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "replication_factor", "1"),
				),
			},
		},
	})
}

func testAccCheckAwsDaxClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).daxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dax_cluster" {
			continue
		}
		res, err := conn.DescribeClusters(&dax.DescribeClustersInput{
			ClusterNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			// Verify the error is what we want
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ClusterNotFoundFault" {
				continue
			}
			return err
		}
		if len(res.Clusters) > 0 {
			return fmt.Errorf("still exist.")
		}
	}
	return nil
}

func testAccCheckAwsDaxClusterExists(n string, v *dax.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DAX cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).daxconn
		resp, err := conn.DescribeClusters(&dax.DescribeClustersInput{
			ClusterNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return fmt.Errorf("DAX error: %v", err)
		}

		for _, c := range resp.Clusters {
			if *c.ClusterName == rs.Primary.ID {
				*v = *c
			}
		}

		return nil
	}
}

var baseConfig = `
provider "aws" {
  region = "us-west-2"
}

data "aws_caller_identity" "test" {}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags {
    Name = "tf-test"
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
       {
            "Effect": "Allow",
            "Principal": {
                "Service": "dax.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = "${aws_iam_role.test.id}"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "dynamodb:*"
            ],
            "Effect": "Allow",
            "Resource": [
                "arn:aws:dynamodb:us-west-2:${data.aws_caller_identity.test.account_id}:*"
            ]
        }
    ]
}
EOF
}

`

var testAccAwsDaxClusterConfig = fmt.Sprintf(`%s
resource "aws_dax_cluster" "test" {
  cluster_name       = "tf-test-dax-cluster"
  iam_role_arn       = "${aws_iam_role.test.arn}"
  node_type          = "dax.r3.large"
  replication_factor = 1

  tags {
    foo = "bar"
  }
}
`, baseConfig)

var testAccAwsDaxClusterConfigResize_singleNode = fmt.Sprintf(`%s
resource "aws_dax_cluster" "test" {
  cluster_name       = "tf-test-dax-cluster"
  iam_role_arn       = "${aws_iam_role.test.arn}"
  node_type          = "dax.r3.large"
  replication_factor = 1
}
`, baseConfig)

var testAccAwsDaxClusterConfigResize_multiNode = fmt.Sprintf(`%s
resource "aws_dax_cluster" "test" {
  cluster_name       = "tf-test-dax-cluster"
  iam_role_arn       = "${aws_iam_role.test.arn}"
  node_type          = "dax.r3.large"
  replication_factor = 2
}
`, baseConfig)
