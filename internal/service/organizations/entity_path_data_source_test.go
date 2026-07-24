// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	ouEntityPathRegexp      = regexache.MustCompile(`o-[a-z0-9]{10,32}/r-[0-9a-z]{4,32}(/ou-[0-9a-z]{4,32}-[a-z0-9]{8,32})+/`)
	accountEntityPathRegexp = regexache.MustCompile(`o-[a-z0-9]{10,32}/r-[0-9a-z]{4,32}(/ou-[0-9a-z]{4,32}-[a-z0-9]{8,32})*/\d{12}/`)
)

func TestAccOrganizationsEntityPathDataSource_ou(t *testing.T) {
	ctx := acctest.Context(t)
	datasource1Name := "data.aws_organizations_entity_path.test1"
	datasource2Name := "data.aws_organizations_entity_path.test2"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEntityPathDataSourceConfig_ou(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(datasource1Name, tfjsonpath.New("entity_path"), knownvalue.StringRegexp(ouEntityPathRegexp)),
					statecheck.ExpectKnownValue(datasource2Name, tfjsonpath.New("entity_path"), knownvalue.StringRegexp(ouEntityPathRegexp)),
				},
			},
		},
	})
}

func TestAccOrganizationsEntityPathDataSource_account(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_organizations_entity_path.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEntityPathDataSourceConfig_account,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(datasourceName, tfjsonpath.New("entity_path"), knownvalue.StringRegexp(accountEntityPathRegexp)),
				},
			},
		},
	})
}

func testAccEntityPathDataSourceConfig_ou(rName string) string {
	return acctest.ConfigCompose(testAccOrganizationalUnitDataSourceConfig_basic(rName), `
data "aws_organizations_entity_path" "test1" {
  entity_id = aws_organizations_organizational_unit.parent.id
}

data "aws_organizations_entity_path" "test2" {
  entity_id = aws_organizations_organizational_unit.child.id
}
`)
}

const testAccEntityPathDataSourceConfig_account = `
data "aws_caller_identity" "current" {}

data "aws_organizations_entity_path" "test" {
  entity_id = data.aws_caller_identity.current.account_id
}
`
