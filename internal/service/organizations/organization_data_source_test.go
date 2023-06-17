package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Creates an new organization so that we are its management account.
func testAccOrganizationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_organizations_organization.test"
	dataSourceName := "data.aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDataSourceConfig_newOrganization,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "accounts.#", dataSourceName, "accounts.#"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "aws_service_access_principals.#", dataSourceName, "aws_service_access_principals.#"),
					resource.TestCheckResourceAttrPair(resourceName, "enabled_policy_types.#", dataSourceName, "enabled_policy_types.#"),
					resource.TestCheckResourceAttrPair(resourceName, "feature_set", dataSourceName, "feature_set"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "master_account_arn", dataSourceName, "master_account_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "master_account_email", dataSourceName, "master_account_email"),
					resource.TestCheckResourceAttrPair(resourceName, "master_account_id", dataSourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.#", dataSourceName, "non_master_accounts.#"),
					resource.TestCheckResourceAttrPair(resourceName, "roots.#", dataSourceName, "roots.#"),
				),
			},
		},
	})
}

// Runs as a member account in an existing organization.
// Certain attributes won't be set.
func testAccOrganizationDataSource_memberAccount(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(dataSourceName, "accounts.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
					resource.TestCheckNoResourceAttr(dataSourceName, "aws_service_access_principals.#"),
					resource.TestCheckNoResourceAttr(dataSourceName, "enabled_policy_types.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "feature_set"),
					resource.TestCheckResourceAttrSet(dataSourceName, "master_account_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "master_account_email"),
					resource.TestCheckResourceAttrSet(dataSourceName, "master_account_id"),
					resource.TestCheckNoResourceAttr(dataSourceName, "non_master_accounts.#"),
					resource.TestCheckNoResourceAttr(dataSourceName, "roots.#"),
				),
			},
		},
	})
}

// Runs as a management account in an existing organization.
// Creates a delegated administrator account and runs the data source under that account.
// All attributes will be set.
func testAccOrganizationDataSource_delegatedAdministrator(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_organizations_organization.test"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDataSourceConfig_delegatedAdministrator(servicePrincipal),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "accounts.#", 2),
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "aws_service_access_principals.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "enabled_policy_types.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "feature_set"),
					resource.TestCheckResourceAttrSet(dataSourceName, "master_account_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "master_account_email"),
					resource.TestCheckResourceAttrSet(dataSourceName, "master_account_id"),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "non_master_accounts.#", 1),
					resource.TestCheckResourceAttrSet(dataSourceName, "roots.#"),
				),
			},
		},
	})
}

const testAccOrganizationDataSourceConfig_newOrganization = `
resource "aws_organizations_organization" "test" {}

data "aws_organizations_organization" "test" {
  depends_on = [aws_organizations_organization.test]
}
`

const testAccOrganizationDataSourceConfig_basic = `
data "aws_organizations_organization" "test" {}
`

func testAccOrganizationDataSourceConfig_delegatedAdministrator(servicePrincipal string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

resource "aws_organizations_delegated_administrator" "test" {
  account_id        = data.aws_caller_identity.delegated.account_id
  service_principal = %[1]q
}

data "aws_organizations_organization" "test" {
  provider = "awsalternate"

  depends_on = [aws_organizations_delegated_administrator.test]
}
`, servicePrincipal))
}
