// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfaccount "github.com/hashicorp/terraform-provider-aws/internal/service/account"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPrimaryContact_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_primary_contact.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccPrimaryConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrimaryContactExists(ctx, resourceName),
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
					resource.TestCheckResourceAttr(resourceName, "website_url", "https://www.examplecorp.com"),
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
					testAccCheckPrimaryContactExists(ctx, resourceName),
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
					resource.TestCheckResourceAttr(resourceName, "website_url", "https://www.examplecorp.com"),
				),
			},
		},
	})
}

func testAccCheckPrimaryContactExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Account Primary Contact ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccountClient(ctx)

		_, err := tfaccount.FindContactInformation(ctx, conn, rs.Primary.Attributes[names.AttrAccountID])

		return err
	}
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
  website_url        = "https://www.examplecorp.com"
}
`, name)
}
