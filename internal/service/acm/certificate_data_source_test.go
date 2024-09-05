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
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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
				ExpectError: regexache.MustCompile(`\d+ matching ACM Certificates found`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_status(domain, string(awstypes.CertificateStatusIssued)),
				ExpectError: regexache.MustCompile(`\d+ matching ACM Certificates found`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_types(domain, string(awstypes.CertificateTypeAmazonIssued)),
				ExpectError: regexache.MustCompile(`\d+ matching ACM Certificates found`),
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
				ExpectError: regexache.MustCompile(`no matching ACM Certificate found`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_status(domain, string(awstypes.CertificateStatusIssued)),
				ExpectError: regexache.MustCompile(`no matching ACM Certificate found`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_types(domain, string(awstypes.CertificateTypeAmazonIssued)),
				ExpectError: regexache.MustCompile(`no matching ACM Certificate found`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecent(domain, true),
				ExpectError: regexache.MustCompile(`no matching ACM Certificate found`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecentAndStatus(domain, string(awstypes.CertificateStatusIssued), true),
				ExpectError: regexache.MustCompile(`no matching ACM Certificate found`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecentAndTypes(domain, string(awstypes.CertificateTypeAmazonIssued), true),
				ExpectError: regexache.MustCompile(`no matching ACM Certificate found`),
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

func TestAccACMCertificateDataSource_byDomain(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byDomain(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainNoMatch(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateDataSourceConfig_byDomainNoMatch(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				ExpectError: regexache.MustCompile(`reading ACM Certificates: empty result`),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateDataSourceConfig_byDomainMultiple(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				ExpectError: regexache.MustCompile(`2 matching ACM Certificates found`),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainAndTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byDomainAndTags(domain, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainAndStatuses(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byDomainAndStatuses(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainAndKeyTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 4096)
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byDomainAndKeyTypes(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byTags(rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byTagsNoMatch(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomDomain().String())
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateDataSourceConfig_byTagsNoMatch(rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				ExpectError: regexache.MustCompile(`no matching ACM Certificate found`),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byTagsAndKeyTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 4096)
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byTagsAndKeyTypes(rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
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

func testAccCertificateDataSourceConfig_byDomain(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain = %[1]q

  depends_on = [aws_acm_certificate.test]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainNoMatch(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain = "not.%[1]s"

  depends_on = [aws_acm_certificate.test]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainMultiple(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test1" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_acm_certificate" "test2" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain = %[1]q

  depends_on = [aws_acm_certificate.test1, aws_acm_certificate.test2]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainAndTags(domain, rName, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[3]s"
  private_key      = "%[4]s"

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
    Name = %[2]q
  }
}

data "aws_acm_certificate" "test" {
  domain = %[1]q

  tags = {
    Key1 = "Value1"
    Name = aws_acm_certificate.test.tags["Name"]
  }
}
`, domain, rName, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainAndStatuses(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain   = %[1]q
  statuses = ["EXPIRED", "ISSUED"]

  depends_on = [aws_acm_certificate.test]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainAndKeyTypes(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain    = %[1]q
  key_types = ["RSA_4096"]

  depends_on = [aws_acm_certificate.test]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byTags(rName, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
    Name = %[1]q
  }
}

data "aws_acm_certificate" "test" {
  tags = {
    Key1 = "Value1"
    Name = aws_acm_certificate.test.tags["Name"]
  }
}
`, rName, certificate, key)
}

func testAccCertificateDataSourceConfig_byTagsNoMatch(rName, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
    Name = %[1]q
  }
}

data "aws_acm_certificate" "test" {
  tags = {
    Key1 = "Value1"
    Key3 = "Value3"
    Name = aws_acm_certificate.test.tags["Name"]
  }
}
`, rName, certificate, key)
}

func testAccCertificateDataSourceConfig_byTagsAndKeyTypes(rName, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  tags = {
    Name = %[1]q
  }
}

data "aws_acm_certificate" "test" {
  key_types = ["RSA_4096"]

  tags = {
    Name = aws_acm_certificate.test.tags["Name"]
  }
}
`, rName, certificate, key)
}
