// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGrafanaWorkspacesDataSource_basic(t *testing.T) { // nosemgrep:ci.caps0-in-func-name
	ctx := acctest.Context(t)
	rCount := sdkacctest.RandIntRange(1, 4)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_grafana_workspaces.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
		},
		ErrorCheck:                acctest.ErrorCheck(t, names.GrafanaServiceID),
		PreventPostDestroyRefresh: true,
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDataSourceConfig_base(rName, rCount),
			},
			{
				Config: testAccWorkspacesDataSourceConfig_basic(rName, rCount),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "names.#", rCount),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "workspace_ids.#", rCount),
				),
			},
		},
	})
}

func TestAccGrafanaWorkspacesDataSource_name(t *testing.T) { // nosemgrep:ci.caps0-in-func-name
	ctx := acctest.Context(t)
	rCount := sdkacctest.RandIntRange(1, 4)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_grafana_workspaces.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
		},
		ErrorCheck:                acctest.ErrorCheck(t, names.GrafanaServiceID),
		PreventPostDestroyRefresh: true,
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDataSourceConfig_name(rName, rCount),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "workspace_ids.#", "1"),
				),
			},
		},
	})
}

func testAccWorkspacesDataSourceConfig_base(rName string, rCount int) string { // nosemgrep:ci.caps0-in-func-name
	return fmt.Sprintf(`
resource "aws_iam_role" "assume" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "grafana.amazonaws.com"
        }
      },
    ]
  })
}
resource "aws_grafana_workspace" "test" {
  count = %[2]d
  account_access_type = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type = "SERVICE_MANAGED"
  role_arn = aws_iam_role.assume.arn
  name = "%[1]s-${count.index}"
}
`, rName, rCount)
}

func testAccWorkspacesDataSourceConfig_basic(rName string, rCount int) string { // nosemgrep:ci.caps0-in-func-name
	return acctest.ConfigCompose(testAccWorkspacesDataSourceConfig_base(rName, rCount), `
data "aws_grafana_workspaces" "test" {}
`)
}

func testAccWorkspacesDataSourceConfig_name(rName string, rCount int) string { // nosemgrep:ci.caps0-in-func-name
	return acctest.ConfigCompose(testAccWorkspacesDataSourceConfig_base(rName, rCount), `
data "aws_grafana_workspaces" "test" {
  name = aws_grafana_workspace.test[0].name
}
`)
}
