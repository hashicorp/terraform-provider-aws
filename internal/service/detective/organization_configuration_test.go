// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	graphResourceName := "aws_detective_graph.test"
	resourceName := "aws_detective_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// Detective Organization Configuration cannot be deleted separately.
		// Ensure parent resource is destroyed instead.
		CheckDestroy: testAccCheckGraphDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_autoEnable(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "graph_arn", graphResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig_autoEnable(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "graph_arn", graphResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccOrganizationConfigurationConfig_autoEnable(autoEnable bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["detective.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_detective_graph" "test" {}

resource "aws_detective_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_detective_organization_configuration" "test" {
  depends_on = [aws_detective_organization_admin_account.test]

  auto_enable = %[1]t
  graph_arn   = aws_detective_graph.test.id
}
`, autoEnable)
}
