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

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var p bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_api_key_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAPIKeyCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyCredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyCredentialProviderConfig_basic(rName, "secret-value-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`token-vault/default/apikeycredentialprovider/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_arn"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"secret_arn": tfknownvalue.RegionalARNRegexp("secretsmanager", regexache.MustCompile(`secret:.+`)),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{"api_key"},
			},
			{
				Config: testAccAPIKeyCredentialProviderConfig_basic(rName, "secret-value-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_writeOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var p bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_api_key_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAPIKeyCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyCredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyCredentialProviderConfig_writeOnly(rName, "write-only-api-key-123", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`token-vault/default/apikeycredentialprovider/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_arn"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"secret_arn": tfknownvalue.RegionalARNRegexp("secretsmanager", regexache.MustCompile(`secret:.+`)),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{"api_key_wo_version"},
			},
			{
				Config: testAccAPIKeyCredentialProviderConfig_writeOnly(rName, "updated-write-only-api-key-456", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var p bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_api_key_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAPIKeyCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyCredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyCredentialProviderConfig_tags1(rName, "secret-value-1", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{"api_key"},
			},
			{
				Config: testAccAPIKeyCredentialProviderConfig_tags2(rName, "secret-value-2", acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccAPIKeyCredentialProviderConfig_tags1(rName, "secret-value-1", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var p bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_api_key_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAPIKeyCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyCredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyCredentialProviderConfig_basic(rName, "secret-value-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceAPIKeyCredentialProvider, resourceName),
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

func testAccCheckAPIKeyCredentialProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_api_key_credential_provider" {
				continue
			}

			_, err := tfbedrockagentcore.FindAPIKeyCredentialProviderByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core API Key Credential Provider %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckAPIKeyCredentialProviderExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindAPIKeyCredentialProviderByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckAPIKeyCredentialProviders(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListApiKeyCredentialProvidersInput{}

	_, err := conn.ListApiKeyCredentialProviders(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAPIKeyCredentialProviderConfig_basic(rName, apiKey string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_api_key_credential_provider" "test" {
  name    = %[1]q
  api_key = %[2]q
}
`, rName, apiKey)
}

func testAccAPIKeyCredentialProviderConfig_writeOnly(rName, apiKeyWo string, version int) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_api_key_credential_provider" "test" {
  name               = %[1]q
  api_key_wo         = %[2]q
  api_key_wo_version = %[3]d
}
`, rName, apiKeyWo, version)
}

func testAccAPIKeyCredentialProviderConfig_tags1(rName, apiKey, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_api_key_credential_provider" "test" {
  name    = %[1]q
  api_key = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, apiKey, tag1Key, tag1Value)
}

func testAccAPIKeyCredentialProviderConfig_tags2(rName, apiKey, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_api_key_credential_provider" "test" {
  name    = %[1]q
  api_key = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, apiKey, tag1Key, tag1Value, tag2Key, tag2Value)
}
