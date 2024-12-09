// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", string(awstypes.BehaviorOnMXFailureUseDefaultValue)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, domain),
					resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain1),
				),
			},
			{
				Config: testAccDomainMailFromConfig_basic(domain, mailFromDomain2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", string(awstypes.BehaviorOnMXFailureUseDefaultValue)),
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
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfses.ResourceDomainMailFrom(), resourceName),
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
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfses.ResourceDomainIdentity(), "aws_ses_domain_identity.test"),
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
				Config: testAccDomainMailFromConfig_behaviorOnMxFailure(domain, string(awstypes.BehaviorOnMXFailureUseDefaultValue)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", string(awstypes.BehaviorOnMXFailureUseDefaultValue)),
				),
			},
			{
				Config: testAccDomainMailFromConfig_behaviorOnMxFailure(domain, string(awstypes.BehaviorOnMXFailureRejectMessage)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainMailFromExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", string(awstypes.BehaviorOnMXFailureRejectMessage)),
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
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		_, err := tfses.FindIdentityMailFromDomainAttributesByIdentity(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckDomainMailFromDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_domain_mail_from" {
				continue
			}

			_, err := tfses.FindIdentityMailFromDomainAttributesByIdentity(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SES MAIL FROM Domain %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainMailFromConfig_basic(domain, mailFromDomain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_domain_mail_from" "test" {
  domain           = aws_ses_domain_identity.test.domain
  mail_from_domain = %[2]q
}
`, domain, mailFromDomain)
}

func testAccDomainMailFromConfig_behaviorOnMxFailure(domain, behaviorOnMxFailure string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_domain_mail_from" "test" {
  behavior_on_mx_failure = %[2]q
  domain                 = aws_ses_domain_identity.test.domain
  mail_from_domain       = "bounce.${aws_ses_domain_identity.test.domain}"
}
`, domain, behaviorOnMxFailure)
}
