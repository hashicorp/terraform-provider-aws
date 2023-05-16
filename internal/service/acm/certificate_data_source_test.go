// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acm_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/acm"
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
	var tagName string
	var tagValue string

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

	if os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_TAG_NAME") != "" && os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_TAG_VALUE") != "" {
		tagName = os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_TAG_NAME")
		tagValue = os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_TAG_VALUE")
	} else {
		tagName = "Name"
		tagValue = fmt.Sprintf("tf-acc-single-issued.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))
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
			{
				Config: testAccCertificateDataSourceConfig_basicAndTags(domain, tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_statusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_typesAndTags(domain, acm.CertificateTypeAmazonIssued, tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndTags(domain, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndStatusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndTypesAndTags(domain, acm.CertificateTypeAmazonIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_basicTags(tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndStatus(acm.CertificateStatusIssued, tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndTypes(acm.CertificateTypeAmazonIssued, tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecent(tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecentAndStatus(acm.CertificateStatusIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecentAndTypes(acm.CertificateTypeAmazonIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
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
	var tagName string
	var tagValue string

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

	if os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_TAG_NAME") != "" && os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_TAG_VALUE") != "" {
		tagName = os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_TAG_NAME")
		tagValue = os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_TAG_VALUE")
	} else {
		tagName = "Name"
		tagValue = fmt.Sprintf("tf-acc-multiple-issued.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))
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
			{
				Config:      testAccCertificateDataSourceConfig_basicAndTags(domain, tagName, tagValue),
				ExpectError: regexp.MustCompile(`multiple ACM Certificates matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_statusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue),
				ExpectError: regexp.MustCompile(`multiple ACM Certificates matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_typesAndTags(domain, acm.CertificateTypeAmazonIssued, tagName, tagValue),
				ExpectError: regexp.MustCompile(`multiple ACM Certificates matching input criteria`),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndTags(domain, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndStatusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndTypesAndTags(domain, acm.CertificateTypeAmazonIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config:      testAccCertificateDataSourceConfig_basicTags(tagName, tagValue),
				ExpectError: regexp.MustCompile(`multiple ACM Certificates matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndStatus(acm.CertificateStatusIssued, tagName, tagValue),
				ExpectError: regexp.MustCompile(`multiple ACM Certificates matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndTypes(acm.CertificateTypeAmazonIssued, tagName, tagValue),
				ExpectError: regexp.MustCompile(`multiple ACM Certificates matching input criteria`),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecent(tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecentAndStatus(acm.CertificateStatusIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecentAndTypes(acm.CertificateTypeAmazonIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
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
	tagName := fmt.Sprintf("tf-acc-nonexistent.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))
	tagValue := fmt.Sprintf("tf-acc-nonexistent.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))

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
			{
				Config:      testAccCertificateDataSourceConfig_basicAndTags(domain, tagName, tagValue),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_statusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_typesAndTags(domain, acm.CertificateTypeAmazonIssued, tagName, tagValue),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecentAndTags(domain, tagName, tagValue, true),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecentAndStatusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue, true),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecentAndTypesAndTags(domain, acm.CertificateTypeAmazonIssued, tagName, tagValue, true),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_basicTags(tagName, tagValue),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndStatus(acm.CertificateStatusIssued, tagName, tagValue),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndTypes(acm.CertificateTypeAmazonIssued, tagName, tagValue),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndMostRecent(tagName, tagValue, true),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndMostRecentAndStatus(acm.CertificateStatusIssued, tagName, tagValue, true),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndMostRecentAndTypes(acm.CertificateTypeAmazonIssued, tagName, tagValue, true),
				ExpectError: regexp.MustCompile(`no ACM Certificate found matching input criteria`),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tagName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tagValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_keyTypes(acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTags, dataSourceName, names.AttrTags),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_keyTypesAndTags(acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
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

func testAccCertificateDataSourceConfig_keyTypes(certificate, key, rName string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"

  tags = {
    Name = %[3]q
  }
}

data "aws_acm_certificate" "test" {
  domain    = aws_acm_certificate.test.domain_name
  key_types = ["RSA_4096"]
}
`, certificate, key, rName)
}

func testAccCertificateDataSourceConfig_basicTags(key, value string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  tags   = {
	%[1]q = %[2]q
  }
}
`, key, value)
}

func testAccCertificateDataSourceConfig_TagsAndStatus(status, key, value string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  statuses = [%[1]q]
  tags     = {
	%[2]q = %[3]q
  }
}
`, status, key, value)
}

func testAccCertificateDataSourceConfig_TagsAndTypes(certType, key, value string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  types  = [%[2]q]
  tags   = {
	%[2]q = %[3]q
  }
}
`, certType, key, value)
}

func testAccCertificateDataSourceConfig_TagsAndMostRecent(key, value string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  most_recent = %[1]t
  tags        = {
	%[2]q = %[3]q
  }
}
`, mostRecent, key, value)
}

func testAccCertificateDataSourceConfig_TagsAndMostRecentAndStatus(status, key, value string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  statuses    = [%[1]q]
  most_recent = %[2]t
  tags 		  = {
	%[3]q = %[4]q
  }
}
`, status, mostRecent, key, value)
}

func testAccCertificateDataSourceConfig_TagsAndMostRecentAndTypes(certType, key, value string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  types       = [%[1]q]
  most_recent = %[2]t
  tags 		  = {
	%[3]q = %[4]q
  }
}
`, certType, mostRecent, key, value)
}

func testAccCertificateDataSourceConfig_basicAndTags(domain, key, value string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain = %[1]q
  tags   = {
	%[2]q = %[3]q
  }
}
`, domain, key, value)
}

func testAccCertificateDataSourceConfig_statusAndTags(domain, status, key, value string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain   = %[1]q
  statuses = [%[2]q]
  tags     = {
	%[3]q = %[4]q
  }
}
`, domain, status, key, value)
}

func testAccCertificateDataSourceConfig_typesAndTags(domain, certType, key, value string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain = %[1]q
  types  = [%[2]q]
  tags   = {
	%[3]q = %[4]q
  }
}
`, domain, certType, key, value)
}

func testAccCertificateDataSourceConfig_mostRecentAndTags(domain, key, value string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = %[1]q
  most_recent = %[2]t
  tags        = {
	%[3]q = %[4]q
  }
}
`, domain, mostRecent, key, value)
}

func testAccCertificateDataSourceConfig_mostRecentAndStatusAndTags(domain, status, key, value string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = %[1]q
  statuses    = [%[2]q]
  most_recent = %[3]t
  tags 		  = {
	%[4]q = %[5]q
  }
}
`, domain, status, mostRecent, key, value)
}

func testAccCertificateDataSourceConfig_mostRecentAndTypesAndTags(domain, certType, key, value string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = %[1]q
  types       = [%[2]q]
  most_recent = %[3]t
  tags 		  = {
	%[4]q = %[5]q
  }
}
`, domain, certType, mostRecent, key, value)
}

func testAccCertificateDataSourceConfig_keyTypesAndTags(certificate, key, tagName, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"

  tags = {
    %[3]q = %[4]q
  }
}

data "aws_acm_certificate" "test" {
  key_types = ["RSA_4096"]
  tags 		= {
	%[3]q = %[4]q
  }
}
`, certificate, key, tagName, tagValue)
}
