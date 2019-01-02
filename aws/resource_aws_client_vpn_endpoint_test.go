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

func TestAccAwsClientVpnEndpoint_importBasic(t *testing.T) {
	resourceName := "aws_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVpnEndpointConfig(acctest.RandString(5)),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsClientVpnEndpoint_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVpnEndpointConfig(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsClientVpnEndpointExists("aws_client_vpn_endpoint.test"),
				),
			},
		},
	})
}

func TestAccAwsClientVpnEndpoint_withLogGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVpnEndpointConfigWithLogGroup(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsClientVpnEndpointExists("aws_client_vpn_endpoint.test"),
				),
			},
		},
	})
}

func TestAccAwsClientVpnEndpoint_withDNSServers(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVpnEndpointConfigWithDNSServers(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsClientVpnEndpointExists("aws_client_vpn_endpoint.test"),
				),
			},
		},
	})
}

func testAccCheckAwsClientVpnEndpointDestroy(s *terraform.State) error {
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

func testAccCheckAwsClientVpnEndpointExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccClientVpnEndpointConfig(rName string) string {
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

resource "aws_client_vpn_endpoint" "test" {
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

func testAccClientVpnEndpointConfigWithLogGroup(rName string) string {
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

resource "aws_client_vpn_endpoint" "test" {
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
  }
}
`, rName, rName)
}

func testAccClientVpnEndpointConfigWithDNSServers(rName string) string {
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

resource "aws_client_vpn_endpoint" "test" {
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
