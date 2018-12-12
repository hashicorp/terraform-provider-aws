package aws

import (
	"fmt"
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
					resource.TestMatchResourceAttr(
						"aws_transfer_server.foo", "endpoint", regexp.MustCompile(fmt.Sprintf("^s-[a-z0-9]+.server.transfer.%s.amazonaws.com$", testAccGetRegion()))),
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
					resource.TestCheckResourceAttrPair(
						"aws_transfer_server.foo", "logging_role", "aws_iam_role.foo", "arn"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_apigateway(t *testing.T) {
	var conf transfer.DescribedServer
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_transfer_server.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_apigateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttrSet(
						"aws_transfer_server.foo", "invocation_role"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.NAME", "tf-acc-test-transfer-server"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.TYPE", "apigateway"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_disappears(t *testing.T) {
	var conf transfer.DescribedServer

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					testAccCheckAWSTransferServerDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
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

func testAccCheckAWSTransferServerDisappears(conf *transfer.DescribedServer) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).transferconn

		params := &transfer.DeleteServerInput{
			ServerId: conf.ServerId,
		}

		_, err := conn.DeleteServer(params)
		if err != nil {
			return err
		}

		return waitForTransferServerDeletion(conn, *conf.ServerId)
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

		if err == nil {
			return fmt.Errorf("Transfer Server (%s) still exists", rs.Primary.ID)
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

func testAccAWSTransferServerConfig_apigateway(rName string) string {

	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
	name = "test"
}

resource "aws_api_gateway_resource" "test" {
	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	parent_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
	path_part = "test"
}

resource "aws_api_gateway_method" "test" {
	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	resource_id = "${aws_api_gateway_resource.test.id}"
	http_method = "GET"
	authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	resource_id = "${aws_api_gateway_resource.test.id}"
	http_method = "${aws_api_gateway_method.test.http_method}"
	status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	resource_id = "${aws_api_gateway_resource.test.id}"
	http_method = "${aws_api_gateway_method.test.http_method}"

	type = "HTTP"
	uri = "https://www.google.de"
	integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	resource_id = "${aws_api_gateway_resource.test.id}"
	http_method = "${aws_api_gateway_integration.test.http_method}"
	status_code = "${aws_api_gateway_method_response.error.status_code}"
}

resource "aws_api_gateway_deployment" "test" {
	depends_on = ["aws_api_gateway_integration.test"]

	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	stage_name = "test"
	description = "%s"
	stage_description = "%s"


  variables = {
    "a" = "2"
  }
}


resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-server-iam-role-for-apigateway-%s"

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
	identity_provider_type	= "API_GATEWAY"
	url 				   	= "https://${aws_api_gateway_rest_api.test.id}.execute-api.us-west-2.amazonaws.com${aws_api_gateway_resource.test.path}"
	invocation_role 	   	= "${aws_iam_role.foo.arn}"
	logging_role 		   	= "${aws_iam_role.foo.arn}"

	tags {
	  NAME     = "tf-acc-test-transfer-server"
	  TYPE	   = "apigateway"
	}
}
`, rName, rName, rName, rName)

}
