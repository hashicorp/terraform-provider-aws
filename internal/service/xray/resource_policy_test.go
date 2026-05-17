// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package xray_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfxray "github.com/hashicorp/terraform-provider-aws/internal/service/xray"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccXRayResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcepolicy types.ResourcePolicy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_xray_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_check", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "policy_revision_id", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrLastUpdatedTime), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_revision_id"), knownvalue.StringExact("1")),
				},
			},
		},
	})
}

func TestAccXRayResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcepolicy types.ResourcePolicy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_xray_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfxray.ResourceResourcePolicy, resourceName),
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

func TestAccXRayResourcePolicy_policyDocument(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcepolicy types.ResourcePolicy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_xray_resource_policy.test"
	policyDocument1 := `{"Version":"2012-10-17","Statement":[{"Sid":"AllowXRayAccess","Effect":"Allow","Principal":{"AWS":"*"},"Action":["xray:*","xray:PutResourcePolicy"],"Resource":"*"}]}`
	policyDocument2 := `{"Version":"2012-10-17","Statement":[{"Sid":"AllowXRayAccessUpdated","Effect":"Allow","Principal":{"AWS":"*"},"Action":["xray:PutTraceSegments","xray:PutTelemetryRecords"],"Resource":"*"}]}`
	policyDocument3 := `{"Version":"2012-10-17","Statement":[{"Sid":"AllowXRayAccessFinal","Effect":"Allow","Principal":{"Service":"sns.amazonaws.com"},"Action":"xray:PutTraceSegments","Resource":"*"}]}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_policyDocument(rName, policyDocument1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					acctest.CheckResourceAttrJSONNoDiff(resourceName, "policy_document", policyDocument1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_revision_id"), knownvalue.StringExact("1")),
				},
			},
			{
				Config: testAccResourcePolicyConfig_policyDocument(rName, policyDocument2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					acctest.CheckResourceAttrJSONNoDiff(resourceName, "policy_document", policyDocument2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_revision_id"), knownvalue.StringExact("2")),
				},
			},
			{
				Config: testAccResourcePolicyConfig_policyDocument(rName, policyDocument3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					acctest.CheckResourceAttrJSONNoDiff(resourceName, "policy_document", policyDocument3),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_revision_id"), knownvalue.StringExact("3")),
				},
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).XRayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_xray_resource_policy" {
				continue
			}

			_, err := tfxray.FindResourcePolicyByName(ctx, conn, rs.Primary.Attributes["policy_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("XRay Resource Policy %s still exists", rs.Primary.Attributes["policy_name"])
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, n string, v *types.ResourcePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).XRayClient(ctx)

		output, err := tfxray.FindResourcePolicyByName(ctx, conn, rs.Primary.Attributes["policy_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccResourcePolicyConfig_policyDocument(rName, policyDocument string) string {
	return fmt.Sprintf(`
resource "aws_xray_resource_policy" "test" {
  policy_name                 = %[1]q
  policy_document             = %[2]q
  bypass_policy_lockout_check = true
}
`, rName, policyDocument)
}
