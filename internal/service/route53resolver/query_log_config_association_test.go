package route53resolver_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
)

func TestAccRoute53ResolverQueryLogConfigAssociation_basic(t *testing.T) {
	var v route53resolver.ResolverQueryLogConfigAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_query_log_config_association.test"
	queryLogConfigResourceName := "aws_route53_resolver_query_log_config.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryLogConfigAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfigAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryLogConfigAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_query_log_config_id", queryLogConfigResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", vpcResourceName, "id"),
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

func TestAccRoute53ResolverQueryLogConfigAssociation_disappears(t *testing.T) {
	var v route53resolver.ResolverQueryLogConfigAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_query_log_config_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryLogConfigAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfigAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryLogConfigAssociationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53resolver.ResourceQueryLogConfigAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQueryLogConfigAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_query_log_config_association" {
			continue
		}

		// Try to find the resource
		_, err := tfroute53resolver.FindResolverQueryLogConfigAssociationByID(conn, rs.Primary.ID)
		// Verify the error is what we want
		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver Query Log Config Association still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckQueryLogConfigAssociationExists(n string, v *route53resolver.ResolverQueryLogConfigAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver Query Log Config Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
		out, err := tfroute53resolver.FindResolverQueryLogConfigAssociationByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccQueryLogConfigAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_resolver_query_log_config" "test" {
  name            = %[1]q
  destination_arn = aws_cloudwatch_log_group.test.arn
}

resource "aws_route53_resolver_query_log_config_association" "test" {
  resolver_query_log_config_id = aws_route53_resolver_query_log_config.test.id
  resource_id                  = aws_vpc.test.id
}
`, rName)
}
