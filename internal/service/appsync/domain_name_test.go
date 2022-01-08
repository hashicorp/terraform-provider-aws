package appsync_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
)

func testAccAppSyncDomainName_basic(t *testing.T) {
	var domainName appsync.DomainNameConfig
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	acmCertificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_appsync_domain_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, appsync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncDomainNameConfig(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", acmCertificateResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", acmCertificateResourceName, "arn"),
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

func testAccAppSyncDomainName_description(t *testing.T) {
	var domainName appsync.DomainNameConfig
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resourceName := "aws_appsync_domain_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, appsync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncDomainNameDescriptionConfig(rootDomain, domain, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAppsyncDomainNameDescriptionConfig(rootDomain, domain, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
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

func testAccAppSyncDomainName_disappears(t *testing.T) {
	var domainName appsync.DomainNameConfig
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resourceName := "aws_appsync_domain_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, appsync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncDomainNameConfig(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					acctest.CheckResourceDisappears(acctest.Provider, tfappsync.ResourceDomainName(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainNameDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_domain_name" {
			continue
		}

		domainName, err := tfappsync.FindDomainNameByID(conn, rs.Primary.ID)
		if err == nil {
			if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		if domainName != nil && aws.StringValue(domainName.DomainName) == rs.Primary.ID {
			return fmt.Errorf("Appsync Domain Name ID %q still exists", rs.Primary.ID)
		}

		return nil

	}
	return nil
}

func testAccCheckDomainNameExists(resourceName string, domainName *appsync.DomainNameConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Appsync Domain Name Not found in state: %s", resourceName)
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn

		domain, err := tfappsync.FindDomainNameByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if domain == nil || domain.DomainName == nil {
			return fmt.Errorf("Appsync Domain Name %q not found", rs.Primary.ID)
		}

		*domainName = *domain

		return nil
	}
}

func testAccAppsyncDomainNamePublicCertConfig(rootDomain, domain string) string {
	return fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}
`, rootDomain, domain)
}

func testAccAppsyncDomainNameDescriptionConfig(rootDomain, domain, desc string) string {
	return testAccAppsyncDomainNamePublicCertConfig(rootDomain, domain) + fmt.Sprintf(`
resource "aws_appsync_domain_name" "test" {
  domain_name     = aws_acm_certificate.test.domain_name
  certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  description     = %[1]q
}
`, desc)
}

func testAccAppsyncDomainNameConfig(rootDomain, domain string) string {
	return testAccAppsyncDomainNamePublicCertConfig(rootDomain, domain) + `
resource "aws_appsync_domain_name" "test" {
  domain_name     = aws_acm_certificate.test.domain_name
  certificate_arn = aws_acm_certificate_validation.test.certificate_arn
}
`
}
