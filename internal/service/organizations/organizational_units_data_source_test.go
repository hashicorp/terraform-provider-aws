// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccOrganizationalUnitsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	topOUDataSourceName := "data.aws_organizations_organizational_units.current"
	newOUDataSourceName := "data.aws_organizations_organizational_units.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(topOUDataSourceName, "children.#", 0),
					resource.TestCheckResourceAttr(newOUDataSourceName, "children.#", "0"),
				),
			},
		},
	})
}

const testAccOrganizationalUnitsDataSourceConfig_basic = `
data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = "test"
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

data "aws_organizations_organizational_units" "current" {
  parent_id = aws_organizations_organizational_unit.test.parent_id
}

data "aws_organizations_organizational_units" "test" {
  parent_id = aws_organizations_organizational_unit.test.id
}
`
