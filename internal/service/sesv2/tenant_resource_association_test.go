// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2TenantResourceAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assoc awstypes.TenantResource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, "|", "tenant_name", names.AttrResourceARN),
				ImportStateVerifyIdentifierAttribute: "tenant_name",
			},
		},
	})
}

func TestAccSESV2TenantResourceAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var assoc awstypes.TenantResource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsesv2.ResourceTenantResource, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
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

			_, err := tfsesv2.FindTenantResourceAssociationByID(ctx, conn, rs.Primary.Attributes["tenant_name"], rs.Primary.Attributes[names.AttrResourceARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SESv2 Tenant Resource Association %s still exists", rs.Primary.Attributes["tenant_name"])
		}

		return nil
	}
}

func testAccCheckTenantResourceAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.TenantResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		output, err := tfsesv2.FindTenantResourceAssociationByID(ctx, conn, rs.Primary.Attributes["tenant_name"], rs.Primary.Attributes[names.AttrResourceARN])

		if err != nil {
			return err
		}

		*v = *output

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
  tenant_name  = aws_sesv2_tenant.test.tenant_name
  resource_arn = aws_sesv2_configuration_set.test.arn
}
`, name)
}
