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
	"github.com/aws/aws-sdk-go/service/acm"
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
	var certType string

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

	if os.Getenv("ACM_CERTIFICATE_TYPE") != "" {
		certType = os.Getenv("ACM_CERTIFICATE_TYPE")
	} else {
		certType = acm.CertificateTypeAmazonIssued
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
				Config: testAccCertificateDataSourceConfig_types(domain, certType),
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
				Config: testAccCertificateDataSourceConfig_mostRecentAndTypes(domain, certType, true),
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
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, acm.CertificateStatusIssued),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_statusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, acm.CertificateStatusIssued),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_typesAndTags(domain, certType, tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndTags(domain, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndStatusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndTypesAndTags(domain, certType, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_basicTags(tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, acm.CertificateStatusIssued),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndStatus(acm.CertificateStatusIssued, tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, acm.CertificateStatusIssued),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndTypes(certType, tagName, tagValue),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecent(tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecentAndStatus(acm.CertificateStatusIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecentAndTypes(certType, tagName, tagValue, true),
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
	var tagName string
	var tagValue string
	var certType string

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

	if os.Getenv("ACM_CERTIFICATE_TYPE") != "" {
		certType = os.Getenv("ACM_CERTIFICATE_TYPE")
	} else {
		certType = acm.CertificateTypeAmazonIssued
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
				Config:      testAccCertificateDataSourceConfig_types(domain, certType),
				ExpectError: regexache.MustCompile(`multiple ACM Certificates matching search criteria`),
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
				Config: testAccCertificateDataSourceConfig_mostRecentAndTypes(domain, certType, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
				),
			},
			{
				Config:      testAccCertificateDataSourceConfig_basicAndTags(domain, tagName, tagValue),
				ExpectError: regexache.MustCompile(`multiple ACM Certificates matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_statusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue),
				ExpectError: regexache.MustCompile(`multiple ACM Certificates matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_typesAndTags(domain, certType, tagName, tagValue),
				ExpectError: regexache.MustCompile(`multiple ACM Certificates matching search criteria`),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndTags(domain, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndStatusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_mostRecentAndTypesAndTags(domain, certType, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
				),
			},
			{
				Config:      testAccCertificateDataSourceConfig_basicTags(tagName, tagValue),
				ExpectError: regexache.MustCompile(`multiple ACM Certificates matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndStatus(acm.CertificateStatusIssued, tagName, tagValue),
				ExpectError: regexache.MustCompile(`multiple ACM Certificates matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndTypes(certType, tagName, tagValue),
				ExpectError: regexache.MustCompile(`multiple ACM Certificates matching search criteria`),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecent(tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecentAndStatus(acm.CertificateStatusIssued, tagName, tagValue, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, names.AttrARN, arnRe),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_TagsAndMostRecentAndTypes(certType, tagName, tagValue, true),
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
	var certType string

	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	if os.Getenv("ACM_CERTIFICATE_TYPE") != "" {
		certType = os.Getenv("ACM_CERTIFICATE_TYPE")
	} else {
		certType = acm.CertificateTypeAmazonIssued
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
				Config:      testAccCertificateDataSourceConfig_types(domain, certType),
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
				Config:      testAccCertificateDataSourceConfig_mostRecentAndTypes(domain, certType, true),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching domain`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_basicAndTags(domain, tagName, tagValue),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_statusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_typesAndTags(domain, certType, tagName, tagValue),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecentAndTags(domain, tagName, tagValue, true),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecentAndStatusAndTags(domain, acm.CertificateStatusIssued, tagName, tagValue, true),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_mostRecentAndTypesAndTags(domain, certType, tagName, tagValue, true),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_basicTags(tagName, tagValue),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndStatus(acm.CertificateStatusIssued, tagName, tagValue),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndTypes(certType, tagName, tagValue),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndMostRecent(tagName, tagValue, true),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndMostRecentAndStatus(acm.CertificateStatusIssued, tagName, tagValue, true),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_TagsAndMostRecentAndTypes(certType, tagName, tagValue, true),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_certTypes(certType),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria. Please use at least domain or tags as search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_certStatus(acm.CertificateStatusIssued),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria. Please use at least domain or tags as search criteria`),
			},
			{
				Config:      testAccCertificateDataSourceConfig_certStatusTypes(acm.CertificateStatusIssued, certType),
				ExpectError: regexache.MustCompile(`no ACM Certificate matching search criteria. Please use at least domain or tags as search criteria`),
			},
		},
	})
}

func TestAccACMCertificateDataSource_keyTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 4096)
	domainName := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domainName)
	tagName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.9.1",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_keyTypes(acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTags, dataSourceName, names.AttrTags),
				),
			},
			{
				Config: testAccCertificateDataSourceConfig_keyTypesAndTags(acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), tagName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTags, dataSourceName, names.AttrTags),
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
  tags = {
    %[1]q = %[2]q
  }
}
`, key, value)
}

func testAccCertificateDataSourceConfig_TagsAndStatus(status, key, value string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  statuses = [%[1]q]
  tags = {
    %[2]q = %[3]q
  }
}
`, status, key, value)
}

func testAccCertificateDataSourceConfig_TagsAndTypes(certType, key, value string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  tags = {
    %[2]q = %[3]q
  }
  types = [%[1]q]
}
`, certType, key, value)
}

func testAccCertificateDataSourceConfig_TagsAndMostRecent(key, value string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  most_recent = %[1]t
  tags = {
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
  tags = {
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
  tags = {
    %[3]q = %[4]q
  }
}
`, certType, mostRecent, key, value)
}

func testAccCertificateDataSourceConfig_basicAndTags(domain, key, value string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain = %[1]q
  tags = {
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
  tags = {
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
  tags = {
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
  tags = {
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
  tags = {
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
  tags = {
    %[4]q = %[5]q
  }
}
`, domain, certType, mostRecent, key, value)
}

func testAccCertificateDataSourceConfig_certTypes(certType string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  types = [%[1]q]
}
`, certType)
}

func testAccCertificateDataSourceConfig_certStatus(certType string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  statuses = [%[1]q]
}
`, certType)
}

func testAccCertificateDataSourceConfig_certStatusTypes(status, certType string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  statuses = [%[1]q]
  types    = [%[2]q]
}
`, status, certType)
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

resource "time_sleep" "wait_1_seconds" {
  depends_on = [aws_acm_certificate.test]

  create_duration = "1s"
}

data "aws_acm_certificate" "test" {
  tags = {
    %[3]q = aws_acm_certificate.test.domain_name
  }
  key_types  = ["RSA_4096"]
  depends_on = [time_sleep.wait_1_seconds]
}
`, certificate, key, tagName, tagValue)
}
