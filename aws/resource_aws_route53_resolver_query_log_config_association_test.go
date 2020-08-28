package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
)

func TestAccAWSRoute53ResolverQueryLogConfigAssociation_basic(t *testing.T) {
	var v route53resolver.ResolverQueryLogConfigAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_query_log_config_association.test"
	queryLogConfigResourceName := "aws_route53_resolver_query_log_config.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverQueryLogConfigAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverQueryLogConfigAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverQueryLogConfigAssociationExists(resourceName, &v),
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

func TestAccAWSRoute53ResolverQueryLogConfigAssociation_disappears(t *testing.T) {
	var v route53resolver.ResolverQueryLogConfigAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_query_log_config_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverQueryLogConfigAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverQueryLogConfigAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverQueryLogConfigAssociationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53ResolverQueryLogConfigAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRoute53ResolverQueryLogConfigAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53resolverconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_query_log_config_association" {
			continue
		}

		// Try to find the resource
		_, err := finder.ResolverQueryLogConfigAssociationByID(conn, rs.Primary.ID)
		// Verify the error is what we want
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver Query Log Config Association still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRoute53ResolverQueryLogConfigAssociationExists(n string, v *route53resolver.ResolverQueryLogConfigAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver Query Log Config Association ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).route53resolverconn
		out, err := finder.ResolverQueryLogConfigAssociationByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccRoute53ResolverQueryLogConfigAssociationConfig(rName string) string {
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
