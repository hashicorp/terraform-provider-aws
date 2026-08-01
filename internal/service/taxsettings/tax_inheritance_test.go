// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package taxsettings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"

	tftaxsettings "github.com/hashicorp/terraform-provider-aws/internal/service/taxsettings"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.TaxSettingsServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Heritage status can only be updated once within 15 minutes",
	)
}

func TestAccTaxSettingsTaxInheritance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_taxsettings_tax_inheritance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TaxSettingsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccTaxInheritanceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTaxInheritanceExists(ctx, t, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTaxInheritanceImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "heritage_status",
			},
		},
	})
}

func testAccTaxInheritanceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrImportStateIdFunc(resourceName, "heritage_status")
}

func testAccCheckTaxInheritanceExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.TaxSettings, create.ErrActionCheckingExistence, tftaxsettings.ResNameTaxInheritance, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["heritage_status"] == "" {
			return create.Error(names.TaxSettings, create.ErrActionCheckingExistence, tftaxsettings.ResNameTaxInheritance, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).TaxSettingsClient(ctx)

		heritageStatus, err := tftaxsettings.FindTaxInheritanceHeritageStatus(ctx, conn)
		if err != nil {
			create.Error(names.TaxSettings, create.ErrActionCheckingExistence, tftaxsettings.ResNameTaxInheritance, name, err)
		}

		if rs.Primary.Attributes["heritage_status"] != string(*heritageStatus) {
			err = errors.New("heritage_status attribute value does match remote")
			return create.Error(names.TaxSettings, create.ErrActionCheckingExistence, tftaxsettings.ResNameTaxInheritance, rs.Primary.Attributes["heritage_status"], err)
		}

		return nil
	}
}

func testAccTaxInheritanceConfig_basic() string {
	return `
resource "aws_taxsettings_tax_inheritance" "test" {
  heritage_status = "OptOut"
}
`
}
