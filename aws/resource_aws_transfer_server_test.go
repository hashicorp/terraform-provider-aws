package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSTransferServer_basic(t *testing.T) {
	var conf transfer.DescribedServer
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_transfer_server.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					testAccMatchResourceAttrRegionalARN("aws_transfer_server.foo", "arn", "transfer", regexp.MustCompile(`server/.+`)),
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
				Config: testAccAWSTransferServerConfig_basicUpdate(rName),
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

		return nil
	}
}

		}

		return nil
	}
}

func testAccCheckAWSTransferServerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_server" {
			continue
		}

		_, err := conn.DescribeServer(&transfer.DescribeServerInput{
			ServerId: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, transfer.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

const testAccAWSTransferServerConfig_basic = `
resource "aws_transfer_server" "foo" {
  identity_provider_type = "SERVICE_MANAGED"

  tags {
	NAME     = "tf-acc-test-transfer-server"
  }
}
`

func testAccAWSTransferServerConfig_basicUpdate(rName string) string {

	return fmt.Sprintf(`

resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-server-iam-role-%s"
  
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

resource "aws_iam_role_policy" "foo" {
	name = "tf-test-transfer-server-iam-policy-%s"
	role = "${aws_iam_role.foo.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Sid": "AllowFullAccesstoCloudWatchLogs",
		"Effect": "Allow",
		"Action": [
			"logs:*"
		],
		"Resource": "*"
		}
	]
}
POLICY
}

resource "aws_transfer_server" "foo" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role = "${aws_iam_role.foo.arn}"

  tags {
	NAME   = "tf-acc-test-transfer-server"
	ENV    = "test"
  }
}
`, rName, rName)
}
