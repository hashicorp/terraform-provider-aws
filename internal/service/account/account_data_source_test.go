// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountDataSource_basic(t *testing.T) { // nosemgrep:ci.account-in-func-name
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_account_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrSet(dataSourceName, "account_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "account_created_date"),
				),
			},
		},
	})
}

func testAccAccountDataSource_accountID(t *testing.T) { // nosemgrep:ci.account-in-func-name
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_account_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "account.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataSourceConfig_organization(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrSet(dataSourceName, "account_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "account_created_date"),
				),
			},
		},
	})
}

func testAccAccountDataSourceConfig_basic() string { // nosemgrep:ci.account-in-func-name
	return `
data "aws_account_account" "test" {}
`
}

func testAccAccountDataSourceConfig_organization() string { // nosemgrep:ci.account-in-func-name
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), `
data "aws_caller_identity" "test" {
  provider = "awsalternate"
}

data "aws_account_account" "test" {
  account_id = data.aws_caller_identity.test.account_id
}
`)
}
