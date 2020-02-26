package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
	"time"
)

func TestAccAwsEc2AuthorizeClientVpnIngress_basic(t *testing.T) {
	var assoc1 ec2.AuthorizationRule
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2RevokeClientVpnIngress,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2AuthorizeClientVpnConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2AuthorizeClientVpnIngressExists("aws_ec2_client_vpn_authorization.test", &assoc1),
				),
			},
		},
	})

}

func TestAccAwsEc2AuthorizeClientVpnIngress_disappears(t *testing.T) {
	var assoc1 ec2.AuthorizationRule
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2RevokeClientVpnIngress,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2AuthorizeClientVpnConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2AuthorizeClientVpnIngressExists("aws_ec2_client_vpn_authorization.test", &assoc1),
					testAccCheckAwsEc2AuthorizeClientVpnIngressDisappears(&assoc1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsEc2RevokeClientVpnIngress(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_client_vpn_route" {
			continue
		}

		req := &ec2.DescribeClientVpnRoutesInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("destinationCidr"),
					Values: []*string{aws.String(rs.Primary.Attributes["target_network_cidr"])},
				},
			},
		}
		resp, _ := conn.DescribeClientVpnRoutes(req)

		for _, a := range resp.Routes {
			if *a.DestinationCidr == rs.Primary.Attributes["target_network_cidr"] {
				return fmt.Errorf("[DESTROY ERROR] Client VPN authorization (%s) not deleted", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsEc2AuthorizeClientVpnIngressDisappears(authorizationRule *ec2.AuthorizationRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		req := &ec2.RevokeClientVpnIngressInput{
			ClientVpnEndpointId: authorizationRule.ClientVpnEndpointId,
			TargetNetworkCidr:   authorizationRule.DestinationCidr,
		}
		_, err := conn.RevokeClientVpnIngress(req)
		if err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{ec2.ClientVpnAuthorizationRuleStatusCodeRevoking},
			Target:  []string{"DELETED"},
			Refresh: clientVpnAuthorizationRefreshFunc(conn,
				aws.StringValue(authorizationRule.DestinationCidr),
				aws.StringValue(authorizationRule.ClientVpnEndpointId)),
			Timeout: 10 * time.Minute,
		}

		_, err = stateConf.WaitForState()

		return err
	}
}

func testAccCheckAwsEc2AuthorizeClientVpnIngressExists(name string, assoc *ec2.AuthorizationRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		req := &ec2.DescribeClientVpnAuthorizationRulesInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("destinationCidr"),
					Values: []*string{aws.String(rs.Primary.Attributes["target_network_cidr"])},
				},
			},
		}

		resp, err := conn.DescribeClientVpnAuthorizationRules(req)
		if err != nil {
			return fmt.Errorf("Error reading Client VPN authorization (%s): %s", rs.Primary.ID, err)
		}

		for _, a := range resp.AuthorizationRules {
			if *a.DestinationCidr == rs.Primary.Attributes["target_network_cidr"] {
				*assoc = *a
				return nil
			}
		}

		return fmt.Errorf("Client VPN authorization (%s) not found", rs.Primary.ID)
	}
}

func testAccEc2AuthorizeClientVpnConfigAcmCertificateBase() string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccEc2AuthorizeClientVpnConfig(rName string) string {
	return testAccEc2AuthorizeClientVpnConfigAcmCertificateBase() + fmt.Sprintf(`

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-%s"
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = "${aws_vpc.test.id}"
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-subnet-%s"
  }
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.test.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.test.arn}"
  }

  connection_log_options {
    enabled = false
  }
}

resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = "${aws_ec2_client_vpn_endpoint.test.id}"
  subnet_id              = "${aws_subnet.test.id}"
}

resource "aws_ec2_client_vpn_authorization" "test" {
  description               = "terraform testing"
  client_vpn_endpoint_id    = "${aws_ec2_client_vpn_endpoint.test.id}"
  target_network_cidr       = "10.200.0.0/16"
  authorize_all_groups      = true
}
`, rName, rName, rName)
}
