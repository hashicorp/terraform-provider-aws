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

func TestAccBedrockAgentCoreOAuth2CredentialProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("client_secret_arn"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"secret_arn": tfknownvalue.RegionalARNRegexp("secretsmanager", regexache.MustCompile(`secret:.+`)),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`token-vault/default/oauth2credentialprovider/.+`))),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_id",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceOAuth2CredentialProvider, resourceName),
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

func TestAccBedrockAgentCoreOAuth2CredentialProvider_customDiscoveryURL(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_customWithDiscoveryURL(rName, "auth0-client-id", "auth0-client-secret", 1, "https://dev-example.auth0.com/.well-known/openid-configuration"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_credentials_wo_version",
				},
			},
			{
				Config: testAccOAuth2CredentialProviderConfig_customWithDiscoveryURL(rName, "updated-client-id", "updated-client-secret", 2, "https://company.okta.com/.well-known/openid-configuration"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
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

func TestAccBedrockAgentCoreOAuth2CredentialProvider_authorizationServerMetadata(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_customWithAuthServerMetadata(rName, "keycloak-client-id", "keycloak-client-secret", 1, "https://auth.company.com/realms/production", "https://auth.company.com/realms/production/protocol/openid-connect/auth", "https://auth.company.com/realms/production/protocol/openid-connect/token", "code", "id_token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_credentials_wo_version",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_full(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_credentials_wo_version",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_atlassian(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_atlassian(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.atlassian_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.atlassian_oauth2_provider_config.0.client_id",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_linkedin(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_linkedin(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.linkedin_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.linkedin_oauth2_provider_config.0.client_id",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_included(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_included(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.included_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.included_oauth2_provider_config.0.client_id",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_customTokenExchange(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_customTokenExchange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config").AtSliceIndex(0).AtMapKey("custom_oauth2_provider_config").AtSliceIndex(0).AtMapKey("client_authentication_method"), knownvalue.StringExact("CLIENT_SECRET_BASIC")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_credentials_wo_version",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
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
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_id",
				},
			},
			{
				Config: testAccOAuth2CredentialProviderConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
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
				Config: testAccOAuth2CredentialProviderConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
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

func testAccCheckOAuth2CredentialProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_oauth2_credential_provider" {
				continue
			}

			_, err := tfbedrockagentcore.FindOAuth2CredentialProviderByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core OAuth2 Credential Provider %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckOAuth2CredentialProviderExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetOauth2CredentialProviderOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindOAuth2CredentialProviderByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckOAuth2CredentialProviders(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListOauth2CredentialProvidersInput{}

	_, err := conn.ListOauth2CredentialProviders(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOAuth2CredentialProviderConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  credential_provider_vendor = "GithubOauth2"
  oauth2_provider_config {
    github_oauth2_provider_config {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"
    }
  }
}
`, rName)
}

func testAccOAuth2CredentialProviderConfig_customTokenExchange(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  credential_provider_vendor = "CustomOauth2"
  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id_wo                  = "token-exchange-client-id"
      client_secret_wo              = "token-exchange-client-secret"
      client_credentials_wo_version = 1
      client_authentication_method  = "CLIENT_SECRET_BASIC"

      oauth_discovery {
        discovery_url = "https://dev-example.auth0.com/.well-known/openid-configuration"
      }

      on_behalf_of_token_exchange_config {
        grant_type = "TOKEN_EXCHANGE"

        token_exchange_grant_type_config {
          actor_token_content = "M2M"
          actor_token_scopes  = ["read", "write"]
        }
      }
    }
  }
}
`, rName)
}

func testAccOAuth2CredentialProviderConfig_customWithDiscoveryURL(rName, clientId, clientSecret string, version int, discoveryURL string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  credential_provider_vendor = "CustomOauth2"
  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id_wo                  = %[2]q
      client_secret_wo              = %[3]q
      client_credentials_wo_version = %[4]d

      oauth_discovery {
        discovery_url = %[5]q
      }
    }
  }
}
`, rName, clientId, clientSecret, version, discoveryURL)
}

func testAccOAuth2CredentialProviderConfig_customWithAuthServerMetadata(rName, clientId, clientSecret string, version int, issuer, authEndpoint, tokenEndpoint, responseType1, responseType2 string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  credential_provider_vendor = "CustomOauth2"
  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id_wo                  = %[2]q
      client_secret_wo              = %[3]q
      client_credentials_wo_version = %[4]d

      oauth_discovery {
        authorization_server_metadata {
          issuer                 = %[5]q
          authorization_endpoint = %[6]q
          token_endpoint         = %[7]q
          response_types         = [%[8]q, %[9]q]
        }
      }
    }
  }
}
`, rName, clientId, clientSecret, version, issuer, authEndpoint, tokenEndpoint, responseType1, responseType2)
}

func testAccOAuth2CredentialProviderConfig_full(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  credential_provider_vendor = "CustomOauth2"
  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id_wo                  = "full-test-client-id"
      client_secret_wo              = "full-test-client-secret"
      client_credentials_wo_version = 1

      oauth_discovery {
        authorization_server_metadata {
          issuer                 = "https://auth.example.com/realms/production"
          authorization_endpoint = "https://auth.example.com/realms/production/protocol/openid-connect/auth"
          token_endpoint         = "https://auth.example.com/realms/production/protocol/openid-connect/token"
          response_types         = ["code", "id_token"]
        }
      }
    }
  }
}
`, rName)
}

func testAccOAuth2CredentialProviderConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  credential_provider_vendor = "GithubOauth2"
  oauth2_provider_config {
    github_oauth2_provider_config {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"
    }
  }
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccOAuth2CredentialProviderConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  credential_provider_vendor = "GithubOauth2"
  oauth2_provider_config {
    github_oauth2_provider_config {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccOAuth2CredentialProviderConfig_atlassian(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  credential_provider_vendor = "AtlassianOauth2"
  oauth2_provider_config {
    atlassian_oauth2_provider_config {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"
    }
  }
}
`, rName)
}

func testAccOAuth2CredentialProviderConfig_linkedin(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  credential_provider_vendor = "LinkedinOauth2"
  oauth2_provider_config {
    linkedin_oauth2_provider_config {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"
    }
  }
}
`, rName)
}

func testAccOAuth2CredentialProviderConfig_included(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  credential_provider_vendor = "XOauth2"
  oauth2_provider_config {
    included_oauth2_provider_config {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"
    }
  }
}
`, rName)
}

func testAccOAuth2CredentialProviderConfig_customExternalSecret(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name                    = %[1]q
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ clientSecret = "external-secret-value" })
}

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name                       = %[1]q
  credential_provider_vendor = "CustomOauth2"

  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_authentication_method = "AWS_IAM_ID_TOKEN_JWT"
      client_secret_source         = "EXTERNAL"

      oauth_discovery {
        discovery_url = "https://example.com/.well-known/openid-configuration"
      }

      client_secret_config {
        json_key  = "clientSecret"
        secret_id = aws_secretsmanager_secret_version.test.secret_id
      }
    }
  }
}
`, rName)
}

func testAccOAuth2CredentialProviderConfig_customExternalInlineSecret(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name                       = %[1]q
  credential_provider_vendor = "CustomOauth2"

  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_authentication_method = "CLIENT_SECRET_BASIC"
      client_id                    = "id"
      client_secret                = "should-not-be-here"
      client_secret_source         = "EXTERNAL"

      oauth_discovery {
        discovery_url = "https://example.com/.well-known/openid-configuration"
      }
    }
  }
}
`, rName)
}

func testAccOAuth2CredentialProviderConfig_customOverridesOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name                       = %[1]q
  credential_provider_vendor = "CustomOauth2"

  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id     = "id"
      client_secret = "secret"

      oauth_discovery {
        discovery_url = "https://example.com/.well-known/openid-configuration"
      }

      private_endpoint_overrides {
        domain = "example.com"
      }
    }
  }
}
`, rName)
}

func testAccOAuth2CredentialProviderConfig_multipleProviders(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name                       = %[1]q
  credential_provider_vendor = "GithubOauth2"

  oauth2_provider_config {
    github_oauth2_provider_config {
      client_id     = "id"
      client_secret = "secret"
    }
    google_oauth2_provider_config {
      client_id     = "id"
      client_secret = "secret"
    }
  }
}
`, rName)
}

// TestAccBedrockAgentCoreOAuth2CredentialProvider_customExternalSecret exercises
// the EXTERNAL client-secret path end-to-end (previously unusable). client_secret_source
// and client_secret_config are input-only fields (absent from GetOauth2CredentialProviderOutput);
// they cannot be recovered on the initial import, so they appear on ImportStateVerifyIgnore
// (a single subsequent `terraform apply` reconciles state and further plans are clean).
func TestAccBedrockAgentCoreOAuth2CredentialProvider_customExternalSecret(t *testing.T) {
	ctx := acctest.Context(t)
	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_customExternalSecret(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					// Input-only fields (absent from GetOauth2CredentialProviderOutput).
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_secret_source",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_secret_config",
				},
			},
		},
	})
}

// TestAccBedrockAgentCoreOAuth2CredentialProvider_validationRules exercises the
// plan-time validators added for the EXTERNAL secret contract, the
// private_endpoint_overrides sibling requirement, and the union ExactlyOneOf.
// These invalid configs previously validated offline and only failed at the API
// (some after creating orphaned resources).
func TestAccBedrockAgentCoreOAuth2CredentialProvider_validationRules(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				// EXTERNAL secret source rejects an inline client_secret.
				Config:      testAccOAuth2CredentialProviderConfig_customExternalInlineSecret(rName),
				ExpectError: regexache.MustCompile(`client_secret must not be set when client_secret_source is EXTERNAL`),
			},
			{
				// private_endpoint_overrides requires the sibling private_endpoint block.
				Config:      testAccOAuth2CredentialProviderConfig_customOverridesOnly(rName),
				ExpectError: regexache.MustCompile(`private_endpoint`),
			},
			{
				// Setting more than one provider block is rejected by ExactlyOneOf.
				Config:      testAccOAuth2CredentialProviderConfig_multipleProviders(rName),
				ExpectError: regexache.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}
