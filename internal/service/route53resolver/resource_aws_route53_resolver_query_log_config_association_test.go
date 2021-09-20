package route53resolver_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
)

func init() {
	resource.AddTestSweepers("aws_route53_resolver_query_log_config_association", &resource.Sweeper{
		Name: "aws_route53_resolver_query_log_config_association",
		F:    testSweepRoute53ResolverQueryLogConfigAssociations,
	})
}

func testSweepRoute53ResolverQueryLogConfigAssociations(region string) error {
	client, err := acctest.SharedRegionalSweeperClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn
	var sweeperErrs *multierror.Error

	err = conn.ListResolverQueryLogConfigAssociationsPages(&route53resolver.ListResolverQueryLogConfigAssociationsInput{}, func(page *route53resolver.ListResolverQueryLogConfigAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queryLogConfigAssociation := range page.ResolverQueryLogConfigAssociations {
			id := aws.StringValue(queryLogConfigAssociation.Id)

			log.Printf("[INFO] Deleting Route53 Resolver Query Log Config Association: %s", id)
			r := ResourceQueryLogConfigAssociation()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})
	if acctest.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Query Log Config Associations sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver Query Log Config Associations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSRoute53ResolverQueryLogConfigAssociation_basic(t *testing.T) {
	var v route53resolver.ResolverQueryLogConfigAssociation
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_query_log_config_association.test"
	queryLogConfigResourceName := "aws_route53_resolver_query_log_config.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
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
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_query_log_config_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverQueryLogConfigAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverQueryLogConfigAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverQueryLogConfigAssociationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceQueryLogConfigAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRoute53ResolverQueryLogConfigAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_query_log_config_association" {
			continue
		}

		// Try to find the resource
		_, err := tfroute53resolver.FindResolverQueryLogConfigAssociationByID(conn, rs.Primary.ID)
		// Verify the error is what we want
		if tfawserr.ErrMessageContains(err, route53resolver.ErrCodeResourceNotFoundException, "") {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
		out, err := tfroute53resolver.FindResolverQueryLogConfigAssociationByID(conn, rs.Primary.ID)
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
