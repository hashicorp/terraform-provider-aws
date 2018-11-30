package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSTransferServer_basic(t *testing.T) {
	var conf transfer.DescribedServer

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_transfer_server.foo",
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					resource.TestCheckResourceAttrSet(
						"aws_transfer_server.foo", "arn"),
					resource.TestCheckResourceAttrSet(
						"aws_transfer_server.foo", "endpoint"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.NAME", "tf-acc-test-transfer-server"),
				),
			},
			{
				ResourceName:      "aws_transfer_server.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSTransferServerConfig_basicUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.NAME", "tf-acc-test-transfer-server"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.ENV", "test"),
					resource.TestCheckResourceAttrSet(
						"aws_transfer_server.foo", "logging_role"),
				),
			},
		},
	})
}

func testAccCheckAWSTransferServerExists(n string, res *transfer.DescribedServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Server ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).transferconn

		describe, err := conn.DescribeServer(&transfer.DescribeServerInput{
			ServerId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*res = *describe.Server

		if res.ServerId != nil {

		}

		return nil
	}
}

const testAccAWSTransferServerConfig_basic = `
resource "aws_transfer_server" "foo" {
  identity_provider_type = "SERVICE_MANAGED"

  tags {
	NAME     = "tf-acc-test-transfer-server"
  }
}
`

const testAccAWSTransferServerConfig_basicUpdate = `

resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-server-iam-role"
  
	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Effect": "Allow",
		"Principal": {
			"Service": "transfer.amazonaws.com"
		},
		"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_policy" "foo" {
	name 		= "tf-test-transfer-server-iam-policy"
	path        = "/"
	description = "IAM policy for for Transfer Server testing"
  
	policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": "s3:*",
			"Resource": "*"
		}
	]
}
EOF
}

resource "aws_iam_policy_attachment" "foo" {
	name       = "%s"
	roles      = ["${aws_iam_role.foo.name}"]
	policy_arn = "${aws_iam_policy.foo.arn}"
}
  

resource "aws_transfer_server" "foo" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role = "${aws_iam_role.foo.arn}"

  tags {
	NAME   = "tf-acc-test-transfer-server"
	ENV    = "test"
  }
}
`
