// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRAMResourceSharePermissionAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ram_resource_share_permission_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSharePermissionAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSharePermissionAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceSharePermissionAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "permission_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_share_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "permission_version"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRAMResourceSharePermissionAssociation_disappears(t *testing.T) {

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ram_resource_share_permission_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSharePermissionAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSharePermissionAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceSharePermissionAssociationExists(ctx, t, resourceName),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceResourceSharePermissionAssociation = newResourceSharePermissionAssociationResource
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfram.ResourceResourceSharePermissionAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckResourceSharePermissionAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ram_resource_share_permission_association" {
				continue
			}

			_, err := tfram.FindResourceSharePermissionByTwoPartKey(ctx, conn, rs.Primary.Attributes["resource_share_arn"], rs.Primary.Attributes["permission_version"])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.RAM, create.ErrActionCheckingDestroyed, tfram.ResNameResourceSharePermissionAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.RAM, create.ErrActionCheckingDestroyed, tfram.ResNameResourceSharePermissionAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResourceSharePermissionAssociationExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.RAM, create.ErrActionCheckingExistence, tfram.ResNameResourceSharePermissionAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.RAM, create.ErrActionCheckingExistence, tfram.ResNameResourceSharePermissionAssociation, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).RAMClient(ctx)

		_, err := tfram.FindResourceSharePermissionByTwoPartKey(ctx, conn, rs.Primary.Attributes["resource_share_arn"], rs.Primary.Attributes["permission_arn"])
		if err != nil {
			return create.Error(names.RAM, create.ErrActionCheckingExistence, tfram.ResNameResourceSharePermissionAssociation, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccResourceSharePermissionAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_permission" "test" {
  name          = %[1]q
  resource_type = "route53profiles:Profile"
  policy_template = jsonencode({
    Effect = "Allow"
    Action = [
      "route53profiles:GetProfile",
      "route53profiles:GetProfileResourceAssociation",
      "route53profiles:ListProfileResourceAssociations",
    ]
  })
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = false
}

resource "aws_ram_resource_share_permission_association" "test" {
  resource_share_arn = aws_ram_resource_share.test.arn
  permission_arn     = aws_ram_permission.test.arn
}
`, rName)
}
