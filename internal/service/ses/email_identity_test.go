package ses_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSESEmailIdentity_basic(t *testing.T) {
	domain := acctest.RandomDomainName()
	email := acctest.RandomEmailAddress(domain)
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sesv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig(email),
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
	domain := acctest.RandomDomainName()
	email := acctest.RandomEmailAddress(domain)
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sesv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig(email),
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_email_identity" {
			continue
		}

		email := rs.Primary.ID
		params := &sesv2.GetEmailIdentityInput{
			EmailIdentity: aws.String(email),
		}

		_, err := conn.GetEmailIdentity(params)
		if err == nil {
			return fmt.Errorf("SES Email Identity %s still exists", email)
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

		params := &sesv2.GetEmailIdentityInput{
			EmailIdentity: aws.String(email),
		}

		_, err := conn.GetEmailIdentity(params)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccEmailIdentityConfig(email string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
  email = %q
}
`, email)
}
