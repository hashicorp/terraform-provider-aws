package route53resolver_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRoute53ResolverFirewallConfig_basic(t *testing.T) {
	var v route53resolver.FirewallConfig
	resourceName := "aws_route53_resolver_firewall_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallConfigExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "firewall_fail_open", "ENABLED"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccRoute53ResolverFirewallConfig_disappears(t *testing.T) {
	var v route53resolver.FirewallConfig
	resourceName := "aws_route53_resolver_firewall_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallConfigExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53resolver.ResourceFirewallConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFirewallConfigDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_firewall_config" {
			continue
		}

		config, err := tfroute53resolver.FindFirewallConfigByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		if aws.StringValue(config.FirewallFailOpen) == route53resolver.FirewallFailOpenStatusDisabled {
			return nil
		}

		return fmt.Errorf("Route 53 Resolver DNS Firewall config still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckFirewallConfigExists(n string, v *route53resolver.FirewallConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver DNS Firewall config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

		out, err := tfroute53resolver.FindFirewallConfigByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccFirewallConfigConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_resolver_firewall_config" "test" {
  resource_id        = aws_vpc.test.id
  firewall_fail_open = "ENABLED"
}
`, rName)
}
