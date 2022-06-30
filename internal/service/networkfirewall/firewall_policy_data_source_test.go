package networkfirewall_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkfirewall"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccFirewallPolicy_arn(t *testing.T) {
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSource_arn(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.#", resourceName, "firewall_policy.0.stateless_fragment_default_actions.#"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.*", resourceName, "firewall_policy.0.stateless_fragment_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.#", resourceName, "firewall_policy.0.stateless_default_actions.#"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.*", resourceName, "firewall_policy.0.stateless_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccFirewallPolicy_name(t *testing.T) {
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSource_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.#", resourceName, "firewall_policy.0.stateless_fragment_default_actions.#"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.*", resourceName, "firewall_policy.0.stateless_fragment_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.#", resourceName, "firewall_policy.0.stateless_default_actions.#"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.*", resourceName, "firewall_policy.0.stateless_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccFirewallPolicy_nameAndArn(t *testing.T) {
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSource_nameAndArn(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.#", resourceName, "firewall_policy.0.stateless_fragment_default_actions.#"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.*", resourceName, "firewall_policy.0.stateless_fragment_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.#", resourceName, "firewall_policy.0.stateless_default_actions.#"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.*", resourceName, "firewall_policy.0.stateless_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccFirewallPolicyDataSource_basic(rName string) string {
	return fmt.Sprintf(`

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}
`, rName)
}

func testAccFirewallPolicyDataSource_arn(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyDataSource_basic(rName),
		`
data "aws_networkfirewall_firewall_policy" "test" {
  arn = aws_networkfirewall_firewall_policy.test.arn
}`)
}

func testAccFirewallPolicyDataSource_name(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyDataSource_basic(rName),
		`
data "aws_networkfirewall_firewall_policy" "test" {
  name = aws_networkfirewall_firewall_policy.test.name
}`)
}

func testAccFirewallPolicyDataSource_nameAndArn(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyDataSource_basic(rName),
		`
data "aws_networkfirewall_firewall_policy" "test" {
  arn  = aws_networkfirewall_firewall_policy.test.arn
  name = aws_networkfirewall_firewall_policy.test.name
}`)
}
