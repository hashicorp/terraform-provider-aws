// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreWorkloadIdentity_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var w1, w2 bedrockagentcorecontrol.GetWorkloadIdentityOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_workload_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkloadIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityConfig(rName, `"https://example.com/callback"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkloadIdentityExists(ctx, resourceName, &w1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "allowed_resource_oauth2_return_urls.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_resource_oauth2_return_urls.*", "https://example.com/callback"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`workload-identity-directory/.+/workload-identity/.+$`)),
				),
			},
			{
				Config: testAccWorkloadIdentityConfig(rName, `"https://app.example.com/auth","https://example.com/callback"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkloadIdentityExists(ctx, resourceName, &w2),
					testAccCheckWorkloadIdentityNotRecreated(&w1, &w2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "allowed_resource_oauth2_return_urls.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_resource_oauth2_return_urls.*", "https://example.com/callback"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_resource_oauth2_return_urls.*", "https://app.example.com/auth"),
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

func TestAccBedrockAgentCoreWorkloadIdentity_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var workloadIdentity bedrockagentcorecontrol.GetWorkloadIdentityOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_workload_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkloadIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityConfig(rName, `"https://example.com/callback"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkloadIdentityExists(ctx, resourceName, &workloadIdentity),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceWorkloadIdentity, resourceName),
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

func testAccCheckWorkloadIdentityDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_workload_identity" {
				continue
			}

			_, err := tfbedrockagentcore.FindWorkloadIdentityByName(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameWorkloadIdentity, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameWorkloadIdentity, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckWorkloadIdentityExists(ctx context.Context, name string, workloadIdentity *bedrockagentcorecontrol.GetWorkloadIdentityOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameWorkloadIdentity, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameWorkloadIdentity, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindWorkloadIdentityByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameWorkloadIdentity, rs.Primary.ID, err)
		}

		*workloadIdentity = *resp

		return nil
	}
}

func testAccCheckWorkloadIdentityNotRecreated(before, after *bedrockagentcorecontrol.GetWorkloadIdentityOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if beforeName, afterName := *before.Name, *after.Name; beforeName != afterName {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingNotRecreated, tfbedrockagentcore.ResNameWorkloadIdentity, beforeName, errors.New("recreated"))
		}
		return nil
	}
}

func testAccWorkloadIdentityConfig(rName, urls string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_workload_identity" "test" {
  name = %[1]q

  allowed_resource_oauth2_return_urls = [%[2]s]
}
`, rName, urls)
}
