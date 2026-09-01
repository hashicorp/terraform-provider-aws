// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreWorkloadIdentity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var w bedrockagentcorecontrol.GetWorkloadIdentityOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_workload_identity.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckWorkloadIdentities(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkloadIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkloadIdentityExists(ctx, t, resourceName, &w),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("workload_identity_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`workload-identity-directory/.+/workload-identity/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_resource_oauth2_return_urls"), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccBedrockAgentCoreWorkloadIdentity_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var w bedrockagentcorecontrol.GetWorkloadIdentityOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_workload_identity.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckWorkloadIdentities(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkloadIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkloadIdentityExists(ctx, t, resourceName, &w),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceWorkloadIdentity, resourceName),
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

func TestAccBedrockAgentCoreWorkloadIdentity_update(t *testing.T) {
	ctx := acctest.Context(t)
	var w bedrockagentcorecontrol.GetWorkloadIdentityOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_workload_identity.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckWorkloadIdentities(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkloadIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityConfig_urls(rName, `"https://example.com/callback"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkloadIdentityExists(ctx, t, resourceName, &w),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_resource_oauth2_return_urls"), knownvalue.SetSizeExact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccWorkloadIdentityConfig_urls(rName, `"https://app.example.com/auth","https://example.com/callback"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkloadIdentityExists(ctx, t, resourceName, &w),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_resource_oauth2_return_urls"), knownvalue.SetSizeExact(2)),
				},
			},
		},
	})
}

func testAccCheckWorkloadIdentityDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_workload_identity" {
				continue
			}

			_, err := tfbedrockagentcore.FindWorkloadIdentityByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Workload Identity %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckWorkloadIdentityExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetWorkloadIdentityOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindWorkloadIdentityByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckWorkloadIdentities(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListWorkloadIdentitiesInput{}

	_, err := conn.ListWorkloadIdentities(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccWorkloadIdentityConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_workload_identity" "test" {
  name = %[1]q
}
`, rName)
}

func testAccWorkloadIdentityConfig_urls(rName, urls string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_workload_identity" "test" {
  name = %[1]q

  allowed_resource_oauth2_return_urls = [%[2]s]
}
`, rName, urls)
}
