package ses_test

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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSESEmailIdentity_basic(t *testing.T) {
	email := acctest.DefaultEmailAddress
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(resourceName),
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

func TestAccSESEmailIdentity_trailingPeriod(t *testing.T) {
	email := fmt.Sprintf("%s.", acctest.DefaultEmailAddress)
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(resourceName),
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

func testAccCheckEmailIdentityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

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

func testAccCheckEmailIdentityExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Email Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Email Identity name not set")
		}

		email := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

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

func testAccEmailIdentityConfig_basic(email string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
  email = %q
}
`, email)
}
