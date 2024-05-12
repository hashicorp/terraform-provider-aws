// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESDomainMailFrom_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dn := acctest.RandomDomain()
	domain := dn.String()
	mailFromDomain1 := dn.Subdomain("bounce1").String()
	mailFromDomain2 := dn.Subdomain("bounce2").String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainMailFromDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainMailFromConfig_basic(domain, mailFromDomain1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, domain),
					resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain1),
				),
			},
			{
				Config: testAccDomainMailFromConfig_basic(domain, mailFromDomain2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, domain),
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
	ctx := acctest.Context(t)
	dn := acctest.RandomDomain()
	domain := dn.String()
	mailFromDomain := dn.Subdomain("bounce").String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainMailFromDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainMailFromConfig_basic(domain, mailFromDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(ctx, resourceName),
					testAccCheckDomainMailFromDisappears(ctx, domain),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESDomainMailFrom_Disappears_identity(t *testing.T) {
	ctx := acctest.Context(t)
	dn := acctest.RandomDomain()
	domain := dn.String()
	mailFromDomain := dn.Subdomain("bounce").String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainMailFromDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainMailFromConfig_basic(domain, mailFromDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(ctx, resourceName),
					testAccCheckDomainIdentityDisappears(ctx, domain),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESDomainMailFrom_behaviorOnMxFailure(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomain().String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainMailFromDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainMailFromConfig_behaviorOnMxFailure(domain, ses.BehaviorOnMXFailureUseDefaultValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
				),
			},
			{
				Config: testAccDomainMailFromConfig_behaviorOnMxFailure(domain, ses.BehaviorOnMXFailureRejectMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(ctx, resourceName),
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

func testAccCheckDomainMailFromExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Domain Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Domain Identity name not set")
		}

		domain := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		params := &ses.GetIdentityMailFromDomainAttributesInput{
			Identities: []*string{
				aws.String(domain),
			},
		}

		response, err := conn.GetIdentityMailFromDomainAttributesWithContext(ctx, params)
		if err != nil {
			return err
		}

		if response.MailFromDomainAttributes[domain] == nil {
			return fmt.Errorf("SES Domain MAIL FROM %s not found in AWS", domain)
		}

		return nil
	}
}

func testAccCheckDomainMailFromDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_domain_mail_from" {
				continue
			}

			input := &ses.GetIdentityMailFromDomainAttributesInput{
				Identities: []*string{aws.String(rs.Primary.ID)},
			}

			out, err := conn.GetIdentityMailFromDomainAttributesWithContext(ctx, input)
			if err != nil {
				return fmt.Errorf("fetching MAIL FROM domain attributes: %s", err)
			}
			if v, ok := out.MailFromDomainAttributes[rs.Primary.ID]; ok && v.MailFromDomain != nil && *v.MailFromDomain != "" {
				return fmt.Errorf("MAIL FROM domain was not removed, found: %s", *v.MailFromDomain)
			}
		}

		return nil
	}
}

func testAccCheckDomainMailFromDisappears(ctx context.Context, identity string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		input := &ses.SetIdentityMailFromDomainInput{
			Identity:       aws.String(identity),
			MailFromDomain: nil,
		}

		_, err := conn.SetIdentityMailFromDomainWithContext(ctx, input)

		return err
	}
}

func testAccDomainMailFromConfig_basic(domain, mailFromDomain string) string {
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
