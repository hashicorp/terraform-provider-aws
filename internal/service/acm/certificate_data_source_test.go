package acm_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/acm"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

const ACMCertificateRe = `^arn:[^:]+:acm:[^:]+:[^:]+:certificate/.+$`

func TestAccACMCertificateDataSource_singleIssued(t *testing.T) {
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	var arnRe *regexp.Regexp
	var domain string

	if os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_MOST_RECENT_ARN") != "" {
		arnRe = regexp.MustCompile(fmt.Sprintf("^%s$", os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_MOST_RECENT_ARN")))
	} else {
		arnRe = regexp.MustCompile(ACMCertificateRe)
	}

	if os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_DOMAIN") != "" {
		domain = os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_DOMAIN")
	} else {
		domain = fmt.Sprintf("tf-acc-single-issued.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))
	}

	resourceName := "data.aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCertificateDataSourceConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCheckCertificateWithStatusDataSourceConfig(domain, acm.CertificateStatusIssued),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCheckCertificateWithTypesDataSourceConfig(domain, acm.CertificateTypeAmazonIssued),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCheckCertificateWithMostRecentDataSourceConfig(domain, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCheckCertificateWithMostRecentAndStatusDataSourceConfig(domain, acm.CertificateStatusIssued, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_chain"),
				),
			},
			{
				Config: testAccCheckCertificateWithMostRecentAndTypesDataSourceConfig(domain, acm.CertificateTypeAmazonIssued, true),
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
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	var arnRe *regexp.Regexp
	var domain string

	if os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_MOST_RECENT_ARN") != "" {
		arnRe = regexp.MustCompile(fmt.Sprintf("^%s$", os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_MOST_RECENT_ARN")))
	} else {
		arnRe = regexp.MustCompile(ACMCertificateRe)
	}

	if os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_DOMAIN") != "" {
		domain = os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_DOMAIN")
	} else {
		domain = fmt.Sprintf("tf-acc-multiple-issued.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))
	}

	resourceName := "data.aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCertificateDataSourceConfig(domain),
				ExpectError: regexp.MustCompile(`Multiple certificates for domain`),
			},
			{
				Config:      testAccCheckCertificateWithStatusDataSourceConfig(domain, acm.CertificateStatusIssued),
				ExpectError: regexp.MustCompile(`Multiple certificates for domain`),
			},
			{
				Config:      testAccCheckCertificateWithTypesDataSourceConfig(domain, acm.CertificateTypeAmazonIssued),
				ExpectError: regexp.MustCompile(`Multiple certificates for domain`),
			},
			{
				Config: testAccCheckCertificateWithMostRecentDataSourceConfig(domain, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCheckCertificateWithMostRecentAndStatusDataSourceConfig(domain, acm.CertificateStatusIssued, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCheckCertificateWithMostRecentAndTypesDataSourceConfig(domain, acm.CertificateTypeAmazonIssued, true),
				Check: resource.ComposeTestCheckFunc(
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_noMatchReturnsError(t *testing.T) {
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	domain := fmt.Sprintf("tf-acc-nonexistent.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCertificateDataSourceConfig(domain),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
			{
				Config:      testAccCheckCertificateWithStatusDataSourceConfig(domain, acm.CertificateStatusIssued),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
			{
				Config:      testAccCheckCertificateWithTypesDataSourceConfig(domain, acm.CertificateTypeAmazonIssued),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
			{
				Config:      testAccCheckCertificateWithMostRecentDataSourceConfig(domain, true),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
			{
				Config:      testAccCheckCertificateWithMostRecentAndStatusDataSourceConfig(domain, acm.CertificateStatusIssued, true),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
			{
				Config:      testAccCheckCertificateWithMostRecentAndTypesDataSourceConfig(domain, acm.CertificateTypeAmazonIssued, true),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
		},
	})
}

func TestAccACMCertificateDataSource_keyTypes(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(4096)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, acctest.RandomDomain().String())
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateKeyTypesDataSourceConfig(acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func testAccCheckCertificateDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain = "%s"
}
`, domain)
}

func testAccCheckCertificateWithStatusDataSourceConfig(domain, status string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain   = "%s"
  statuses = ["%s"]
}
`, domain, status)
}

func testAccCheckCertificateWithTypesDataSourceConfig(domain, certType string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain = "%s"
  types  = ["%s"]
}
`, domain, certType)
}

func testAccCheckCertificateWithMostRecentDataSourceConfig(domain string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = "%s"
  most_recent = %t
}
`, domain, mostRecent)
}

func testAccCheckCertificateWithMostRecentAndStatusDataSourceConfig(domain, status string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = "%s"
  statuses    = ["%s"]
  most_recent = %t
}
`, domain, status, mostRecent)
}

func testAccCheckCertificateWithMostRecentAndTypesDataSourceConfig(domain, certType string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = "%s"
  types       = ["%s"]
  most_recent = %t
}
`, domain, certType, mostRecent)
}

func testAccCertificateKeyTypesDataSourceConfig(certificate, key, rName string) string {
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
