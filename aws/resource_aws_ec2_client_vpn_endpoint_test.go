package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsEc2ClientVpnEndpoint_basic(t *testing.T) {
	rStr := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
				),
			},

			{
				ResourceName:      "aws_ec2_client_vpn_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_withLogGroup(t *testing.T) {
	rStr := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
				),
			},

			{
				Config: testAccEc2ClientVpnEndpointConfigWithLogGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
				),
			},

			{
				ResourceName:      "aws_ec2_client_vpn_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_withDNSServers(t *testing.T) {
	rStr := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
				),
			},

			{
				Config: testAccEc2ClientVpnEndpointConfigWithDNSServers(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
				),
			},

			{
				ResourceName:      "aws_ec2_client_vpn_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsEc2ClientVpnEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_client_vpn_endpoint" {
			continue
		}

		input := &ec2.DescribeClientVpnEndpointsInput{
			ClientVpnEndpointIds: []*string{aws.String(rs.Primary.ID)},
		}

		resp, _ := conn.DescribeClientVpnEndpoints(input)
		for _, v := range resp.ClientVpnEndpoints {
			if *v.ClientVpnEndpointId == rs.Primary.ID && !(*v.Status.Code == ec2.ClientVpnEndpointStatusCodeDeleted) {
				return fmt.Errorf("[DESTROY ERROR] Client VPN endpoint (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsEc2ClientVpnEndpointExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccEc2ClientVpnEndpointConfig(rName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block = "10.0.0.0/16"

  authentication_options {
    type = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithLogGroup(rName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_cloudwatch_log_group" "lg" {
  name = "terraform-testacc-clientvpn-loggroup-%s"
}

resource "aws_cloudwatch_log_stream" "ls" {
  name           = "${aws_cloudwatch_log_group.lg.name}-stream"
  log_group_name = "${aws_cloudwatch_log_group.lg.name}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block = "10.0.0.0/16"

  authentication_options {
    type = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = true
    cloudwatch_log_group = "${aws_cloudwatch_log_group.lg.name}"
    cloudwatch_log_stream = "${aws_cloudwatch_log_stream.ls.name}"
  }
}
`, rName, rName)
}

func testAccEc2ClientVpnEndpointConfigWithDNSServers(rName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block = "10.0.0.0/16"

  dns_servers = ["8.8.8.8", "8.8.4.4"]

  authentication_options {
    type = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}
