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

func TestAccAwsClientVpnNetworkAssociation_basic(t *testing.T) {
	rStr := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsClientVpnNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVpnNetworkAssociationConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsClientVpnNetworkAssociationExists("aws_client_vpn_network_association.test"),
				),
			},

			{
				ResourceName:      "aws_client_vpn_network_association.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsClientVpnNetworkAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_client_vpn_network_association" {
			continue
		}

		resp, _ := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			AssociationIds:      []*string{aws.String(rs.Primary.ID)},
		})

		for _, v := range resp.ClientVpnTargetNetworks {
			if *v.AssociationId == rs.Primary.ID && !(*v.Status.Code == ec2.AssociationStatus.Code) {
				return fmt.Errorf("[DESTROY ERROR] Client VPN network association (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsClientVpnNetworkAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccClientVpnNetworkAssociationConfig(rName string) string {
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
