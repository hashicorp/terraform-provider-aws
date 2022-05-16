package sesv2_test

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

func TestAccSESv2EmailIdentity_basic(t *testing.T) {
	identity := acctest.DefaultEmailAddress
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sesv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig(identity),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ses", regexp.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(identity)))),
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

func TestAccSESv2EmailIdentity_domain(t *testing.T) {
	identity := acctest.RandomDomainName()
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sesv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig(identity),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ses", regexp.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(identity)))),
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

func TestAccSESv2EmailIdentity_trailingPeriod(t *testing.T) {
	identity := fmt.Sprintf("%s.", acctest.DefaultEmailAddress)
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sesv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig(identity),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ses", regexp.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(strings.TrimSuffix(identity, "."))))),
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

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

	input := &sesv2.ListEmailIdentitiesInput{}

	_, err := conn.ListEmailIdentities(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckEmailIdentityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sesv2_email_identity" {
			continue
		}

		identity := rs.Primary.ID
		params := &sesv2.GetEmailIdentityInput{
			EmailIdentity: aws.String(identity),
		}

		_, err := conn.GetEmailIdentity(params)
		if err != nil {
			if _, ok := err.(*sesv2.NotFoundException); ok {
				return nil
			}
			return err
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

		identity := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

		params := &sesv2.GetEmailIdentityInput{
			EmailIdentity: aws.String(identity),
		}

		_, err := conn.GetEmailIdentity(params)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccEmailIdentityConfig(identity string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  identity = %q
}
`, identity)
}
