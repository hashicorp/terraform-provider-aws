// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2TenantResourceAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var assoc awstypes.TenantResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_tenant_resource_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTenantResourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTenantResourceAssociationExists(ctx, t, resourceName, &assoc),
					resource.TestCheckResourceAttr(resourceName, "tenant_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceARN),
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

func TestAccSESV2TenantResourceAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var assoc awstypes.TenantResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_tenant_resource_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTenantResourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTenantResourceAssociationExists(ctx, t, resourceName, &assoc),
					acctest.CheckFrameworkResourceDisappears(
						ctx,
						t,
						tfsesv2.ResourceTenantResource,
						resourceName,
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTenantResourceAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_tenant_resource_association" {
				continue
			}

			_, err := tfsesv2.FindTenantResourceAssociationByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(
					names.SESV2,
					create.ErrActionCheckingDestroyed,
					tfsesv2.ResNameTenantResourceAssociation,
					rs.Primary.ID,
					err,
				)
			}

			return create.Error(
				names.SESV2,
				create.ErrActionCheckingDestroyed,
				tfsesv2.ResNameTenantResourceAssociation,
				rs.Primary.ID,
				errors.New("still exists"),
			)
		}

		return nil
	}
}

func testAccCheckTenantResourceAssociationExists(
	ctx context.Context,
	t *testing.T,
	name string,
	out *awstypes.TenantResource,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return errors.New("resource not found in state")
		}

		if rs.Primary.ID == "" {
			return errors.New("resource ID not set")
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		resp, err := tfsesv2.FindTenantResourceAssociationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*out = *resp
		return nil
	}
}

func testAccTenantResourceAssociationConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_tenant" "test" {
  tenant_name = %[1]q
}

resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_tenant_resource_association" "test" {
  tenant_name   = aws_sesv2_tenant.test.tenant_name
  resource_arn  = aws_sesv2_configuration_set.test.arn
}
`, name)
}
