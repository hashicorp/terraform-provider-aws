// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAMPWorkspacesDataSource_basic(t *testing.T) { // nosemgrep:ci.caps0-in-func-name
	ctx := acctest.Context(t)
	rCount := sdkacctest.RandIntRange(1, 4)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_prometheus_workspaces.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:                acctest.ErrorCheck(t, names.AMPServiceID),
		PreventPostDestroyRefresh: true,
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDataSourceConfig_base(rName, rCount),
			},
			{
				Config: testAccWorkspacesDataSourceConfig_basic(rName, rCount),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "aliases.#", rCount),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "arns.#", rCount),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "workspace_ids.#", rCount),
				),
			},
		},
	})
}

func TestAccAMPWorkspacesDataSource_aliasPrefix(t *testing.T) { // nosemgrep:ci.caps0-in-func-name
	ctx := acctest.Context(t)
	rCount := sdkacctest.RandIntRange(1, 4)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_prometheus_workspaces.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:                acctest.ErrorCheck(t, names.AMPServiceID),
		PreventPostDestroyRefresh: true,
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDataSourceConfig_aliasPrefix(rName, rCount),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "aliases.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "workspace_ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccWorkspacesDataSourceConfig_base(rName string, rCount int) string { // nosemgrep:ci.caps0-in-func-name
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  count = %[2]d
  alias = "%[1]s-${count.index}"
}
`, rName, rCount)
}

func testAccWorkspacesDataSourceConfig_basic(rName string, rCount int) string { // nosemgrep:ci.caps0-in-func-name
	return acctest.ConfigCompose(testAccWorkspacesDataSourceConfig_base(rName, rCount), `
data "aws_prometheus_workspaces" "test" {}
`)
}

func testAccWorkspacesDataSourceConfig_aliasPrefix(rName string, rCount int) string { // nosemgrep:ci.caps0-in-func-name
	return acctest.ConfigCompose(testAccWorkspacesDataSourceConfig_base(rName, rCount), `
data "aws_prometheus_workspaces" "test" {
  alias_prefix = aws_prometheus_workspace.test[0].alias
}
`)
}
