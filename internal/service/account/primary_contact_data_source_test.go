// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package account_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPrimaryContactDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_primary_contact.test"
	dataSourceName := "data.aws_account_primary_contact.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccPrimaryContactDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "address_line_1", resourceName, "address_line_1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "address_line_2", resourceName, "address_line_2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "address_line_3", resourceName, "address_line_3"),
					resource.TestCheckResourceAttrPair(dataSourceName, "city", resourceName, "city"),
					resource.TestCheckResourceAttrPair(dataSourceName, "company_name", resourceName, "company_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "country_code", resourceName, "country_code"),
					resource.TestCheckResourceAttrPair(dataSourceName, "district_or_county", resourceName, "district_or_county"),
					resource.TestCheckResourceAttrPair(dataSourceName, "full_name", resourceName, "full_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "phone_number", resourceName, "phone_number"),
					resource.TestCheckResourceAttrPair(dataSourceName, "postal_code", resourceName, "postal_code"),
					resource.TestCheckResourceAttrPair(dataSourceName, "state_or_region", resourceName, "state_or_region"),
					resource.TestCheckResourceAttrPair(dataSourceName, "website_url", resourceName, "website_url"),
				),
			},
		},
	})
}

func testAccPrimaryContactDataSource_accountID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_primary_contact.test"
	dataSourceName := "data.aws_account_primary_contact.test"

	acctest.Test(ctx, t, resource.TestCase{
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
				Config: testAccPrimaryContactDataSourceConfig_organization(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAccountID, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(dataSourceName, "address_line_1", resourceName, "address_line_1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "address_line_2", resourceName, "address_line_2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "address_line_3", resourceName, "address_line_3"),
					resource.TestCheckResourceAttrPair(dataSourceName, "city", resourceName, "city"),
					resource.TestCheckResourceAttrPair(dataSourceName, "company_name", resourceName, "company_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "country_code", resourceName, "country_code"),
					resource.TestCheckResourceAttrPair(dataSourceName, "district_or_county", resourceName, "district_or_county"),
					resource.TestCheckResourceAttrPair(dataSourceName, "full_name", resourceName, "full_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "phone_number", resourceName, "phone_number"),
					resource.TestCheckResourceAttrPair(dataSourceName, "postal_code", resourceName, "postal_code"),
					resource.TestCheckResourceAttrPair(dataSourceName, "state_or_region", resourceName, "state_or_region"),
					resource.TestCheckResourceAttrPair(dataSourceName, "website_url", resourceName, "website_url"),
				),
			},
		},
	})
}

func testAccPrimaryContactDataSourceConfig_basic() string {
	return `
resource "aws_account_primary_contact" "test" {
  address_line_1     = "123 Any Street"
  address_line_2     = "234 Any Street"
  address_line_3     = "345 Any Street"
  city               = "Seattle"
  company_name       = "Example Corp, Inc."
  country_code       = "US"
  district_or_county = "King"
  full_name          = "Foo Bar"
  phone_number       = "+64211111111"
  postal_code        = "98101"
  state_or_region    = "WA"
  website_url        = "https://www.examplecorp.com"
}

data "aws_account_primary_contact" "test" {
  depends_on = [aws_account_primary_contact.test]
}
`
}

func testAccPrimaryContactDataSourceConfig_organization() string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), `
data "aws_caller_identity" "test" {
  provider = "awsalternate"
}

resource "aws_account_primary_contact" "test" {
  account_id         = data.aws_caller_identity.test.account_id
  address_line_1     = "123 Any Street"
  address_line_2     = "234 Any Street"
  address_line_3     = "345 Any Street"
  city               = "Seattle"
  company_name       = "Example Corp, Inc."
  country_code       = "US"
  district_or_county = "King"
  full_name          = "Foo Bar"
  phone_number       = "+64211111111"
  postal_code        = "98101"
  state_or_region    = "WA"
  website_url        = "https://www.examplecorp.com"
}

data "aws_account_primary_contact" "test" {
  account_id = data.aws_caller_identity.test.account_id
  depends_on = [aws_account_primary_contact.test]
}
`)
}
