package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOrganizationsDelegatedServicesDataSource_basic(t *testing.T) {
	var providers []*schema.Provider
	dataSourceName := "data.aws_organizations_delegated_services.test"
	dataSourceIdentity := "data.aws_caller_identity.delegated"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedServicesDataSourceConfig_basic(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "delegated_services.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", dataSourceIdentity, "account_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "delegated_services.0.delegation_enabled_date"),
					resource.TestCheckResourceAttr(dataSourceName, "delegated_services.0.service_principal", servicePrincipal),
				),
			},
		},
	})
}

func TestAccOrganizationsDelegatedServicesDataSource_empty(t *testing.T) {
	var providers []*schema.Provider
	dataSourceName := "data.aws_organizations_delegated_services.test"
	dataSourceIdentity := "data.aws_caller_identity.delegated"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedServicesDataSourceConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "delegated_services.#", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", dataSourceIdentity, "account_id"),
				),
			},
		},
	})
}

func TestAccOrganizationsDelegatedServicesDataSource_multiple(t *testing.T) {
	var providers []*schema.Provider
	dataSourceName := "data.aws_organizations_delegated_services.test"
	dataSourceIdentity := "data.aws_caller_identity.delegated"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"
	servicePrincipal2 := "config.amazonaws.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedServicesDataSourceConfig_multiple(servicePrincipal, servicePrincipal2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "delegated_services.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", dataSourceIdentity, "account_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "delegated_services.0.delegation_enabled_date"),
					resource.TestCheckResourceAttr(dataSourceName, "delegated_services.0.service_principal", servicePrincipal),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "delegated_services.1.delegation_enabled_date"),
					resource.TestCheckResourceAttr(dataSourceName, "delegated_services.1.service_principal", servicePrincipal2),
				),
			},
		},
	})
}

func testAccDelegatedServicesDataSourceConfig_empty() string {
	return acctest.ConfigAlternateAccountProvider() + `
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

data "aws_organizations_delegated_services" "test" {
  account_id = data.aws_caller_identity.delegated.account_id
}
`
}

func testAccDelegatedServicesDataSourceConfig_basic(servicePrincipal string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
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

func testAccDelegatedServicesDataSourceConfig_multiple(servicePrincipal, servicePrincipal2 string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

resource "aws_organizations_delegated_administrator" "delegated" {
  account_id        = data.aws_caller_identity.delegated.account_id
  service_principal = %[1]q
}

resource "aws_organizations_delegated_administrator" "other_delegated" {
  account_id        = data.aws_caller_identity.delegated.account_id
  service_principal = %[2]q
}

data "aws_organizations_delegated_services" "test" {
  account_id = aws_organizations_delegated_administrator.other_delegated.account_id
}
`, servicePrincipal, servicePrincipal2)
}
