// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockProvisionedModelThroughput_basic(t *testing.T) {
	acctest.Skip(t, "Bedrock Provisioned Model Throughput has a minimum 1 month commitment and costs > $10K/month")

	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_provisioned_model_throughput.test"
	var v bedrock.GetProvisionedModelThroughputOutput

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedModelThroughputDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedModelThroughputConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisionedModelThroughputExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "commitment_duration", "OneMonth"),
					resource.TestCheckResourceAttrSet(resourceName, "model_arn"),
					resource.TestCheckResourceAttr(resourceName, "model_units", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "provisioned_model_arn"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_model_name", rName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TODO TestAccBedrockProvisionedModelThroughput_disappears
// TODO TestAccBedrockProvisionedModelThroughput_tags

func TestAccBedrockProvisionedModelThroughput_Identity_ExistingResource(t *testing.T) {
	acctest.Skip(t, "Bedrock Provisioned Model Throughput has a minimum 1 month commitment and costs > $10K/month")

	ctx := acctest.Context(t)
	var v bedrock.GetProvisionedModelThroughputOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_provisioned_model_throughput.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.BedrockServiceID),
		CheckDestroy: testAccCheckProvisionedModelThroughputDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccProvisionedModelThroughputConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedModelThroughputExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.0.0",
					},
				},
				Config: testAccProvisionedModelThroughputConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedModelThroughputExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrARN)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccProvisionedModelThroughputConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedModelThroughputExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrARN)),
				},
			},
		},
	})
}

func testAccCheckProvisionedModelThroughputDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_custom_model" {
				continue
			}

			_, err := tfbedrock.FindProvisionedModelThroughputByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Provisioned Model Throughput %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProvisionedModelThroughputExists(ctx context.Context, t *testing.T, n string, v *bedrock.GetProvisionedModelThroughputOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		output, err := tfbedrock.FindProvisionedModelThroughputByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccProvisionedModelThroughputConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_provisioned_model_throughput" "test" {
  provisioned_model_name = %[1]q
  model_arn              = data.aws_bedrock_foundation_model.test.model_arn
  commitment_duration    = "OneMonth"
  model_units            = 1
}

data "aws_bedrock_foundation_model" "test" {
  model_id = "amazon.titan-text-express-v1:0:8k"
}
`, rName)
}
