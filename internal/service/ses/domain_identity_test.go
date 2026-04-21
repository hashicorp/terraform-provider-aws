// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESDomainIdentity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	resourceName := "aws_ses_domain_identity.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityConfig_basic(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainIdentityExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ses", "identity/{domain}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, domain),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrDomain),
					resource.TestCheckResourceAttrSet(resourceName, "verification_token"),
				),
			},
		},
	})
}

func TestAccSESDomainIdentity_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	resourceName := "aws_ses_domain_identity.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityConfig_basic(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainIdentityExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfses.ResourceDomainIdentity(), resourceName),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccDomainIdentityConfig_basic(domain),
				ExpectError: regexache.MustCompile(`invalid value for domain \(cannot end with a period\)`),
			},
		},
	})
}

func testAccCheckDomainIdentityDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_domain_identity" {
				continue
			}

			_, err := tfses.FindIdentityVerificationAttributesByIdentity(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckDomainIdentityExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		_, err := tfses.FindIdentityVerificationAttributesByIdentity(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

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
