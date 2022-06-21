package route53resolver_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
)

func TestAccRoute53ResolverDNSSECConfig_basic(t *testing.T) {
	resourceName := "aws_route53_resolver_dnssec_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDNSSECConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSSECConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSECConfigExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "route53resolver", regexp.MustCompile(`resolver-dnssec-config/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttr(resourceName, "validation_status", "ENABLED"),
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

func TestAccRoute53ResolverDNSSECConfig_disappear(t *testing.T) {
	resourceName := "aws_route53_resolver_dnssec_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDNSSECConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSSECConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSECConfigExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53resolver.ResourceDNSSECConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDNSSECConfigDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_dnssec_config" {
			continue
		}

		config, err := tfroute53resolver.FindResolverDNSSECConfigByID(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if config == nil {
			continue
		}

		return fmt.Errorf("Route 53 Resolver Dnssec config still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckDNSSECConfigExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver Dnssec config ID is set")
		}

		id := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

		config, err := tfroute53resolver.FindResolverDNSSECConfigByID(conn, id)

		if err != nil {
			return err
		}

		if config == nil {
			return fmt.Errorf("Route 53 Resolver Dnssec config (%s) not found", id)
		}

		if aws.StringValue(config.ValidationStatus) != route53resolver.ResolverDNSSECValidationStatusEnabled {
			return fmt.Errorf("Route 53 Resolver Dnssec config (%s) is not enabled", aws.StringValue(config.Id))
		}

		return nil
	}
}

func testAccDNSSECConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %q
  }
}
`, rName)
}

func testAccDNSSECConfigConfig_basic(rName string) string {
	return testAccDNSSECConfigBase(rName) + `
resource "aws_route53_resolver_dnssec_config" "test" {
  resource_id = aws_vpc.test.id
}
`
}
