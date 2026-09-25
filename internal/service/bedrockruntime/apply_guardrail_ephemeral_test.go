// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockruntime_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockRuntimeApplyGuardrailEphemeral_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataPath := tfjsonpath.New("data")
	textCoveragePath := dataPath.AtMapKey("guardrail_coverage").AtSliceIndex(0).AtMapKey("text_characters").AtSliceIndex(0)
	wordPolicyUsagePath := dataPath.AtMapKey("usage").AtSliceIndex(0).AtMapKey("word_policy_units")
	positiveUsage := knownvalue.Int64Func(func(v int64) error {
		if v <= 0 {
			return fmt.Errorf("expected positive word policy usage, got %d", v)
		}
		return nil
	})

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockRuntimeServiceID),
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_10_0)},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccApplyGuardrailEphemeralConfig(rName, "Hello, world.", "INPUT", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test", dataPath.AtMapKey(names.AttrAction), knownvalue.StringExact("NONE")),
					statecheck.ExpectKnownValue("echo.test", dataPath.AtMapKey("guardrail_identifier"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.test", dataPath.AtMapKey("guardrail_version"), knownvalue.StringExact("DRAFT")),
					statecheck.ExpectKnownValue("echo.test", dataPath.AtMapKey(names.AttrSource), knownvalue.StringExact("INPUT")),
					statecheck.ExpectKnownValue("echo.test", textCoveragePath, knownvalue.ObjectExact(map[string]knownvalue.Check{
						"guarded": knownvalue.Int64Exact(13),
						"total":   knownvalue.Int64Exact(13),
					})),
					statecheck.ExpectKnownValue("echo.test", wordPolicyUsagePath, positiveUsage),
					statecheck.ExpectKnownValue("echo.test", dataPath.AtMapKey("content"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"text": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"qualifiers": knownvalue.Null(),
									"text":       knownvalue.StringExact("Hello, world."),
								}),
							}),
						}),
					})),
				},
			},
			{
				Config: testAccApplyGuardrailEphemeralConfig(rName, "exampleblockedword", "OUTPUT", ""),
				// Echo resources are immutable; recreate to capture this step's ephemeral result.
				Taint: []string{"echo.test"},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test", dataPath.AtMapKey(names.AttrAction), knownvalue.StringExact("GUARDRAIL_INTERVENED")),
					statecheck.ExpectKnownValue("echo.test", dataPath.AtMapKey(names.AttrSource), knownvalue.StringExact("OUTPUT")),
					statecheck.ExpectKnownValue("echo.test", dataPath.AtMapKey("assessments"), knownvalue.ListPartial(map[int]knownvalue.Check{
						0: knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"word_policy": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"custom_words": knownvalue.ListPartial(map[int]knownvalue.Check{
										0: knownvalue.ObjectPartial(map[string]knownvalue.Check{
											"action":   knownvalue.StringExact("BLOCKED"),
											"detected": knownvalue.Bool(true),
											"match":    knownvalue.StringExact("exampleblockedword"),
										}),
									}),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue("echo.test", textCoveragePath, knownvalue.ObjectExact(map[string]knownvalue.Check{
						"guarded": knownvalue.Int64Exact(18),
						"total":   knownvalue.Int64Exact(18),
					})),
					statecheck.ExpectKnownValue("echo.test", wordPolicyUsagePath, positiveUsage),
					statecheck.ExpectKnownValue("echo.test", dataPath.AtMapKey("output"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"text": knownvalue.StringExact("Output blocked."),
						}),
					})),
				},
			},
		},
	})
}

func TestAccBedrockRuntimeApplyGuardrailEphemeral_postcondition(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockRuntimeServiceID),
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_10_0)},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccApplyGuardrailEphemeralConfig(rName, "exampleblockedword", "INPUT", `
  lifecycle {
    postcondition {
      condition     = self.action == "NONE"
      error_message = "Content failed guardrail validation."
    }
  }
`),
				ExpectError: regexache.MustCompile(`Content failed guardrail validation\.`),
			},
			{
				Config: testAccApplyGuardrailEphemeralConfig(rName, "Hello, world.", "INPUT", `
  lifecycle {
    postcondition {
      condition     = self.action == "NONE"
      error_message = "Content failed guardrail validation."
    }
  }
`),
			},
		},
	})
}

func testAccApplyGuardrailEphemeralConfig(rName, text, source, extra string) string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_bedrockruntime_apply_guardrail.test"),
		fmt.Sprintf(`
resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "Input blocked."
  blocked_outputs_messaging = "Output blocked."

  word_policy_config {
    words_config {
      text = "exampleblockedword"
    }
  }
}

ephemeral "aws_bedrockruntime_apply_guardrail" "test" {
  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_id
  guardrail_version    = "DRAFT"
  source               = %[3]q

  content {
    text {
      text = %[2]q
    }
  }
%[4]s
}
`, rName, text, source, extra))
}
