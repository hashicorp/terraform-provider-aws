package ses_test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSESDomainDKIM_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ses_domain_dkim.test"
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ses.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDKIMDestroy(ctx),
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: fmt.Sprintf(testAccDomainDKIMConfig, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainDKIMExists(ctx, resourceName),
					testAccCheckDomainDKIMTokens(resourceName),
				),
			},
		},
	})
}

func testAccCheckDomainDKIMDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_domain_dkim" {
				continue
			}

			domain := rs.Primary.ID
			params := &ses.GetIdentityDkimAttributesInput{
				Identities: []*string{
					aws.String(domain),
				},
			}

			resp, err := conn.GetIdentityDkimAttributesWithContext(ctx, params)

			if err != nil {
				return err
			}

			if resp.DkimAttributes[domain] != nil {
				return fmt.Errorf("SES Domain Dkim %s still exists.", domain)
			}
		}

		return nil
	}
}

func testAccCheckDomainDKIMExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Domain Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Domain Identity name not set")
		}

		domain := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn()

		params := &ses.GetIdentityDkimAttributesInput{
			Identities: []*string{
				aws.String(domain),
			},
		}

		response, err := conn.GetIdentityDkimAttributesWithContext(ctx, params)
		if err != nil {
			return err
		}

		if response.DkimAttributes[domain] == nil {
			return fmt.Errorf("SES Domain DKIM %s not found in AWS", domain)
		}

		return nil
	}
}

func testAccCheckDomainDKIMTokens(n string) resource.TestCheckFunc {
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

const testAccDomainDKIMConfig = `
resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_domain_dkim" "test" {
  domain = aws_ses_domain_identity.test.domain
}
`
