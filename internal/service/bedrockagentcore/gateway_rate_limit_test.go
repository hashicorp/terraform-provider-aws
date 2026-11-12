// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreGatewayRateLimit_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_rate_limit.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRateLimitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRateLimitExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// No rate_limit_id in the config, so the service generates one.
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rate_limit_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dimension_keys"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("targetName"),
					})),
					// description was never set, so it must stay null rather than
					// becoming "": the API omits an absent description entirely.
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportState:                          true,
				ImportStateVerify:                    true,
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccGatewayRateLimitImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "rate_limit_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayRateLimit_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_rate_limit.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRateLimitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRateLimitExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceGatewayRateLimit, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// _update walks the description through all three of its states in one gateway.
// The API distinguishes them: an omitted description clears, whereas "" sets an
// empty string.
func TestAccBedrockAgentCoreGatewayRateLimit_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_rate_limit.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRateLimitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:       config.StringVariable(rName),
					names.AttrDescription: config.StringVariable("initial description"),
					"rate":                config.IntegerVariable(100),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRateLimitExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial description")),
				},
			},
			{
				// Entries changed and description emptied. An in-place update:
				// neither dimension_keys nor rate_limit_id moved.
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:       config.StringVariable(rName),
					names.AttrDescription: config.StringVariable(""),
					"rate":                config.IntegerVariable(250),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("")),
				},
			},
			{
				// Description removed from config entirely. The provider omits it
				// from the request, which clears it server-side and returns null.
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
				},
			},
		},
	})
}

// dimension_keys is immutable server-side, so changing it must replace.
func TestAccBedrockAgentCoreGatewayRateLimit_forceNew(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_rate_limit.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRateLimitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/dimension_keys/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"dimension_key": config.StringVariable("targetName"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRateLimitExists(ctx, t, resourceName),
				),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/dimension_keys/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"dimension_key": config.StringVariable("toolName"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dimension_keys"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("toolName"),
					})),
				},
			},
		},
	})
}

// _wildcards covers the behaviours the live API spike verified: a trailing "*",
// several entries under one limit, rate = 0 as a hard block, and a fractional
// rate. All in one gateway.
func TestAccBedrockAgentCoreGatewayRateLimit_wildcards(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_rate_limit.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRateLimitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/wildcards/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRateLimitExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("entries"), knownvalue.SetSizeExact(4)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rate_limit_id"), knownvalue.StringExact("wildcards")),
				},
			},
			{
				// The real drift check: everything above must round-trip with no
				// diff, including the fractional rate and rate = 0.
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/wildcards/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				PlanOnly: true,
			},
		},
	})
}

// Several rate limits on one gateway is the shape issue #49344's example never
// shows, and the case the shared per-gateway mutex exists to serialise.
func TestAccBedrockAgentCoreGatewayRateLimit_multipleLimits(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_rate_limit.test"
	secondResourceName := "aws_bedrockagentcore_gateway_rate_limit.second"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRateLimitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/multiple/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRateLimitExists(ctx, t, resourceName),
					testAccCheckGatewayRateLimitExists(ctx, t, secondResourceName),
				),
			},
		},
	})
}

// Two rate limits sharing a dimension_keys tuple on one gateway is rejected.
// Note the API returns ValidationException, NOT the ConflictException the AWS
// developer guide's error table claims. Verified against the live API on
// 2026-08-12; this test exists to catch anyone "correcting" it back.
func TestAccBedrockAgentCoreGatewayRateLimit_duplicateDimensionKeys(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRateLimitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimit/duplicate/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				// Kept short: Terraform wraps long diagnostics, so a pattern
				// spanning "exists for this gateway" would straddle a newline.
				ExpectError: regexache.MustCompile(`ValidationException: A limit with dimensionKeys`),
			},
		},
	})
}

func testAccCheckGatewayRateLimitDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_gateway_rate_limit" {
				continue
			}

			_, err := tfbedrockagentcore.FindGatewayRateLimitByTwoPartKey(ctx, conn, rs.Primary.Attributes["gateway_identifier"], rs.Primary.Attributes["rate_limit_id"])
			// A deleted gateway cascade-deletes its rate limits, and the API then
			// reports the gateway as missing. Both arrive as NotFound.
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Gateway Rate Limit %s still exists", rs.Primary.Attributes["rate_limit_id"])
		}

		return nil
	}
}

func testAccCheckGatewayRateLimitExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		_, err := tfbedrockagentcore.FindGatewayRateLimitByTwoPartKey(ctx, conn, rs.Primary.Attributes["gateway_identifier"], rs.Primary.Attributes["rate_limit_id"])

		return err
	}
}

func testAccGatewayRateLimitImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "gateway_identifier", "rate_limit_id")
}
