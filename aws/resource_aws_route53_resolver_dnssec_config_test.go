package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_route53_resolver_dnssec_config", &resource.Sweeper{
		Name: "aws_route53_resolver_dnssec_config",
		F:    testSweepRoute53ResolverDnssecConfig,
	})
}

func testSweepRoute53ResolverDnssecConfig(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).route53resolverconn

	var errors error
	err = conn.ListResolverDnssecConfigsPages(&route53resolver.ListResolverDnssecConfigsInput{}, func(page *route53resolver.ListResolverDnssecConfigsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, resolverDnssecConfig := range page.ResolverDnssecConfigs {
			id := aws.StringValue(resolverDnssecConfig.ResourceId)

			log.Printf("[INFO] Deleting Route53 Resolver Dnssec config: %s", id)
			_, err := conn.UpdateResolverDnssecConfig(&route53resolver.UpdateResolverDnssecConfigInput{
				ResourceId: aws.String(id),
				Validation: aws.String(route53resolver.ResolverDNSSECValidationStatusDisabled),
			})
			if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
				continue
			}
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error deleting Route53 Resolver Resolver Dnssec config (%s): %w", id, err))
				continue
			}

			err = route53ResolverEndpointWaitUntilTargetState(conn, id, 10*time.Minute,
				[]string{route53resolver.ResolverDNSSECValidationStatusDisabling},
				[]string{route53resolver.ResolverDNSSECValidationStatusDisabled})
			if err != nil {
				errors = multierror.Append(errors, err)
				continue
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver Resolver Dnssec config sweep for %s: %s", region, err)
			return nil
		}
		errors = multierror.Append(errors, fmt.Errorf("error retrieving Route53 Resolver Resolver Dnssec config: %w", err))
	}

	return errors
}

func TestAccAWSRoute53ResolverDnssecConfig_basic(t *testing.T) {
	var config route53resolver.ResolverDnssecConfig
	resourceName := "aws_route53_resolver_dnssec_config.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverDnssecConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverDnssecConfigConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverDnssecConfigExists(resourceName, &config),
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

func TestAccAWSRoute53ResolverDnssecConfig_disappear(t *testing.T) {
	var config route53resolver.ResolverDnssecConfig
	resourceName := "aws_route53_resolver_dnssec_config.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverDnssecConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverDnssecConfigConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverDnssecConfigExists(resourceName, &config),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53ResolverDnssecConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ResolverDnssecConfig_disappear_VPC(t *testing.T) {
	var config route53resolver.ResolverDnssecConfig
	resourceName := "aws_route53_resolver_dnssec_config.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverDnssecConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverDnssecConfigConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverDnssecConfigExists(resourceName, &config),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpc(), vpcResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRoute53ResolverDnssecConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53resolverconn
	ec2Conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_dnssec_config" {
			continue
		}

		vpc, err := vpcDescribe(ec2Conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		// The VPC has been deleted
		if vpc == nil {
			continue
		}

		// Try to find the resource
		out, err := conn.GetResolverDnssecConfig(&route53resolver.GetResolverDnssecConfigInput{
			ResourceId: aws.String(rs.Primary.ID),
		})
		// Verify the error is what we want
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		if aws.StringValue(out.ResolverDNSSECConfig.ValidationStatus) == route53resolver.ResolverDNSSECValidationStatusDisabled {
			continue
		}

		return fmt.Errorf("Route 53 Resolver Dnssec config still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRoute53ResolverDnssecConfigExists(n string, c *route53resolver.ResolverDnssecConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver Dnssec config ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).route53resolverconn
		resp, err := conn.GetResolverDnssecConfig(&route53resolver.GetResolverDnssecConfigInput{
			ResourceId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*c = *resp.ResolverDNSSECConfig

		return nil
	}
}

func testAccRoute53ResolverDnssecConfigBase(rName string) string {
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

func testAccRoute53ResolverDnssecConfigConfigBasic(rName string) string {
	return testAccRoute53ResolverDnssecConfigBase(rName) + `
resource "aws_route53_resolver_dnssec_config" "test" {
  resource_id = aws_vpc.test.id
}
`
}
