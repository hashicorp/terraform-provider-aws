package ses_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSESDomainMailFrom_basic(t *testing.T) {
	dn := acctest.RandomDomain()
	domain := dn.String()
	mailFromDomain1 := dn.Subdomain("bounce1").String()
	mailFromDomain2 := dn.Subdomain("bounce2").String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainMailFromConfig(domain, mailFromDomain1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain1),
				),
			},
			{
				Config: testAccDomainMailFromConfig(domain, mailFromDomain2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
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

func TestAccSESDomainMailFrom_disappears(t *testing.T) {
	dn := acctest.RandomDomain()
	domain := dn.String()
	mailFromDomain := dn.Subdomain("bounce").String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainMailFromConfig(domain, mailFromDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(resourceName),
					testAccCheckDomainMailFromDisappears(domain),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESDomainMailFrom_Disappears_identity(t *testing.T) {
	dn := acctest.RandomDomain()
	domain := dn.String()
	mailFromDomain := dn.Subdomain("bounce").String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainMailFromConfig(domain, mailFromDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(resourceName),
					testAccCheckDomainIdentityDisappears(domain),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESDomainMailFrom_behaviorOnMxFailure(t *testing.T) {
	domain := acctest.RandomDomain().String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainMailFromConfig_behaviorOnMxFailure(domain, ses.BehaviorOnMXFailureUseDefaultValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
				),
			},
			{
				Config: testAccDomainMailFromConfig_behaviorOnMxFailure(domain, ses.BehaviorOnMXFailureRejectMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(resourceName),
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

func testAccCheckDomainMailFromExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Domain Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Domain Identity name not set")
		}

		domain := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

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

func testAccCheckDomainMailFromDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_domain_mail_from" {
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

func testAccCheckDomainMailFromDisappears(identity string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		input := &ses.SetIdentityMailFromDomainInput{
			Identity:       aws.String(identity),
			MailFromDomain: nil,
		}

		_, err := conn.SetIdentityMailFromDomain(input)

		return err
	}
}

func testAccDomainMailFromConfig(domain, mailFromDomain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_ses_domain_mail_from" "test" {
  domain           = aws_ses_domain_identity.test.domain
  mail_from_domain = "%s"
}
`, domain, mailFromDomain)
}

func testAccDomainMailFromConfig_behaviorOnMxFailure(domain, behaviorOnMxFailure string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_ses_domain_mail_from" "test" {
  behavior_on_mx_failure = "%s"
  domain                 = aws_ses_domain_identity.test.domain
  mail_from_domain       = "bounce.${aws_ses_domain_identity.test.domain}"
}
`, domain, behaviorOnMxFailure)
}
