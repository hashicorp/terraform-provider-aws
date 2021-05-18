package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDataSourceAwsOrganizationsDelegatedServices_basic(t *testing.T) {
	var providers []*schema.Provider
	dataSourceName := "data.aws_organizations_delegated_services.test"
	dataSourceIdentity := "data.aws_caller_identity.delegated"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ErrorCheck:        testAccErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationsDelegatedServicesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "delegated_services.#", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", dataSourceIdentity, "account_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsOrganizationsDelegatedServices_administrator(t *testing.T) {
	var providers []*schema.Provider
	dataSourceName := "data.aws_organizations_delegated_services.test"
	dataSourceIdentity := "data.aws_caller_identity.delegated"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ErrorCheck:        testAccErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationsDelegatedServicesAdministratorConfig(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "delegated_services.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", dataSourceIdentity, "account_id"),
					testAccCheckResourceAttrRfc3339(dataSourceName, "delegated_services.0.delegation_enabled_date"),
					resource.TestCheckResourceAttr(dataSourceName, "delegated_services.0.service_principal", servicePrincipal),
				),
			},
		},
	})
}

func testAccDataSourceAwsOrganizationsDelegatedServicesConfig() string {
	return testAccAlternateAccountProviderConfig() + `
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

data "aws_organizations_delegated_services" "test" {
  account_id = data.aws_caller_identity.delegated.account_id
}
`
}

func testAccDataSourceAwsOrganizationsDelegatedServicesAdministratorConfig(servicePrincipal string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

resource "aws_organizations_delegated_administrator" "delegated" {
  account_id        = data.aws_caller_identity.delegated.account_id
  service_principal = %[1]q
}

data "aws_organizations_delegated_services" "test" {
  account_id = aws_organizations_delegated_administrator.delegated.account_id
}
`, servicePrincipal)
}
