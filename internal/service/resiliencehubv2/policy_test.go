// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfresiliencehubv2 "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehubv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubV2Policy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "availability_slo.0.target", "99.9"),
					resource.TestCheckResourceAttr(resourceName, "multi_az.0.disaster_recovery_approach", "ACTIVE_ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "multi_az.0.rpo_in_minutes", "5"),
					resource.TestCheckResourceAttr(resourceName, "multi_az.0.rto_in_minutes", "10"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
		},
	})
}

func TestAccResilienceHubV2Policy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfresiliencehubv2.ResourcePolicy, resourceName),
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

func TestAccResilienceHubV2Policy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_full(rName, "initial description", 99.9, 5, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "initial description"),
					resource.TestCheckResourceAttr(resourceName, "availability_slo.0.target", "99.9"),
					resource.TestCheckResourceAttr(resourceName, "multi_az.0.rpo_in_minutes", "5"),
					resource.TestCheckResourceAttr(resourceName, "multi_az.0.rto_in_minutes", "10"),
				),
			},
			{
				Config: testAccPolicyConfig_full(rName, "updated description", 99.99, 1, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated description"),
					resource.TestCheckResourceAttr(resourceName, "availability_slo.0.target", "99.99"),
					resource.TestCheckResourceAttr(resourceName, "multi_az.0.rpo_in_minutes", "1"),
					resource.TestCheckResourceAttr(resourceName, "multi_az.0.rto_in_minutes", "5"),
				),
			},
		},
	})
}

func testAccCheckPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehubv2_policy" {
				continue
			}

			_, err := tfresiliencehubv2.FindPolicyByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if err == nil {
				return fmt.Errorf("Resilience Hub V2 Policy %s still exists", rs.Primary.Attributes[names.AttrARN])
			}
		}

		return nil
	}
}

func testAccCheckPolicyExists(ctx context.Context, t *testing.T, n string, v *awstypes.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Policy not found: %s", n)
		}

		if rs.Primary.Attributes[names.AttrARN] == "" {
			return fmt.Errorf("No Resilience Hub V2 Policy ARN is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		output, err := tfresiliencehubv2.FindPolicyByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

	_, err := conn.ListPolicies(ctx, &resiliencehubv2.ListPoliciesInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehubv2_policy" "test" {
  name = %[1]q

  availability_slo {
    target = 99.9
  }

  multi_az {
    disaster_recovery_approach = "ACTIVE_ACTIVE"
    rpo_in_minutes             = 5
    rto_in_minutes             = 10
  }
}
`, rName)
}

func testAccPolicyConfig_full(rName, description string, availTarget float64, rpo, rto int) string {
	return fmt.Sprintf(`
resource "aws_resiliencehubv2_policy" "test" {
  name        = %[1]q
  description = %[2]q

  availability_slo {
    target = %[3]g
  }

  multi_az {
    disaster_recovery_approach = "ACTIVE_ACTIVE"
    rpo_in_minutes             = %[4]d
    rto_in_minutes             = %[5]d
  }
}
`, rName, description, availTarget, rpo, rto)
}
