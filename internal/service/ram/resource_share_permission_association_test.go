// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
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
			acctest.PreCheckPartitionHasService(t, names.RAMEndpointID)
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcesharepermissionassociation ram.DescribeResourceSharePermissionAssociationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ram_resource_share_permission_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RAMEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSharePermissionAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSharePermissionAssociationConfig_basic(rName, testAccResourceSharePermissionAssociationVersionNewer),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceSharePermissionAssociationExists(ctx, t, resourceName, &resourcesharepermissionassociation),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceResourceSharePermissionAssociation = newResourceSharePermissionAssociationResource
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfram.ResourceResourceSharePermissionAssociation, resourceName),
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

			// TIP: ==== FINDERS ====
			// The find function should be exported. Since it won't be used outside of the package, it can be exported
			// in the `exports_test.go` file.
			_, err := tfram.FindResourceSharePermissionAssociationByID(ctx, conn, rs.Primary.ID)
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

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).RAMClient(ctx)

	input := &ram.ListResourceSharePermissionAssociationsInput{}

	_, err := conn.ListResourceSharePermissionAssociations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckResourceSharePermissionAssociationNotRecreated(before, after *ram.DescribeResourceSharePermissionAssociationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.ResourceSharePermissionAssociationId), aws.ToString(after.ResourceSharePermissionAssociationId); before != after {
			return create.Error(names.RAM, create.ErrActionCheckingNotRecreated, tfram.ResNameResourceSharePermissionAssociation, aws.ToString(before.ResourceSharePermissionAssociationId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccResourceSharePermissionAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_ram_resource_share_permission_association" "test" {
  resource_share_permission_association_name             = %[1]q
  engine_type             = "ActiveRAM"
  engine_version          = %[2]q
  host_instance_type      = "ram.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}
