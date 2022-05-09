package appsync_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
)

func testAccDomainNameAPIAssociation_basic(t *testing.T) {
	var providers []*schema.Provider
	var association appsync.ApiAssociation
	appsyncCertDomain := getCertDomain(t)

	rName := sdkacctest.RandString(8)
	resourceName := "aws_appsync_domain_name_api_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckDomainNameApiAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameAPIAssociationConfig_basic(appsyncCertDomain, rName),
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
				Config: testAccDomainNameAPIAssociationConfig_updated(appsyncCertDomain, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameApiAssociationExists(resourceName, &association),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", "aws_appsync_domain_name.test", "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test2", "id"),
				),
			},
		},
	})
}

func testAccDomainNameAPIAssociation_disappears(t *testing.T) {
	var association appsync.ApiAssociation
	var providers []*schema.Provider
	appsyncCertDomain := getCertDomain(t)

	rName := sdkacctest.RandString(8)
	resourceName := "aws_appsync_domain_name_api_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckDomainNameApiAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameAPIAssociationConfig_basic(appsyncCertDomain, rName),
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

func testAccDomainNameAPIAssociationBaseConfig(domain, rName string) string {
	return acctest.ConfigAlternateRegionProvider() + fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  provider    = "awsalternate"
  domain      = "*.%[1]s"
  most_recent = true
}

resource "aws_appsync_domain_name" "test" {
  domain_name     = "%[2]s.%[1]s"
  certificate_arn = data.aws_acm_certificate.test.arn
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[2]q
}
`, domain, rName)
}

func testAccDomainNameAPIAssociationConfig_basic(domain, rName string) string {
	return testAccDomainNameAPIAssociationBaseConfig(domain, rName) + `
resource "aws_appsync_domain_name_api_association" "test" {
  api_id      = aws_appsync_graphql_api.test.id
  domain_name = aws_appsync_domain_name.test.domain_name
}
`
}

func testAccDomainNameAPIAssociationConfig_updated(domain, rName string) string {
	return testAccDomainNameAPIAssociationBaseConfig(domain, rName) + fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test2" {
  authentication_type = "API_KEY"
  name                = "%[1]s-2"
}

resource "aws_appsync_domain_name_api_association" "test" {
  api_id      = aws_appsync_graphql_api.test2.id
  domain_name = aws_appsync_domain_name.test.domain_name
}
`, rName)
}
