package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	RegisterServiceErrorCheckFunc(transfer.EndpointsID, testAccErrorCheckSkipTransfer)

	resource.AddTestSweepers("aws_transfer_server", &resource.Sweeper{
		Name: "aws_transfer_server",
		F:    testSweepTransferServers,
	})
}

func testSweepTransferServers(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).transferconn
	input := &transfer.ListServersInput{}

	err = conn.ListServersPages(input, func(page *transfer.ListServersOutput, lastPage bool) bool {
		for _, server := range page.Servers {
			id := aws.StringValue(server.ServerId)
			input := &transfer.DeleteServerInput{
				ServerId: server.ServerId,
			}

			log.Printf("[INFO] Deleting Transfer Server: %s", id)
			_, err := conn.DeleteServer(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting Transfer Server (%s): %s", id, err)
				continue
			}

			if err := waitForTransferServerDeletion(conn, id); err != nil {
				log.Printf("[ERROR] Error waiting for Transfer Server (%s) deletion: %s", id, err)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Transfer Server sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Transfer Servers: %s", err)
	}

	return nil
}

func testAccErrorCheckSkipTransfer(t *testing.T) resource.ErrorCheckFunc {
	return testAccErrorCheckSkipMessagesContaining(t,
		"Invalid server type: PUBLIC",
	)
}

func TestAccAWSTransferServer_basic(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					testAccMatchResourceAttrRegionalHostname(resourceName, "endpoint", "server.transfer", regexp.MustCompile(`s-[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "logging_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "SERVICE_MANAGED"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_securityPolicy(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerSecurityPolicyConfig("TransferSecurityPolicy-2020-06"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2020-06"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerSecurityPolicyConfig("TransferSecurityPolicy-2018-11"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_Vpc(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_Vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerConfig_VpcUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_protocols_basic(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.foo"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		IDRefreshName: "aws_transfer_server.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_protocols_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestMatchResourceAttr(
						resourceName, "endpoint", regexp.MustCompile(fmt.Sprintf("^s-[a-z0-9]+.server.transfer.%s.amazonaws.com$", testAccGetRegion()))),
					resource.TestCheckResourceAttr(
						resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", transfer.ProtocolSftp),
				),
			},
			{
				ResourceName:            "aws_transfer_server.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerConfig_protocols_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", transfer.ProtocolFtp),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_apigateway(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_apigateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttrPair(resourceName, "invocation_role", "aws_iam_role.test", "arn"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_disappears(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsTransferServer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSTransferServer_forcedestroy(t *testing.T) {
	var conf transfer.DescribedServer
	var roleConf iam.GetRoleOutput
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_forcedestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
					testAccCheckAWSRoleExists("aws_iam_role.test", &roleConf),
					testAccCheckAWSTransferCreateUser(&conf, &roleConf, rName),
					testAccCheckAWSTransferCreateSshKey(&conf, rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func TestAccAWSTransferServer_vpcEndpointId(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	if testAccGetPartition() == "aws-us-gov" {
		t.Skip("Transfer Server VPC_ENDPOINT endpoint type is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_VpcEndPoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func TestAccAWSTransferServer_hostKey(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	hostKey := "test-fixtures/transfer-ssh-rsa-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_hostKey(hostKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "host_key_fingerprint", "SHA256:Z2pW9sPKDD/T34tVfCoolsRcECNTlekgaKvDn9t+9sg="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
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

func testAccCheckAWSTransferCreateUser(describedServer *transfer.DescribedServer, getRoleOutput *iam.GetRoleOutput, userName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).transferconn

		input := &transfer.CreateUserInput{
			ServerId: describedServer.ServerId,
			UserName: aws.String(userName),
			Role:     getRoleOutput.Role.Arn,
		}

		if _, err := conn.CreateUser(input); err != nil {
			return fmt.Errorf("error creating Transfer User (%s) on Server (%s): %s", userName, aws.StringValue(describedServer.ServerId), err)
		}

		return nil
	}
}

func testAccCheckAWSTransferCreateSshKey(describedServer *transfer.DescribedServer, userName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).transferconn

		input := &transfer.ImportSshPublicKeyInput{
			ServerId:         describedServer.ServerId,
			UserName:         aws.String(userName),
			SshPublicKeyBody: aws.String("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"),
		}

		if _, err := conn.ImportSshPublicKey(input); err != nil {
			return fmt.Errorf("error creating Transfer SSH Public Key for  (%s/%s): %s", userName, aws.StringValue(describedServer.ServerId), err)
		}

		return nil
	}
}

func testAccPreCheckAWSTransfer(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	input := &transfer.ListServersInput{}

	_, err := conn.ListServers(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSTransferServerBasicConfig() string {
	return `
resource "aws_transfer_server" "test" {}
`
}

func testAccAWSTransferServerSecurityPolicyConfig(policy string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  security_policy_name = %[1]q
}
`, policy)
}

func testAccAWSTransferServerUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoCloudWatchLogs",
    "Effect": "Allow",
    "Action": [
      "logs:*"
    ],
    "Resource": "*"
  }]
}
POLICY
}

resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role           = aws_iam_role.test.arn
}
`, rName)
}

func testAccAWSTransferServerConfig_apigateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_name        = "test"
  description       = %[1]q
  stage_description = %[1]q

  variables = {
    "a" = "2"
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoCloudWatchLogs",
    "Effect": "Allow",
    "Action": [
      "logs:*"
    ],
    "Resource": "*"
  }]
}
POLICY
}

resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "${aws_api_gateway_deployment.test.invoke_url}${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn
}
`, rName)
}

func testAccAWSTransferServerConfig_forcedestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoS3",
    "Effect": "Allow",
    "Action": [
      "s3:*"
    ],
    "Resource": "*"
  }]
}
POLICY
}
`, rName)
}

func testAccAWSTransferServerConfig_VpcEndPoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Allow TLS inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoint_service" "test" {
  service = "transfer.server"
}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  vpc_endpoint_type = "Interface"
  service_name      = data.aws_vpc_endpoint_service.test.service_name

  security_group_ids = [
    aws_security_group.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  tags = {
    Name = %[1]q
  }

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_transfer_server" "test" {
  endpoint_type = "VPC_ENDPOINT"

  endpoint_details {
    vpc_endpoint_id = aws_vpc_endpoint.test.id
  }
}
`, rName)
}

func testAccAWSTransferServerConfig_VpcDefault(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  count = 2

  vpc = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSTransferServerConfig_Vpc(rName string) string {
	return composeConfig(testAccAWSTransferServerConfig_VpcDefault(rName), `
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[0].id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`)
}

func testAccAWSTransferServerConfig_VpcUpdate(rName string) string {
	return composeConfig(testAccAWSTransferServerConfig_VpcDefault(rName), `
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[1].id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`)
}

func testAccAWSTransferServerConfig_hostKey(hostKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  host_key = file(%[1]q)
}
`, hostKey)
}

func testAccAWSTransferServerConfig_protocols_basic(rName string) string {

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
  role = aws_iam_role.foo.id

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
  logging_role           = aws_iam_role.foo.arn
  protocols              = ["SFTP"]

  tags = {
    NAME = "tf-acc-test-transfer-server"
    ENV  = "test"
  }
}
`, rName, rName)
}

func testAccAWSTransferServerConfig_protocols_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_name        = "test"
  description       = "%[1]s"
  stage_description = "%[1]s"

  variables = {
	"a" = "2"
  }
}

resource "aws_iam_role" "foo" {
  name = "tf-test-transfer-server-iam-role-for-apigateway-%[1]s"

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

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-igw"
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = [aws_internet_gateway.test]
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
	cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = "terraform-testacc-subnet"
  }
}

resource "aws_security_group" "test" {
  name   = "terraform-testacc-security-group"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-security-group"
  }
}


resource "aws_transfer_server" "foo" {
  identity_provider_type = "API_GATEWAY"
  url                    = "https://${aws_api_gateway_rest_api.test.id}.execute-api.${data.aws_region.current.name}.amazonaws.com${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.foo.arn
  logging_role           = aws_iam_role.foo.arn
  protocols              = ["FTP"]

  endpoint_type = "VPC"
  endpoint_details {
    subnet_ids = [aws_subnet.test.id]
    vpc_id     = aws_vpc.test.id
  }

  tags = {
    NAME = "tf-acc-test-transfer-server"
    TYPE = "apigateway"
  }
}
`, rName)

}
