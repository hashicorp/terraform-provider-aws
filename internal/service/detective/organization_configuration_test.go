// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package detective_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdetective "github.com/hashicorp/terraform-provider-aws/internal/service/detective"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	graphResourceName := "aws_detective_graph.test"
	resourceName := "aws_detective_organization_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_autoEnable(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
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
					testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "graph_arn", graphResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccCheckOrganizationConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DetectiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_detective_organization_configuration" {
				continue
			}

			_, err := tfdetective.FindOrganizationConfigurationByGraphARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Detective Organization Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOrganizationConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DetectiveClient(ctx)

		_, err := tfdetective.FindOrganizationConfigurationByGraphARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccOrganizationConfigurationConfig_autoEnable(autoEnable bool) string {
	return fmt.Sprintf(`
resource "aws_detective_organization_configuration" "test" {
  auto_enable = %[1]t
  graph_arn   = aws_detective_graph.test.graph_arn

  depends_on = [aws_detective_organization_admin_account.test]
}

resource "aws_detective_graph" "test" {}

resource "aws_detective_organization_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}

data "aws_caller_identity" "current" {}
`, autoEnable)
}
