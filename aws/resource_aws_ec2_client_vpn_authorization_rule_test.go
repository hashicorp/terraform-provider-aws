package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
)

func TestAccAwsEc2ClientVpnAuthorizationRule_basic(t *testing.T) {
	var v ec2.AuthorizationRule
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_authorization_rule.test"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnAuthorizationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnAuthorizationRuleConfigBasic(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "target_network_cidr", subnetResourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "authorize_all_groups", "true"),
				),
			},
		},
	})
}

func TestAccAwsEc2ClientVpnAuthorizationRule_disappears(t *testing.T) {
	var v ec2.AuthorizationRule
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_authorization_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnAuthorizationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnAuthorizationRuleConfigBasic(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ClientVpnAuthorizationRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsEc2ClientVpnAuthorizationRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_client_vpn_authorization_rule" {
			continue
		}

		endpointID, _ /*targetNetworkCidr*/, _ /*accessGroupID*/, err := tfec2.ClientVpnAuthorizationRuleParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &ec2.DescribeClientVpnAuthorizationRulesInput{
			ClientVpnEndpointId: aws.String(endpointID),
			// TODO: filters
		}

		_, err = conn.DescribeClientVpnAuthorizationRules(input)

		if err == nil {
			return fmt.Errorf("Client VPN authorization rule (%s) still exists", rs.Primary.ID)
		}
		if isAWSErr(err, tfec2.ErrCodeClientVpnEndpointAuthorizationRuleNotFound, "") || isAWSErr(err, errCodeClientVpnEndpointIdNotFound, "") {
			continue
		}
		return err
	}

	return nil
}

func testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(name string, assoc *ec2.AuthorizationRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DescribeClientVpnAuthorizationRulesInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
		}
		result, err := conn.DescribeClientVpnAuthorizationRules(input)

		if err != nil {
			return fmt.Errorf("error reading Client VPN authorization rule (%s): %w", rs.Primary.ID, err)
		}

		if result != nil || len(result.AuthorizationRules) == 1 || result.AuthorizationRules[0] != nil {
			*assoc = *result.AuthorizationRules[0]
			return nil
		}

		return fmt.Errorf("Client VPN network association (%s) not found", rs.Primary.ID)
	}
}

func testAccEc2ClientVpnAuthorizationRuleConfigBasic(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # InvalidParameterValue: AZ us-west-2d is not currently supported. Please choose another az in this region
  blacklisted_zone_ids = ["usw2-az4"]
  state                = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-%[1]s"
  }
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = "${aws_vpc.test.id}"
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-subnet-%[1]s"
  }
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
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

resource "aws_ec2_client_vpn_authorization_rule" "test" {
  client_vpn_endpoint_id = "${aws_ec2_client_vpn_endpoint.test.id}"
  target_network_cidr    = "${aws_subnet.test.cidr_block}"
  authorize_all_groups   = true
}
`, rName)
}
