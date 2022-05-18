package ses_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccDomainIdentityDomainFromEnv(t *testing.T) string {
	rootDomain := os.Getenv("SES_DOMAIN_IDENTITY_ROOT_DOMAIN")
	if rootDomain == "" {
		t.Skip(
			"Environment variable SES_DOMAIN_IDENTITY_ROOT_DOMAIN is not set. " +
				"For DNS verification requests, this domain must be publicly " +
				"accessible and configurable via Route53 during the testing. ")
	}
	return rootDomain
}

func TestAccSESDomainIdentityVerification_basic(t *testing.T) {
	rootDomain := testAccDomainIdentityDomainFromEnv(t)
	domain := fmt.Sprintf("tf-acc-%d.%s", sdkacctest.RandInt(), rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityVerification_basic(rootDomain, domain),
				Check:  testAccCheckDomainIdentityVerificationPassed("aws_ses_domain_identity_verification.test"),
			},
		},
	})
}

func TestAccSESDomainIdentityVerification_timeout(t *testing.T) {
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDomainIdentityVerification_timeout(domain),
				ExpectError: regexp.MustCompile("Expected domain verification Success, but was in state Pending"),
			},
		},
	})
}

func TestAccSESDomainIdentityVerification_nonexistent(t *testing.T) {
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDomainIdentityVerification_nonexistent(domain),
				ExpectError: regexp.MustCompile(fmt.Sprintf("SES Domain Identity %s not found in AWS", domain)),
			},
		},
	})
}

func testAccCheckDomainIdentityVerificationPassed(n string) resource.TestCheckFunc {
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

		if aws.StringValue(response.VerificationAttributes[domain].VerificationStatus) != ses.VerificationStatusSuccess {
			return fmt.Errorf("SES Domain Identity %s not successfully verified.", domain)
		}

		expected := arn.ARN{
			AccountID: acctest.Provider.Meta().(*conns.AWSClient).AccountID,
			Partition: acctest.Provider.Meta().(*conns.AWSClient).Partition,
			Region:    acctest.Provider.Meta().(*conns.AWSClient).Region,
			Resource:  fmt.Sprintf("identity/%s", domain),
			Service:   "ses",
		}

		if rs.Primary.Attributes["arn"] != expected.String() {
			return fmt.Errorf("Incorrect ARN: expected %q, got %q", expected, rs.Primary.Attributes["arn"])
		}

		return nil
	}
}

func testAccDomainIdentityVerification_basic(rootDomain string, domain string) string {
	return fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = "%s."
  private_zone = false
}

resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_route53_record" "domain_identity_verification" {
  zone_id = data.aws_route53_zone.test.id
  name    = "_amazonses.${aws_ses_domain_identity.test.id}"
  type    = "TXT"
  ttl     = "600"
  records = [aws_ses_domain_identity.test.verification_token]
}

resource "aws_ses_domain_identity_verification" "test" {
  domain = aws_ses_domain_identity.test.id

  depends_on = [aws_route53_record.domain_identity_verification]
}
`, rootDomain, domain)
}

func testAccDomainIdentityVerification_timeout(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_ses_domain_identity_verification" "test" {
  domain = aws_ses_domain_identity.test.id

  timeouts {
    create = "5s"
  }
}
`, domain)
}

func testAccDomainIdentityVerification_nonexistent(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity_verification" "test" {
  domain = "%s"
}
`, domain)
}
