// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
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

func checkAPIKeyCredentialProviderARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`token-vault/default/apikeycredentialprovider/`+name))
}

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
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"api_key":       config.StringVariable("secret-value-1"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key"), knownvalue.StringExact("secret-value-1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_arn"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"secret_arn": knownvalue.NotNull(),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_config"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_source"), tfknownvalue.StringExact(awstypes.SecretSourceTypeManaged)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_wo"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_wo_version"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_arn"), checkAPIKeyCredentialProviderARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"api_key":       config.StringVariable("secret-value-1"),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{"api_key"},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"api_key":       config.StringVariable("secret-value-2"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key"), knownvalue.StringExact("secret-value-2")),
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
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key_wo/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"api_key_wo":         config.StringVariable("write-only-api-key-123"),
					"api_key_wo_version": config.IntegerVariable(1),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_arn"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"secret_arn": knownvalue.NotNull(),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_config"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_source"), tfknownvalue.StringExact(awstypes.SecretSourceTypeManaged)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_wo"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_wo_version"), knownvalue.Int64Exact(1)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key_wo/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"api_key_wo":         config.StringVariable("write-only-api-key-123"),
					"api_key_wo_version": config.IntegerVariable(1),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{"api_key_wo_version"},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key_wo/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"api_key_wo":         config.StringVariable("write-only-api-key-456"),
					"api_key_wo_version": config.IntegerVariable(2),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_arn"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"secret_arn": knownvalue.NotNull(),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_config"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_source"), tfknownvalue.StringExact(awstypes.SecretSourceTypeManaged)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_wo"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_wo_version"), knownvalue.Int64Exact(2)),
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
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
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

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_externalSecret(t *testing.T) {
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
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key_secret_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_arn"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"secret_arn": knownvalue.NotNull(),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"json_key":  knownvalue.StringExact("apiKey"),
						"secret_id": knownvalue.NotNull(),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_source"), tfknownvalue.StringExact(awstypes.SecretSourceTypeExternal)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_wo"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_wo_version"), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key_secret_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				// api_key_secret_config is input-only; the API does not echo it back on read.
				ImportStateVerifyIgnore: []string{"api_key_secret_config"},
			},
			{
				// Re-apply the same config to ensure the Optional+Computed
				// api_key_secret_source and input-only api_key_secret_config block
				// do not produce a perpetual diff.
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key_secret_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_switchSecretSource(t *testing.T) {
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
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"api_key":       config.StringVariable("secret-value-1"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_source"), tfknownvalue.StringExact(awstypes.SecretSourceTypeManaged)),
				},
			},
			{
				// The API cannot change the secret source between MANAGED and EXTERNAL in place.
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key_secret_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_source"), tfknownvalue.StringExact(awstypes.SecretSourceTypeExternal)),
				},
			},
			{
				// Switching back removes api_key_secret_source from configuration; the
				// effective source must be derived from api_key and still force replacement.
				ConfigDirectory: config.StaticDirectory("testdata/APIKeyCredentialProvider/api_key/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"api_key":       config.StringVariable("secret-value-2"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, t, resourceName, &p),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_key_secret_source"), tfknownvalue.StringExact(awstypes.SecretSourceTypeManaged)),
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
