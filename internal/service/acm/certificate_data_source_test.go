// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acm_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const certificateRE = `^arn:[^:]+:acm:[^:]+:[^:]+:certificate/.+$`

func TestAccACMCertificateDataSource_singleIssued(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	var arnRe *regexp.Regexp
	var domain string

	if os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_MOST_RECENT_ARN") != "" {
		arnRe = regexache.MustCompile(fmt.Sprintf("^%s$", os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_MOST_RECENT_ARN")))
	} else {
		arnRe = regexache.MustCompile(certificateRE)
	}

	if os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_DOMAIN") != "" {
		domain = os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_DOMAIN")
	} else {
		domain = fmt.Sprintf("tf-acc-single-issued.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))
	}

	resourceName := "data.aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.CertificateStatusIssued)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_status(domain, string(awstypes.CertificateStatusIssued)),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.CertificateStatusIssued)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_types(domain, string(awstypes.CertificateTypeAmazonIssued)),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecent(domain, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndStatus(domain, string(awstypes.CertificateStatusIssued), true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndTypes(domain, string(awstypes.CertificateTypeAmazonIssued), true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_multipleIssued(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	var arnRe *regexp.Regexp
	var domain string

	if os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_MOST_RECENT_ARN") != "" {
		arnRe = regexache.MustCompile(fmt.Sprintf("^%s$", os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_MOST_RECENT_ARN")))
	} else {
		arnRe = regexache.MustCompile(certificateRE)
	}

	if os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_DOMAIN") != "" {
		domain = os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_DOMAIN")
	} else {
		domain = fmt.Sprintf("tf-acc-multiple-issued.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))
	}

	resourceName := "data.aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateDataSourceConfig_basic(domain),
				ExpectError: regexache.MustCompile(`multiple ACM Certificates matching domain`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_status(domain, string(awstypes.CertificateStatusIssued)),
				ExpectError: regexache.MustCompile(`multiple ACM Certificates matching domain`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_types(domain, string(awstypes.CertificateTypeAmazonIssued)),
				ExpectError: regexache.MustCompile(`multiple ACM Certificates matching domain`),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecent(domain, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndStatus(domain, string(awstypes.CertificateStatusIssued), true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndTypes(domain, string(awstypes.CertificateTypeAmazonIssued), true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_noMatchReturnsError(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	domain := fmt.Sprintf("tf-acc-nonexistent.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateDataSourceConfig_basic(domain),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching domain`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_status(domain, string(awstypes.CertificateStatusIssued)),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching domain`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_types(domain, string(awstypes.CertificateTypeAmazonIssued)),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching domain`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecent(domain, true),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching domain`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecentAndStatus(domain, string(awstypes.CertificateStatusIssued), true),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching domain`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecentAndTypes(domain, string(awstypes.CertificateTypeAmazonIssued), true),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching domain`),
			},
		},
	})
}

func TestAccACMCertificateDataSource_keyTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 4096)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomDomain().String())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_keyTypes(acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccCertificateDataSourceConfig_basic(domain string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain = %[1]q
}
`, domain)
}

func testAccCertificateDataSourceConfig_status(domain, status string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain   = %[1]q
  statuses = [%[2]q]
}
`, domain, status)
}

func testAccCertificateDataSourceConfig_types(domain, certType string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain = %[1]q
  types  = [%[2]q]
}
`, domain, certType)
}

func testAccCertificateDataSourceConfig_mostRecent(domain string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = %[1]q
  most_recent = %[2]t
}
`, domain, mostRecent)
}

func testAccCertificateDataSourceConfig_mostRecentAndStatus(domain, status string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = %[1]q
  statuses    = [%[2]q]
  most_recent = %[3]t
}
`, domain, status, mostRecent)
}

func testAccCertificateDataSourceConfig_mostRecentAndTypes(domain, certType string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = %[1]q
  types       = [%[2]q]
  most_recent = %[3]t
}
`, domain, certType, mostRecent)
}

func testAccCertificateDataSourceConfig_keyTypes(certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}

data "aws_acm_certificate" "test" {
  domain    = aws_acm_certificate.test.domain_name
  key_types = ["RSA_4096"]
}
`, certificate, key)
}
