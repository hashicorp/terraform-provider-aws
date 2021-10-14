package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_ses_email_identity", &resource.Sweeper{
		Name: "aws_ses_email_identity",
		F:    func(region string) error { return testSweepSesIdentities(region, ses.IdentityTypeEmailAddress) },
	})
}

func TestAccAWSSESEmailIdentity_basic(t *testing.T) {
	email := acctest.DefaultEmailAddress
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsSESEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESEmailIdentityConfig(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailIdentityExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ses", regexp.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(email)))),
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

func TestAccAWSSESEmailIdentity_trailingPeriod(t *testing.T) {
	email := fmt.Sprintf("%s.", acctest.DefaultEmailAddress)
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsSESEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESEmailIdentityConfig(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailIdentityExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ses", regexp.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(strings.TrimSuffix(email, "."))))),
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

func testAccCheckAwsSESEmailIdentityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*AWSClient).sesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_email_identity" {
			continue
		}

		email := rs.Primary.ID
		params := &ses.GetIdentityVerificationAttributesInput{
			Identities: []*string{
				aws.String(email),
			},
		}

		response, err := conn.GetIdentityVerificationAttributes(params)
		if err != nil {
			return err
		}

		if response.VerificationAttributes[email] != nil {
			return fmt.Errorf("SES Email Identity %s still exists. Failing!", email)
		}
	}

	return nil
}

func testAccCheckAwsSESEmailIdentityExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Email Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Email Identity name not set")
		}

		email := rs.Primary.ID
		conn := acctest.Provider.Meta().(*AWSClient).sesconn

		params := &ses.GetIdentityVerificationAttributesInput{
			Identities: []*string{
				aws.String(email),
			},
		}

		response, err := conn.GetIdentityVerificationAttributes(params)
		if err != nil {
			return err
		}

		if response.VerificationAttributes[email] == nil {
			return fmt.Errorf("SES Email Identity %s not found in AWS", email)
		}

		return nil
	}
}

func testAccAwsSESEmailIdentityConfig(email string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
  email = %q
}
`, email)
}
