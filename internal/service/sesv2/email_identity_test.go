package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/names"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSESEmailIdentity_basic(t *testing.T) {
	email := fmt.Sprintf("res-basic%s", acctest.DefaultEmailAddress)
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointId),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy,
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
	email := fmt.Sprintf("res-trailing%s.", acctest.DefaultEmailAddress)
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointId),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy,
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

func TestAccSESEmailIdentity_complete(t *testing.T) {
	email := fmt.Sprintf("res-complete%s", acctest.DefaultEmailAddress)
	resourceName := "aws_ses_email_identity.test"
	configSetName := "default-config-set-name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointId),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_complete(email, configSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "default_configuration_set", configSetName),
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

func testAccCheckEmailIdentityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_email_identity" {
			continue
		}

		email := rs.Primary.ID
		params := &sesv2.GetEmailIdentityInput{EmailIdentity: aws.String(email)}
		ctx := context.Background()
		response, err := conn.GetEmailIdentity(ctx, params)

		if err != nil {
			var nfe *types.NotFoundException
			if errors.As(err, &nfe) {
				// Destroy succeeded - Email not found
				return nil
			}
			return err
		}

		if response != nil && err == nil {
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

		params := &sesv2.GetEmailIdentityInput{EmailIdentity: aws.String(email)}
		ctx := context.Background()
		_, err := conn.GetEmailIdentity(ctx, params)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, (&types.NotFoundException{}).ErrorMessage()) {
				return fmt.Errorf("SES Email Identity %s not found in AWS", email)
			}
			return err
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

func testAccEmailIdentityConfig_complete(email string, configSet string) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = %q
}

resource "aws_ses_email_identity" "test" {
  email = %q
  default_configuration_set = aws_ses_configuration_set.test.name
}
`, configSet, email)
}
