// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2TenantSuppressionAttributes_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_sesv2_tenant_suppression_attributes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantSuppressionAttributesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppression_scope"), knownvalue.StringExact(string(types.SuppressionListScopeTenant))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetSizeExact(2)),
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

func TestAccSESV2TenantSuppressionAttributes_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_sesv2_tenant_suppression_attributes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantSuppressionAttributesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetSizeExact(2)),
				},
			},
			{
				Config: testAccTenantSuppressionAttributesConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppression_scope"), knownvalue.StringExact(string(types.SuppressionListScopeAccount))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetExact(
						[]knownvalue.Check{
							knownvalue.StringExact(string(types.SuppressionListReasonComplaint)),
						}),
					),
				},
			},
		},
	})
}

func TestAccSESV2TenantSuppressionAttributes_disappears_tenant(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_sesv2_tenant_suppression_attributes.test"
	tenantResourceName := "aws_sesv2_tenant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantSuppressionAttributesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantSuppressionAttributesExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsesv2.ResourceTenant, tenantResourceName),
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

func testAccCheckTenantSuppressionAttributesExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindTenantSuppressionAttributesByName(ctx, conn, rs.Primary.Attributes["tenant_name"])

		return err
	}
}

// testAccCheckTenantSuppressionAttributesDestroy verifies the parent tenant is
// gone after the test. Suppression attributes have no delete API of their own
// (the resource uses a NoOp delete); they cease to exist when the tenant that
// owns them is destroyed.
func testAccCheckTenantSuppressionAttributesDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_tenant_suppression_attributes" {
				continue
			}

			_, err := tfsesv2.FindTenantByName(ctx, conn, rs.Primary.Attributes["tenant_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SESv2 Tenant %s still exists", rs.Primary.Attributes["tenant_name"])
		}

		return nil
	}
}

func testAccTenantSuppressionAttributesConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_tenant" "test" {
  tenant_name = %[1]q
}

resource "aws_sesv2_tenant_suppression_attributes" "test" {
  tenant_name        = aws_sesv2_tenant.test.tenant_name
  suppression_scope  = "TENANT"
  suppressed_reasons = ["BOUNCE", "COMPLAINT"]
}
`, rName)
}

func testAccTenantSuppressionAttributesConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_tenant" "test" {
  tenant_name = %[1]q
}

resource "aws_sesv2_tenant_suppression_attributes" "test" {
  tenant_name        = aws_sesv2_tenant.test.tenant_name
  suppression_scope  = "ACCOUNT"
  suppressed_reasons = ["COMPLAINT"]
}
`, rName)
}
