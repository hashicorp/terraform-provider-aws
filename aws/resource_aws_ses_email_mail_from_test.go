package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSESEmailMailFrom_basic(t *testing.T) {
	dn := testAccRandomDomain()
	email := testAccDefaultEmailAddress
	mailFromDomain1 := dn.Subdomain("bounce1").String()
	mailFromDomain2 := dn.Subdomain("bounce2").String()
	resourceName := "aws_ses_email_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESEmailMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESEmailMailFromConfig(email, mailFromDomain1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain1),
				),
			},
			{
				Config: testAccAwsSESEmailMailFromConfig(email, mailFromDomain2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain2),
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

func TestAccAWSSESEmailMailFrom_disappears(t *testing.T) {
	dn := testAccRandomDomain()
	email := testAccDefaultEmailAddress
	mailFromDomain := dn.Subdomain("bounce").String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESEmailMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESEmailMailFromConfig(email, mailFromDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailMailFromExists(resourceName),
					testAccCheckAwsSESEmailMailFromDisappears(email),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSESEmailMailFrom_disappears_Identity(t *testing.T) {
	dn := testAccRandomDomain()
	email := testAccDefaultEmailAddress
	mailFromDomain := dn.Subdomain("bounce").String()
	resourceName := "aws_ses_email_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESEmailMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESEmailMailFromConfig(email, mailFromDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailMailFromExists(resourceName),
					testAccCheckAwsSESEmailMailFromDisappears(email),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSESEmailMailFrom_behaviorOnMxFailure(t *testing.T) {
	email := testAccDefaultEmailAddress
	resourceName := "aws_ses_email_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESEmailMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESEmailMailFromConfig_behaviorOnMxFailure(email, ses.BehaviorOnMXFailureUseDefaultValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
				),
			},
			{
				Config: testAccAwsSESEmailMailFromConfig_behaviorOnMxFailure(email, ses.BehaviorOnMXFailureRejectMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureRejectMessage),
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

func testAccCheckAwsSESEmailMailFromExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Email Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Email Identity name not set")
		}

		email := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).sesconn

		params := &ses.GetIdentityMailFromDomainAttributesInput{
			Identities: []*string{
				aws.String(email),
			},
		}

		response, err := conn.GetIdentityMailFromDomainAttributes(params)
		if err != nil {
			return err
		}

		if response.MailFromDomainAttributes[email] == nil {
			return fmt.Errorf("SES Domain MAIL FROM %s not found in AWS", email)
		}

		return nil
	}
}

func testAccCheckSESEmailMailFromDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_email_mail_from" {
			continue
		}

		input := &ses.GetIdentityMailFromDomainAttributesInput{
			Identities: []*string{aws.String(rs.Primary.ID)},
		}

		out, err := conn.GetIdentityMailFromDomainAttributes(input)
		if err != nil {
			return fmt.Errorf("error fetching MAIL FROM domain attributes: %s", err)
		}
		if v, ok := out.MailFromDomainAttributes[rs.Primary.ID]; ok && v.MailFromDomain != nil && *v.MailFromDomain != "" {
			return fmt.Errorf("MAIL FROM domain was not removed, found: %s", *v.MailFromDomain)
		}
	}

	return nil
}

func testAccCheckAwsSESEmailMailFromDisappears(identity string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sesconn

		input := &ses.SetIdentityMailFromDomainInput{
			Identity:       aws.String(identity),
			MailFromDomain: nil,
		}

		_, err := conn.SetIdentityMailFromDomain(input)

		return err
	}
}

func testAccAwsSESEmailMailFromConfig(email, mailFromDomain string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
  email = "%s"
}

resource "aws_ses_email_mail_from" "test" {
  email            = aws_ses_email_identity.test.email
  mail_from_domain = "%s"
}
`, email, mailFromDomain)
}

func testAccAwsSESEmailMailFromConfig_behaviorOnMxFailure(email, behaviorOnMxFailure string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
  email = "%s"
}

resource "aws_ses_email_mail_from" "test" {
  behavior_on_mx_failure = "%s"
  email                  = aws_ses_email_identity.test.email
  mail_from_domain       = "bounce.${aws_ses_domain_identity.test.domain}"
}
`, email, behaviorOnMxFailure)
}
