package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func TestAccAwsEc2ClientVpnRoute_basic(t *testing.T) {
	var route1 ec2.ClientVpnRoute
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnRouteConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnRouteExists("aws_ec2_client_vpn_route.test", &route1),
				),
			},
		},
	})
}

func TestAccAwsEc2ClientVpnRoute_disappears(t *testing.T) {
	var route1 ec2.ClientVpnRoute
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnRouteConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnRouteExists("aws_ec2_client_vpn_route.test", &route1),
					testAccCheckAwsEc2ClientVpnRouteDisappears("aws_ec2_client_vpn_route.test"),
				),
			},
		},
	})
}

func testAccCheckAwsEc2ClientVpnRouteDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		_, err := conn.DeleteClientVpnRoute(&ec2.DeleteClientVpnRouteInput{
			ClientVpnEndpointId:  aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			DestinationCidrBlock: aws.String(rs.Primary.Attributes["destination_cidr_block"]),
			TargetVpcSubnetId:    aws.String(rs.Primary.Attributes["target_vpc_subnet_id"]),
		})

		return err
	}
}

func testAccCheckAwsEc2ClientVpnRouteDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_client_vpn_route" {
			continue
		}

		resp, _ := conn.DescribeClientVpnRoutes(&ec2.DescribeClientVpnRoutesInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
		})

		for _, r := range resp.Routes {
			if *r.ClientVpnEndpointId == rs.Primary.Attributes["client_vpn_endpoint_id"] && !(*r.Status.Code == ec2.ClientVpnRouteStatusCodeDeleting) {
				return fmt.Errorf("[DESTROY ERROR] client VPN route (%s) was not destroyed", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsEc2ClientVpnRouteExists(name string, route *ec2.ClientVpnRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeClientVpnRoutes(&ec2.DescribeClientVpnRoutesInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
		})

		if err != nil {
			return fmt.Errorf("Error reading client VPN route (%s): %s", rs.Primary.ID, err)
		}

		for _, r := range resp.Routes {
			if *r.ClientVpnEndpointId == rs.Primary.Attributes["client_vpn_endpoint_id"] && !(*r.Status.Code == ec2.ClientVpnRouteStatusCodeDeleting) {
				*route = *r
				return nil
			}
		}

		return fmt.Errorf("client VPN route route (%s) not found", rs.Primary.ID)
	}
}

func testAccEc2ClientVpnRouteConfigAcmCertificateBase() string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccEc2ClientVpnRouteConfig(rName string) string {
	return testAccEc2ClientVpnRouteConfigAcmCertificateBase() + fmt.Sprintf(`
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

resource "aws_ec2_client_vpn_route" "test" {
  client_vpn_endpoint_id = "${aws_ec2_client_vpn_endpoint.test.id}"
  destination_cidr_block = "0.0.0.0/0"
  target_vpc_subnet_id   = "${aws_ec2_client_vpn_network_association.test.subnet_id}"
  description            = "test client VPN route"
}
`, rName, rName, rName)
}
