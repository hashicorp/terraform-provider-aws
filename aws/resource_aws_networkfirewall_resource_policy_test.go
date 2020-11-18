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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/networkfirewall/finder"
)

func TestAccAwsNetworkFirewallResourcePolicy_firewallPolicy(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsNetworkFirewallResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallResourcePolicy_firewallPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallResourcePolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_iam_user.test", "arn"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`\"Action\":[\"network\-firewall:ListFirewallPolicies\"]`)),
				),
			},
			{
				Config: testAccNetworkFirewallResourcePolicy_firewallPolicy_updatePolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallResourcePolicyExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`\"Action\":[\"network\-firewall:ListFirewallPolicies\", \"network\-firewall:AssociateFirewallPolicy\"]`)),
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

func TestAccAwsNetworkFirewallResourcePolicy_ruleGroup(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallResourcePolicy_ruleGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallResourcePolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_iam_user.test", "arn"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`\"Action\":[\"network\-firewall:ListRuleGroups\"]`)),
				),
			},
			{
				Config: testAccNetworkFirewallResourcePolicy_ruleGroup_updatePolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallResourcePolicyExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`\"Action\":[\"network\-firewall:ListRuleGroups\", \"network\-firewall:CreateFirewallPolicy\"]`)),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallResourcePolicy_firewallPolicy(rName),
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

func testAccNetworkFirewallResourcePolicyFirewallPolicyBaseConfig(rName string) string {
	return composeConfig(
		testAccAlternateAccountProviderConfig(),
		fmt.Sprintf(`
data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_networkfirewall_firewall_policy.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}
`, rName))
}

func testAccNetworkFirewallResourcePolicy_firewallPolicy(rName string) string {
	return composeConfig(
		testAccNetworkFirewallResourcePolicyFirewallPolicyBaseConfig(rName), `
resource "aws_networkfirewall_resource_policy" "test" {
  resource_arn = data.aws_caller_identity.alternate.arn
  policy       = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "network-firewall:ListFirewallPolicies",
      "Resource": "${aws_networkfirewall_firewall_policy.test.arn}"
    }
  ]
}
POLICY

  depends_on = [aws_ram_resource_association.test]
}
`)
}

func testAccNetworkFirewallResourcePolicy_firewallPolicy_updatePolicy(rName string) string {
	return composeConfig(
		testAccNetworkFirewallResourcePolicyFirewallPolicyBaseConfig(rName), `
resource "aws_networkfirewall_resource_policy" "test" {
  resource_arn = data.aws_caller_identity.alternate.arn
  policy       = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": [
        "network-firewall:ListFirewallPolicies",
        "network-firewall:AssociateFirewallPolicy"
      ],
      "Resource": "${aws_networkfirewall_firewall_policy.test.arn}"
    }
  ]
}
POLICY
}
  depends_on = [aws_ram_resource_association.test]
`)
}

func testAccNetworkFirewallResourcePolicyRuleGroupBaseConfig(rName string) string {
	return composeConfig(
		testAccAlternateAccountProviderConfig(),
		fmt.Sprintf(`
data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"
  rule_group {
    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_networkfirewall_rule_group.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}
`, rName))
}

func testAccNetworkFirewallResourcePolicy_ruleGroup(rName string) string {
	return composeConfig(
		testAccNetworkFirewallResourcePolicyRuleGroupBaseConfig(rName), `
resource "aws_networkfirewall_resource_policy" "test" {
  resource_arn = data.aws_caller_identity.alternate.arn
  policy       = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "network-firewall:ListRuleGroups",
      "Resource": "${aws_networkfirewall_rule_group.test.arn}"
    }
  ]
}
POLICY

  depends_on = [aws_ram_resource_association.test]
}
`)
}

func testAccNetworkFirewallResourcePolicy_ruleGroup_updatePolicy(rName string) string {
	return composeConfig(
		testAccNetworkFirewallResourcePolicyRuleGroupBaseConfig(rName), `
resource "aws_networkfirewall_resource_policy" "test" {
  resource_arn = data.aws_caller_identity.alternate.arn
  policy       = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": [
        "network-firewall:ListRuleGroups",
        "network-firewall:CreateFirewallPolicy"
      ],
      "Resource": "${aws_networkfirewall_rule_group.test.arn}"
    }
  ]
}
POLICY

  depends_on = [aws_ram_resource_association.test]
}
`)
}
