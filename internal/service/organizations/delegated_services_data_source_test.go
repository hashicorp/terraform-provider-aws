// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDelegatedServicesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_organizations_delegated_services.test"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedServicesDataSourceConfig_basic(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "delegated_services.#", 1),
				),
			},
		},
	})
}

func testAccDelegatedServicesDataSource_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_organizations_delegated_services.test"
	servicePrincipal1 := "config-multiaccountsetup.amazonaws.com"
	servicePrincipal2 := "config.amazonaws.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedServicesDataSourceConfig_multiple(servicePrincipal1, servicePrincipal2),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "delegated_services.#", 2),
				),
			},
		},
	})
}

func testAccDelegatedServicesDataSourceConfig_basic(servicePrincipal string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

resource "aws_organizations_delegated_administrator" "delegated" {
  account_id        = data.aws_caller_identity.delegated.account_id
  service_principal = %[1]q
}

data "aws_organizations_delegated_services" "test" {
  account_id = data.aws_caller_identity.delegated.account_id

  depends_on = [aws_organizations_delegated_administrator.delegated]
}
`, servicePrincipal))
}

func testAccDelegatedServicesDataSourceConfig_multiple(servicePrincipal1, servicePrincipal2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
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
  account_id = data.aws_caller_identity.delegated.account_id

  depends_on = [aws_organizations_delegated_administrator.delegated, aws_organizations_delegated_administrator.other_delegated]
}
`, servicePrincipal1, servicePrincipal2))
}
