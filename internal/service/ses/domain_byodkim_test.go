package ses_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSESDomainBYODKIM_basic(t *testing.T) {
	resourceName := "aws_ses_domain_dkim.test"
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainBYODKIMDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccDomainBYODKIMConfig, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainBYODKIMExists(resourceName),
					testAccCheckDomainBYODKIMTokens(resourceName),
				),
			},
		},
	})
}

func testAccCheckDomainBYODKIMDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_domain_dkim" {
			continue
		}

		domain := rs.Primary.ID
		params := &sesv2.GetEmailIdentityInput{
			EmailIdentity: aws.String(domain),
		}

		res, err := conn.GetEmailIdentity(params)
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}

		if res.DkimAttributes != nil {
			return fmt.Errorf("SES Domain Dkim %s still exists.", domain)
		}
	}

	return nil
}

func testAccCheckDomainBYODKIMExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Domain Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Domain Identity name not set")
		}

		domain := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

		params := &sesv2.GetEmailIdentityInput{
			EmailIdentity: aws.String(domain),
		}

		res, err := conn.GetEmailIdentity(params)
		if err != nil {
			return err
		}

		if res.DkimAttributes == nil {
			return fmt.Errorf("SES Domain DKIM %s not found in AWS", domain)
		}

		return nil
	}
}

func testAccCheckDomainBYODKIMTokens(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		expectedNum := 3
		expectedFormat := regexp.MustCompile("[a-z0-9]{32}")

		tokenNum, _ := strconv.Atoi(rs.Primary.Attributes["dkim_tokens.#"])
		if expectedNum != tokenNum {
			return fmt.Errorf("Incorrect number of DKIM tokens, expected: %d, got: %d", expectedNum, tokenNum)
		}
		for i := 0; i < expectedNum; i++ {
			key := fmt.Sprintf("dkim_tokens.%d", i)
			token := rs.Primary.Attributes[key]
			if !expectedFormat.MatchString(token) {
				return fmt.Errorf("Incorrect format of DKIM token: %v", token)
			}
		}

		return nil
	}
}

const testAccDomainBYODKIMConfig = `
resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_domain_byodkim" "test" {
  domain = aws_ses_domain_identity.test.domain
  selector = default
  private_key = default
}
`
