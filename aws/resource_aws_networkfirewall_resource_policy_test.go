package aws

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/networkfirewall/finder"
)

func TestAccAwsNetworkFirewallResourcePolicy_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckHasIAMRole(t, "AWSPrivatePreviewRoleForVPCFirewall") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallResourcePolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallResourcePolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_networkfirewall_firewall.test", "arn"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("^{\"Version\":\"2012-10-17\".+")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
func TestAccAwsNetworkFirewallResourcePolicy_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckHasIAMRole(t, "AWSPrivatePreviewRoleForVPCFirewall") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallResourcePolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallResourcePolicyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsNetworkFirewallResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsNetworkFirewallResourcePolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkfirewall_resource_policy" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn
		policy, err := finder.ResourcePolicy(context.Background(), conn, rs.Primary.ID)
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		if policy != nil {
			return fmt.Errorf("NetworkFirewall Resource Policy (for resource: %s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsNetworkFirewallResourcePolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No NetworkFirewall Resource Policy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn
		policy, err := finder.ResourcePolicy(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if policy == nil {
			return fmt.Errorf("NetworkFirewall Resource Policy (for resource: %s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccNetworkFirewallResourcePolicy_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name = %[1]q
  }
}
resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id
  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_networkfirewall_resource_policy" "test" {
  resource_arn = aws_networkfirewall_firewall.test.arn
  policy       = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": [
        "ec2:AttachNetworkInterface",
        "ec2:CreateNetworkInterface",
        "ec2:CreateNetworkInterfacePermission",
        "ec2:DeleteNetworkInterface",
        "ec2:DeleteNetworkInterfacePermission",
        "ec2:DescribeInstances",
        "ec2:DescribeNetworkInterfaceAttribute",
        "ec2:DescribeNetworkInterfacePermissions",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DescribeSubnets",
        "ec2:DescribeVpcs",
        "ec2:DetachNetworkInterface",
        "ec2:ModifyNetworkInterfaceAttribute"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}
`, rName)
}
