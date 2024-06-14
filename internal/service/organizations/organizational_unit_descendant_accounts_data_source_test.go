// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationalUnitDescendantAccountsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	topOUDataSourceName := "data.aws_organizations_organizational_unit_descendant_accounts.current"
	newOU1DataSourceName := "data.aws_organizations_organizational_unit_descendant_accounts.test0"
	newOU2DataSourceName := "data.aws_organizations_organizational_unit_descendant_accounts.test1"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitDescendantAccountsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(topOUDataSourceName, "accounts.#", 0),
					resource.TestCheckResourceAttr(newOU1DataSourceName, "accounts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(newOU2DataSourceName, "accounts.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccOrganizationalUnitDescendantAccountsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test0" {
  name      = "%[1]s-0"
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "%[1]s-1"
  parent_id = aws_organizations_organizational_unit.test0.id
}

data "aws_organizations_organizational_unit_descendant_accounts" "current" {
  parent_id = data.aws_organizations_organization.current.roots[0].id

  depends_on = [aws_organizations_organizational_unit.test0, aws_organizations_organizational_unit.test1]
}

data "aws_organizations_organizational_unit_descendant_accounts" "test0" {
  parent_id = aws_organizations_organizational_unit.test0.id

  depends_on = [aws_organizations_organizational_unit.test1]
}

data "aws_organizations_organizational_unit_descendant_accounts" "test1" {
  parent_id = aws_organizations_organizational_unit.test1.id
}
`, rName)
}
