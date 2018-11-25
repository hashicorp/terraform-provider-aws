package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSESEmailIdentity_basic(t *testing.T) {
	email := fmt.Sprintf(
		"%s@terraformtesting.com",
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESEmailIdentityConfig(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailIdentityExists("aws_ses_email_identity.test"),
					testAccCheckAwsSESEmailIdentityArn("aws_ses_email_identity.test", email),
				),
			},
		},
	})
}

func TestAccAWSSESEmailIdentity_trailingPeriod(t *testing.T) {
	email := fmt.Sprintf(
		"%s@terraformtesting.com.",
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESEmailIdentityConfig(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailIdentityExists("aws_ses_email_identity.test"),
					testAccCheckAwsSESEmailIdentityArn("aws_ses_email_identity.test", email),
				),
			},
		},
	})
}

func testAccCheckAwsSESEmailIdentityDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesConn

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
		conn := testAccProvider.Meta().(*AWSClient).sesConn

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

func testAccCheckAwsSESEmailIdentityArn(n string, email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		awsClient := testAccProvider.Meta().(*AWSClient)

		expected := arn.ARN{
			AccountID: awsClient.accountid,
			Partition: awsClient.partition,
			Region:    awsClient.region,
			Resource:  fmt.Sprintf("identity/%s", strings.TrimSuffix(email, ".")),
			Service:   "ses",
		}

		if rs.Primary.Attributes["arn"] != expected.String() {
			return fmt.Errorf("Incorrect ARN: expected %q, got %q", expected, rs.Primary.Attributes["arn"])
		}

		return nil
	}
}

func testAccAwsSESEmailIdentityConfig(email string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
	email = "%s"
}
`, email)
}
