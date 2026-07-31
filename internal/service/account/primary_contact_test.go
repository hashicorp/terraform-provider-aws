// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package account_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfaccount "github.com/hashicorp/terraform-provider-aws/internal/service/account"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPrimaryContact_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_primary_contact.test"
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccPrimaryConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrimaryContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, ""),
					resource.TestCheckResourceAttr(resourceName, "address_line_1", "123 Any Street"),
					resource.TestCheckResourceAttr(resourceName, "city", "Seattle"),
					resource.TestCheckResourceAttr(resourceName, "company_name", "Example Corp, Inc."),
					resource.TestCheckResourceAttr(resourceName, "country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "district_or_county", "King"),
					resource.TestCheckResourceAttr(resourceName, "full_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "+64211111111"),
					resource.TestCheckResourceAttr(resourceName, "postal_code", "98101"),
					resource.TestCheckResourceAttr(resourceName, "state_or_region", "WA"),
					resource.TestCheckResourceAttr(resourceName, "website_url", "https://www.example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPrimaryConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrimaryContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, ""),
					resource.TestCheckResourceAttr(resourceName, "address_line_1", "123 Any Street"),
					resource.TestCheckResourceAttr(resourceName, "city", "Seattle"),
					resource.TestCheckResourceAttr(resourceName, "company_name", "Example Corp, Inc."),
					resource.TestCheckResourceAttr(resourceName, "country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "district_or_county", "King"),
					resource.TestCheckResourceAttr(resourceName, "full_name", rName2),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "+64211111111"),
					resource.TestCheckResourceAttr(resourceName, "postal_code", "98101"),
					resource.TestCheckResourceAttr(resourceName, "state_or_region", "WA"),
					resource.TestCheckResourceAttr(resourceName, "website_url", "https://www.example.com"),
				),
			},
		},
	})
}

func testAccPrimaryContact_accountID(t *testing.T) { // nosemgrep:ci.account-in-func-name
	ctx := acctest.Context(t)
	resourceName := "aws_account_primary_contact.test"
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccPrimaryContactConfig_organization(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrimaryContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "address_line_1", "123 Any Street"),
					resource.TestCheckResourceAttr(resourceName, "city", "Seattle"),
					resource.TestCheckResourceAttr(resourceName, "country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "full_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "+64211111111"),
					resource.TestCheckResourceAttr(resourceName, "postal_code", "98101"),
				),
			},
			{
				// Regression test for https://github.com/hashicorp/terraform-provider-aws/issues/41154:
				// importing by member account ID must not produce a diff.
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testAccPrimaryContactConfig_organization(rName1),
			},
			{
				Config: testAccPrimaryContactConfig_organization(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrimaryContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "full_name", rName2),
				),
			},
		},
	})
}

func testAccCheckPrimaryContactExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Account Primary Contact ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).AccountClient(ctx)

		_, err := tfaccount.FindContactInformation(ctx, conn, rs.Primary.Attributes[names.AttrAccountID])

		return err
	}
}

func testAccPrimaryContactConfig_organization(name string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "test" {
  provider = "awsalternate"
}

resource "aws_account_primary_contact" "test" {
  account_id      = data.aws_caller_identity.test.account_id
  address_line_1  = "123 Any Street"
  city            = "Seattle"
  country_code    = "US"
  full_name       = %[1]q
  phone_number    = "+64211111111"
  postal_code     = "98101"
  state_or_region = "WA"
}
`, name))
}

func testAccPrimaryConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_account_primary_contact" "test" {
  address_line_1     = "123 Any Street"
  city               = "Seattle"
  company_name       = "Example Corp, Inc."
  country_code       = "US"
  district_or_county = "King"
  full_name          = %[1]q
  phone_number       = "+64211111111"
  postal_code        = "98101"
  state_or_region    = "WA"
  website_url        = "https://www.example.com"
}
`, name)
}
