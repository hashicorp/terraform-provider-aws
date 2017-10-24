package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsSESDomainMailFrom_basic(t *testing.T) {
	domain := fmt.Sprintf(
		"%s.terraformtesting.com",
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccAwsSESDomainMailFromConfig, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainMailFromExists("aws_ses_domain_mail_from.test"),
				),
			},
		},
	})
}

func testAccCheckAwsSESDomainMailFromExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Domain Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Domain Identity name not set")
		}

		domain := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).sesConn

		params := &ses.GetIdentityMailFromDomainAttributesInput{
			Identities: []*string{
				aws.String(domain),
			},
		}

		response, err := conn.GetIdentityMailFromDomainAttributes(params)
		if err != nil {
			return err
		}

		if response.MailFromDomainAttributes[domain] == nil {
			return fmt.Errorf("SES Domain MAIL FROM %s not found in AWS", domain)
		}

		return nil
	}
}

const testAccAwsSESDomainMailFromConfig = `
resource "aws_ses_domain_identity" "test" {
	domain = "%s"
}
resource "aws_ses_domain_mail_from" "test" {
	domain = "${aws_ses_domain_identity.test.domain}"
	mail_from_domain = "bounce.${aws_ses_domain_identity.test.domain}"
}
`
