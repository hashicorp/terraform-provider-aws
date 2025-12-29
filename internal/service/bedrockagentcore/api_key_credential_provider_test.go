// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var p bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_api_key_credential_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAPIKeyCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyCredentialProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyCredentialProviderConfig_basic(rName, "secret-value-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, resourceName, &p),
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
					testAccCheckAPIKeyCredentialProviderExists(ctx, resourceName, &p),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_api_key_credential_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAPIKeyCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyCredentialProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyCredentialProviderConfig_writeOnly(rName, "write-only-api-key-123", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, resourceName, &p),
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
					testAccCheckAPIKeyCredentialProviderExists(ctx, resourceName, &p),
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

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var p bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_api_key_credential_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAPIKeyCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyCredentialProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyCredentialProviderConfig_basic(rName, "secret-value-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, resourceName, &p),
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

func testAccCheckAPIKeyCredentialProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

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

func testAccCheckAPIKeyCredentialProviderExists(ctx context.Context, n string, v *bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindAPIKeyCredentialProviderByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckAPIKeyCredentialProviders(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

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
