package ses_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSESDomainIdentity_basic(t *testing.T) {
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainIdentityExists("aws_ses_domain_identity.test"),
					testAccCheckDomainIdentityARN("aws_ses_domain_identity.test", domain),
				),
			},
		},
	})
}

func TestAccSESDomainIdentity_disappears(t *testing.T) {
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainIdentityExists("aws_ses_domain_identity.test"),
					testAccCheckDomainIdentityDisappears(domain),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccSESDomainIdentity_trailingPeriod updated in 3.0 to account for domain plan-time validation
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13510
func TestAccSESDomainIdentity_trailingPeriod(t *testing.T) {
	domain := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDomainIdentityConfig_basic(domain),
				ExpectError: regexp.MustCompile(`invalid value for domain \(cannot end with a period\)`),
			},
		},
	})
}

func testAccCheckDomainIdentityDestroy(s *terraform.State) error {
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

func testAccCheckDomainIdentityExists(n string) resource.TestCheckFunc {
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

func testAccCheckDomainIdentityDisappears(identity string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		input := &ses.DeleteIdentityInput{
			Identity: aws.String(identity),
		}

		_, err := conn.DeleteIdentity(input)

		return err
	}
}

func testAccCheckDomainIdentityARN(n string, domain string) resource.TestCheckFunc {
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

func testAccPreCheck(t *testing.T) {
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

func testAccDomainIdentityConfig_basic(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}
`, domain)
}
