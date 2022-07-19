package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOrganizationsDelegatedAdministratorsDataSource_basic(t *testing.T) {
	var providers []*schema.Provider
	dataSourceName := "data.aws_organizations_delegated_administrators.test"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"
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
				Config: testAccDelegatedAdministratorsDataSourceConfig_basic(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "delegated_administrators.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "delegated_administrators.0.id", dataSourceIdentity, "account_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "delegated_administrators.0.delegation_enabled_date"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "delegated_administrators.0.joined_timestamp"),
				),
			},
		},
	})
}

func TestAccOrganizationsDelegatedAdministratorsDataSource_multiple(t *testing.T) {
	var providers []*schema.Provider
	dataSourceName := "data.aws_organizations_delegated_administrators.test"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"
	servicePrincipal2 := "config.amazonaws.com"
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
				Config: testAccDelegatedAdministratorsDataSourceConfig_multiple(servicePrincipal, servicePrincipal2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "delegated_administrators.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "delegated_administrators.0.id", dataSourceIdentity, "account_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "delegated_administrators.0.delegation_enabled_date"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "delegated_administrators.0.joined_timestamp"),
				),
			},
		},
	})
}

func TestAccOrganizationsDelegatedAdministratorsDataSource_servicePrincipal(t *testing.T) {
	var providers []*schema.Provider
	dataSourceName := "data.aws_organizations_delegated_administrators.test"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"
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
				Config: testAccDelegatedAdministratorsDataSourceConfig_servicePrincipal(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "delegated_administrators.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "delegated_administrators.0.id", dataSourceIdentity, "account_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "delegated_administrators.0.delegation_enabled_date"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "delegated_administrators.0.joined_timestamp"),
				),
			},
		},
	})
}

func TestAccOrganizationsDelegatedAdministratorsDataSource_empty(t *testing.T) {
	dataSourceName := "data.aws_organizations_delegated_administrators.test"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedAdministratorsDataSourceConfig_empty(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "delegated_administrators.#", "0"),
				),
			},
		},
	})
}

func testAccDelegatedAdministratorsDataSourceConfig_empty(servicePrincipal string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_organizations_delegated_administrators" "test" {
  service_principal = %[1]q
}
`, servicePrincipal)
}

func testAccDelegatedAdministratorsDataSourceConfig_basic(servicePrincipal string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

resource "aws_organizations_delegated_administrator" "test" {
  account_id        = data.aws_caller_identity.delegated.account_id
  service_principal = %[1]q
}

data "aws_organizations_delegated_administrators" "test" {}
`, servicePrincipal)
}

func testAccDelegatedAdministratorsDataSourceConfig_multiple(servicePrincipal, servicePrincipal2 string) string {
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

data "aws_organizations_delegated_administrators" "test" {}
`, servicePrincipal, servicePrincipal2)
}

func testAccDelegatedAdministratorsDataSourceConfig_servicePrincipal(servicePrincipal string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

resource "aws_organizations_delegated_administrator" "test" {
  account_id        = data.aws_caller_identity.delegated.account_id
  service_principal = %[1]q
}

data "aws_organizations_delegated_administrators" "test" {
  service_principal = aws_organizations_delegated_administrator.test.service_principal
}
`, servicePrincipal)
}
