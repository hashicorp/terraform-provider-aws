// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDomainIdentityDomainFromEnv(t *testing.T) string {
	rootDomain := os.Getenv("SES_DOMAIN_IDENTITY_ROOT_DOMAIN")
	if rootDomain == "" {
		t.Skip(
			"Environment variable SES_DOMAIN_IDENTITY_ROOT_DOMAIN is not set. " +
				"For DNS verification requests, this domain must be publicly " +
				"accessible and configurable via Route53 during the testing. ")
	}
	return rootDomain
}

func TestAccSESDomainIdentityVerification_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := testAccDomainIdentityDomainFromEnv(t)
	domain := fmt.Sprintf("tf-acc-%d.%s", sdkacctest.RandInt(), rootDomain)
	resourceName := "aws_ses_domain_identity_verification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityVerificationConfig_basic(rootDomain, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ses", "identity/{domain}"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, "aws_ses_domain_identity.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, domain),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrDomain),
				),
			},
		},
	})
}

func TestAccSESDomainIdentityVerification_timeout(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccDomainIdentityVerificationConfig_timeout(domain),
				ExpectError: regexache.MustCompile(`output = Pending, want Success`),
			},
		},
	})
}

func TestAccSESDomainIdentityVerification_nonexistent(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccDomainIdentityVerificationConfig_nonexistent(domain),
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
		},
	})
}

func testAccDomainIdentityVerificationConfig_basic(rootDomain, domain string) string {
	return fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = "%[1]s."
  private_zone = false
}

resource "aws_ses_domain_identity" "test" {
  domain = %[2]q
}

resource "aws_route53_record" "domain_identity_verification" {
  zone_id = data.aws_route53_zone.test.id
  name    = "_amazonses.${aws_ses_domain_identity.test.domain}"
  type    = "TXT"
  ttl     = "600"
  records = [aws_ses_domain_identity.test.verification_token]
}

resource "aws_ses_domain_identity_verification" "test" {
  domain = aws_ses_domain_identity.test.domain

  depends_on = [aws_route53_record.domain_identity_verification]
}
`, rootDomain, domain)
}

func testAccDomainIdentityVerificationConfig_timeout(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_domain_identity_verification" "test" {
  domain = aws_ses_domain_identity.test.domain

  timeouts {
    create = "5s"
  }
}
`, domain)
}

func testAccDomainIdentityVerificationConfig_nonexistent(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity_verification" "test" {
  domain = %[1]q
}
`, domain)
}
