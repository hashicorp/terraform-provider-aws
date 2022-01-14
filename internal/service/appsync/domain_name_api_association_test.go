package appsync_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
)

func testAccAppSyncDomainNameApiAssociation_basic(t *testing.T) {
	var association appsync.ApiAssociation
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resourceName := "aws_appsync_domain_name.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, appsync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainNameApiAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncDomainNameApiAssociationConfig(rootDomain, domain, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameApiAssociationExists(resourceName, &association),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", "aws_appsync_domain_name.test", "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppsyncDomainNameApiAssociationUpdatedConfig(rootDomain, domain, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameApiAssociationExists(resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", "aws_appsync_domain_name.test", "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test2", "id"),
				),
			},
		},
	})
}

func testAccAppSyncDomainNameApiAssociation_disappears(t *testing.T) {
	var association appsync.ApiAssociation
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_appsync_domain_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, appsync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainNameApiAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncDomainNameApiAssociationConfig(rootDomain, domain, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameApiAssociationExists(resourceName, &association),
					acctest.CheckResourceDisappears(acctest.Provider, tfappsync.ResourceDomainNameApiAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainNameApiAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_domain_name" {
			continue
		}

		association, err := tfappsync.FindDomainNameApiAssociationByID(conn, rs.Primary.ID)
		if err == nil {
			if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		if association != nil && aws.StringValue(association.DomainName) == rs.Primary.ID {
			return fmt.Errorf("Appsync Domain Name ID %q still exists", rs.Primary.ID)
		}

		return nil

	}
	return nil
}

func testAccCheckDomainNameApiAssociationExists(resourceName string, DomainNameApiAssociation *appsync.ApiAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Appsync Domain Name Not found in state: %s", resourceName)
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn

		association, err := tfappsync.FindDomainNameApiAssociationByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if association == nil || association.DomainName == nil {
			return fmt.Errorf("Appsync Domain Name %q not found", rs.Primary.ID)
		}

		*DomainNameApiAssociation = *association

		return nil
	}
}

func testAccAppsyncDomainNameApiAssociationBaseConfig(rootDomain, domain string) string {
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

resource "aws_appsync_domain_name" "test" {
  domain_name     = aws_acm_certificate.test.domain_name
  certificate_arn = aws_acm_certificate_validation.test.certificate_arn
}
`, rootDomain, domain)
}

func testAccAppsyncDomainNameApiAssociationConfig(rootDomain, domain, rName string) string {
	return testAccAppsyncDomainNameApiAssociationBaseConfig(rootDomain, domain) + fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_api_association" "test" {
  api_id      = aws_appsync_graphql_api.test.id
  domain_name = aws_appsync_domain_name.test.domain_name
}
`, rName)
}

func testAccAppsyncDomainNameApiAssociationUpdatedConfig(rootDomain, domain, rName string) string {
	return testAccAppsyncDomainNameApiAssociationBaseConfig(rootDomain, domain) + fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_graphql_api" "test2" {
  authentication_type = "API_KEY"
  name                = "%[1]s-2"
}

resource "aws_appsync_api_association" "test" {
  api_id      = aws_appsync_graphql_api.test2.id
  domain_name = aws_appsync_domain_name.test.domain_name
}
`, rName)
}
