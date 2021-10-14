package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	resource.AddTestSweepers("aws_ses_domain_identity", &resource.Sweeper{
		Name: "aws_ses_domain_identity",
		F:    func(region string) error { return testSweepSesIdentities(region, ses.IdentityTypeDomain) },
	})
}

func testSweepSesIdentities(region, identityType string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SESConn
	input := &ses.ListIdentitiesInput{
		IdentityType: aws.String(identityType),
	}
	var sweeperErrs *multierror.Error

	err = conn.ListIdentitiesPages(input, func(page *ses.ListIdentitiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, identity := range page.Identities {
			identity := aws.StringValue(identity)

			log.Printf("[INFO] Deleting SES Identity: %s", identity)
			_, err = conn.DeleteIdentity(&ses.DeleteIdentityInput{
				Identity: aws.String(identity),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting SES Identity (%s): %w", identity, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SES Identities sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SES Identities: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSESDomainIdentity_basic(t *testing.T) {
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsSESDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESDomainIdentityConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainIdentityExists("aws_ses_domain_identity.test"),
					testAccCheckAwsSESDomainIdentityArn("aws_ses_domain_identity.test", domain),
				),
			},
		},
	})
}

func TestAccAWSSESDomainIdentity_disappears(t *testing.T) {
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsSESDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESDomainIdentityConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainIdentityExists("aws_ses_domain_identity.test"),
					testAccCheckAwsSESDomainIdentityDisappears(domain),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccAWSSESDomainIdentity_trailingPeriod updated in 3.0 to account for domain plan-time validation
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13510
func TestAccAWSSESDomainIdentity_trailingPeriod(t *testing.T) {
	domain := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsSESDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsSESDomainIdentityConfig(domain),
				ExpectError: regexp.MustCompile(`invalid value for domain \(cannot end with a period\)`),
			},
		},
	})
}

func testAccCheckAwsSESDomainIdentityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_domain_identity" {
			continue
		}

		domain := rs.Primary.ID
		params := &ses.GetIdentityVerificationAttributesInput{
			Identities: []*string{
				aws.String(domain),
			},
		}

		response, err := conn.GetIdentityVerificationAttributes(params)
		if err != nil {
			return err
		}

		if response.VerificationAttributes[domain] != nil {
			return fmt.Errorf("SES Domain Identity %s still exists. Failing!", domain)
		}
	}

	return nil
}

func testAccCheckAwsSESDomainIdentityExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Domain Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Domain Identity name not set")
		}

		domain := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		params := &ses.GetIdentityVerificationAttributesInput{
			Identities: []*string{
				aws.String(domain),
			},
		}

		response, err := conn.GetIdentityVerificationAttributes(params)
		if err != nil {
			return err
		}

		if response.VerificationAttributes[domain] == nil {
			return fmt.Errorf("SES Domain Identity %s not found in AWS", domain)
		}

		return nil
	}
}

func testAccCheckAwsSESDomainIdentityDisappears(identity string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		input := &ses.DeleteIdentityInput{
			Identity: aws.String(identity),
		}

		_, err := conn.DeleteIdentity(input)

		return err
	}
}

func testAccCheckAwsSESDomainIdentityArn(n string, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		awsClient := acctest.Provider.Meta().(*conns.AWSClient)

		expected := arn.ARN{
			AccountID: awsClient.AccountID,
			Partition: awsClient.Partition,
			Region:    awsClient.Region,
			Resource:  fmt.Sprintf("identity/%s", strings.TrimSuffix(domain, ".")),
			Service:   "ses",
		}

		if rs.Primary.Attributes["arn"] != expected.String() {
			return fmt.Errorf("Incorrect ARN: expected %q, got %q", expected, rs.Primary.Attributes["arn"])
		}

		return nil
	}
}

func testAccPreCheckAWSSES(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	input := &ses.ListIdentitiesInput{}

	_, err := conn.ListIdentities(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsSESDomainIdentityConfig(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}
`, domain)
}
