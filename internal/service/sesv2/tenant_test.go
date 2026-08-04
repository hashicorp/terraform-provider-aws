// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2Tenant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_sesv2_tenant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTenantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sending_status"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("tenant_arn"), tfknownvalue.RegionalARNRegexp("ses", regexache.MustCompile(`tenant/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("tenant_id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "tenant_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tenant_name",
			},
		},
	})
}

func TestAccSESV2Tenant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_sesv2_tenant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTenantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsesv2.ResourceTenant, resourceName),
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

func testAccCheckTenantDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_tenant" {
				continue
			}

			_, err := tfsesv2.FindTenantByName(ctx, conn, rs.Primary.Attributes["tenant_name"])

			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SESv2 Tenant %s still exists", rs.Primary.Attributes["tenant_name"])
		}

		return nil
	}
}

func testAccCheckTenantExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindTenantByName(ctx, conn, rs.Primary.Attributes["tenant_name"])

		return err
	}
}

func testAccTenantConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_tenant" "test" {
  tenant_name = %[1]q
}
`, rName)
}
