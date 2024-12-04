// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESDomainIdentity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	resourceName := "aws_ses_domain_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainIdentityExists(ctx, resourceName),
					testAccCheckDomainIdentityARN(ctx, resourceName, domain),
				),
			},
		},
	})
}

func TestAccSESDomainIdentity_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	resourceName := "aws_ses_domain_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainIdentityExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfses.ResourceDomainIdentity(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccSESDomainIdentity_trailingPeriod updated in 3.0 to account for domain plan-time validation
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13510
func TestAccSESDomainIdentity_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDomainIdentityConfig_basic(domain),
				ExpectError: regexache.MustCompile(`invalid value for domain \(cannot end with a period\)`),
			},
		},
	})
}

func testAccCheckDomainIdentityDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_domain_identity" {
				continue
			}

			_, err := tfses.FindIdentityVerificationAttributesByIdentity(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SES Domain Identity %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainIdentityExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		_, err := tfses.FindIdentityVerificationAttributesByIdentity(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckDomainIdentityARN(ctx context.Context, n string, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		expected := acctest.Provider.Meta().(*conns.AWSClient).RegionalARN(ctx, "ses", fmt.Sprintf("identity/%s", strings.TrimSuffix(domain, ".")))

		if rs.Primary.Attributes[names.AttrARN] != expected {
			return fmt.Errorf("Incorrect ARN: expected %q, got %q", expected, rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

	input := &ses.ListIdentitiesInput{}

	_, err := conn.ListIdentities(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDomainIdentityConfig_basic(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}
`, domain)
}
