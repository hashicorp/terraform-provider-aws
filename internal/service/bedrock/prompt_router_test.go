// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	promptRouterFallbackModelID = "anthropic.claude-3-5-sonnet-20240620-v1:0" // lintignore:AWSAT003,AWSAT005
	promptRouterModelID1        = "anthropic.claude-3-5-sonnet-20240620-v1:0" // lintignore:AWSAT003,AWSAT005
	promptRouterModelID2        = "anthropic.claude-3-haiku-20240307-v1:0"    // lintignore:AWSAT003,AWSAT005
)

func TestAccBedrockPromptRouter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var promptRouter bedrock.GetPromptRouterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_prompt_router.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRouterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptRouterExists(ctx, resourceName, &promptRouter),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("prompt_router_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("prompt_router_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("fallback_model"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("models"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("routing_criteria"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "prompt_router_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "prompt_router_arn",
			},
		},
	})
}

func TestAccBedrockPromptRouter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var promptRouter bedrock.GetPromptRouterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_prompt_router.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRouterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptRouterExists(ctx, resourceName, &promptRouter),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrock.ResourcePromptRouter, resourceName),
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

func TestAccBedrockPromptRouter_description(t *testing.T) {
	ctx := acctest.Context(t)
	var promptRouter bedrock.GetPromptRouterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_prompt_router.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRouterConfig_description(rName, "Test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPromptRouterExists(ctx, resourceName, &promptRouter),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Test description")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "prompt_router_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "prompt_router_arn",
			},
		},
	})
}

func TestAccBedrockPromptRouter_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var promptRouter bedrock.GetPromptRouterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_prompt_router.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRouterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPromptRouterExists(ctx, resourceName, &promptRouter),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "prompt_router_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "prompt_router_arn",
			},
		},
	})
}

func testAccCheckPromptRouterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_prompt_router" {
				continue
			}

			_, err := tfbedrock.FindPromptRouterByARN(ctx, conn, rs.Primary.Attributes["prompt_router_arn"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Prompt Router %s still exists", rs.Primary.Attributes["prompt_router_arn"])
		}

		return nil
	}
}

func testAccCheckPromptRouterExists(ctx context.Context, n string, v *bedrock.GetPromptRouterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		resp, err := tfbedrock.FindPromptRouterByARN(ctx, conn, rs.Primary.Attributes["prompt_router_arn"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPromptRouterConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_bedrock_prompt_router" "test" {
  prompt_router_name = %[1]q

  fallback_model {
    model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
  }

  models {
    model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[3]s"
  }

  models {
    model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[4]s"
  }

  routing_criteria {
    response_quality_difference = 25
  }
}
`, rName, promptRouterFallbackModelID, promptRouterModelID1, promptRouterModelID2)
}

func testAccPromptRouterConfig_description(rName, description string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_bedrock_prompt_router" "test" {
  prompt_router_name = %[1]q
  description        = %[2]q

  fallback_model {
    model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[3]s"
  }

  models {
    model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[4]s"
  }

  models {
    model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[5]s"
  }

  routing_criteria {
    response_quality_difference = 25
  }
}
`, rName, description, promptRouterFallbackModelID, promptRouterModelID1, promptRouterModelID2)
}

func testAccPromptRouterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_bedrock_prompt_router" "test" {
  prompt_router_name = %[1]q

  fallback_model {
    model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[4]s"
  }

  models {
    model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[5]s"
  }

  models {
    model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[6]s"
  }

  routing_criteria {
    response_quality_difference = 25
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1, promptRouterFallbackModelID, promptRouterModelID1, promptRouterModelID2)
}
