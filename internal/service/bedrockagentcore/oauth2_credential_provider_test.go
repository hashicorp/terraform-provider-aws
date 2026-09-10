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

func checkOAuth2CredentialProviderARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`token-vault/default/oauth2credentialprovider/`+name))
}

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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("client_secret_arn"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"secret_arn": tfknownvalue.RegionalARNRegexp("secretsmanager", regexache.MustCompile(`secret:.+`)),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_arn"), checkOAuth2CredentialProviderARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_vendor"), tfknownvalue.StringExact(awstypes.CredentialProviderVendorTypeGithubOauth2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"atlassian_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"custom_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"github_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Null(),
							names.AttrClientID:              knownvalue.StringExact("test-client-id"),
							"client_id_wo":                  knownvalue.Null(),
							names.AttrClientSecret:          knownvalue.StringExact("test-client-secret"),
							"client_secret_config":          knownvalue.ListSizeExact(0),
							"client_secret_source":          knownvalue.Null(),
							"client_secret_wo":              knownvalue.Null(),
							"oauth_discovery":               knownvalue.ListSizeExact(1),
						})}),
						"google_oauth2_provider_config":     knownvalue.ListSizeExact(0),
						"included_oauth2_provider_config":   knownvalue.ListSizeExact(0),
						"linkedin_oauth2_provider_config":   knownvalue.ListSizeExact(0),
						"microsoft_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"salesforce_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"slack_oauth2_provider_config":      knownvalue.ListSizeExact(0),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_secret_source",
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/atlassian_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_vendor"), tfknownvalue.StringExact(awstypes.CredentialProviderVendorTypeAtlassianOauth2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"atlassian_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Null(),
							names.AttrClientID:              knownvalue.StringExact("test-client-id"),
							"client_id_wo":                  knownvalue.Null(),
							names.AttrClientSecret:          knownvalue.StringExact("test-client-secret"),
							"client_secret_config":          knownvalue.ListSizeExact(0),
							"client_secret_source":          knownvalue.Null(),
							"client_secret_wo":              knownvalue.Null(),
							"oauth_discovery":               knownvalue.ListSizeExact(1),
						})}),
						"custom_oauth2_provider_config":     knownvalue.ListSizeExact(0),
						"github_oauth2_provider_config":     knownvalue.ListSizeExact(0),
						"google_oauth2_provider_config":     knownvalue.ListSizeExact(0),
						"included_oauth2_provider_config":   knownvalue.ListSizeExact(0),
						"linkedin_oauth2_provider_config":   knownvalue.ListSizeExact(0),
						"microsoft_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"salesforce_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"slack_oauth2_provider_config":      knownvalue.ListSizeExact(0),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/atlassian_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.atlassian_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.atlassian_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.atlassian_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.atlassian_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.atlassian_oauth2_provider_config.0.client_secret_source",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_google(t *testing.T) {
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/google_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_vendor"), tfknownvalue.StringExact(awstypes.CredentialProviderVendorTypeGoogleOauth2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"atlassian_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"custom_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"github_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"google_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Null(),
							names.AttrClientID:              knownvalue.StringExact("test-client-id"),
							"client_id_wo":                  knownvalue.Null(),
							names.AttrClientSecret:          knownvalue.StringExact("test-client-secret"),
							"client_secret_config":          knownvalue.ListSizeExact(0),
							"client_secret_source":          knownvalue.Null(),
							"client_secret_wo":              knownvalue.Null(),
							"oauth_discovery":               knownvalue.ListSizeExact(1),
						})}),
						"included_oauth2_provider_config":   knownvalue.ListSizeExact(0),
						"linkedin_oauth2_provider_config":   knownvalue.ListSizeExact(0),
						"microsoft_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"salesforce_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"slack_oauth2_provider_config":      knownvalue.ListSizeExact(0),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/google_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.google_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.google_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.google_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.google_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.google_oauth2_provider_config.0.client_secret_source",
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/included_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_vendor"), tfknownvalue.StringExact(awstypes.CredentialProviderVendorTypeTwitchOauth2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"atlassian_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"custom_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"github_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"google_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"included_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Null(),
							names.AttrClientID:              knownvalue.StringExact("test-client-id"),
							"client_id_wo":                  knownvalue.Null(),
							names.AttrClientSecret:          knownvalue.StringExact("test-client-secret"),
							"client_secret_config":          knownvalue.ListSizeExact(0),
							"client_secret_source":          knownvalue.Null(),
							"client_secret_wo":              knownvalue.Null(),
							"oauth_discovery":               knownvalue.ListSizeExact(1),
						})}),
						"linkedin_oauth2_provider_config":   knownvalue.ListSizeExact(0),
						"microsoft_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"salesforce_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"slack_oauth2_provider_config":      knownvalue.ListSizeExact(0),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/included_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.included_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.included_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.included_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.included_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.included_oauth2_provider_config.0.client_secret_source",
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/linkedin_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_vendor"), tfknownvalue.StringExact(awstypes.CredentialProviderVendorTypeLinkedinOauth2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"atlassian_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"custom_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"github_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"google_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"included_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"linkedin_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Null(),
							names.AttrClientID:              knownvalue.StringExact("test-client-id"),
							"client_id_wo":                  knownvalue.Null(),
							names.AttrClientSecret:          knownvalue.StringExact("test-client-secret"),
							"client_secret_config":          knownvalue.ListSizeExact(0),
							"client_secret_source":          knownvalue.Null(),
							"client_secret_wo":              knownvalue.Null(),
							"oauth_discovery":               knownvalue.ListSizeExact(1),
						})}),
						"microsoft_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"salesforce_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"slack_oauth2_provider_config":      knownvalue.ListSizeExact(0),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/linkedin_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.linkedin_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.linkedin_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.linkedin_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.linkedin_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.linkedin_oauth2_provider_config.0.client_secret_source",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_microsoft(t *testing.T) {
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/microsoft_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_vendor"), tfknownvalue.StringExact(awstypes.CredentialProviderVendorTypeMicrosoftOauth2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"atlassian_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"custom_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"github_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"google_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"included_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"linkedin_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"microsoft_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Null(),
							names.AttrClientID:              knownvalue.StringExact("test-client-id"),
							"client_id_wo":                  knownvalue.Null(),
							names.AttrClientSecret:          knownvalue.StringExact("test-client-secret"),
							"client_secret_config":          knownvalue.ListSizeExact(0),
							"client_secret_source":          knownvalue.Null(),
							"client_secret_wo":              knownvalue.Null(),
							"oauth_discovery":               knownvalue.ListSizeExact(1),
						})}),
						"salesforce_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"slack_oauth2_provider_config":      knownvalue.ListSizeExact(0),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/microsoft_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.microsoft_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.microsoft_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.microsoft_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.microsoft_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.microsoft_oauth2_provider_config.0.client_secret_source",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_salesforce(t *testing.T) {
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/salesforce_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_vendor"), tfknownvalue.StringExact(awstypes.CredentialProviderVendorTypeSalesforceOauth2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"atlassian_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"custom_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"github_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"google_oauth2_provider_config":    knownvalue.ListSizeExact(0),
						"included_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"linkedin_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"microsoft_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"salesforce_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Null(),
							names.AttrClientID:              knownvalue.StringExact("test-client-id"),
							"client_id_wo":                  knownvalue.Null(),
							names.AttrClientSecret:          knownvalue.StringExact("test-client-secret"),
							"client_secret_config":          knownvalue.ListSizeExact(0),
							"client_secret_source":          knownvalue.Null(),
							"client_secret_wo":              knownvalue.Null(),
							"oauth_discovery":               knownvalue.ListSizeExact(1),
						})}),
						"slack_oauth2_provider_config": knownvalue.ListSizeExact(0),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/salesforce_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.salesforce_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.salesforce_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.salesforce_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.salesforce_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.salesforce_oauth2_provider_config.0.client_secret_source",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_slack(t *testing.T) {
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/slack_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_vendor"), tfknownvalue.StringExact(awstypes.CredentialProviderVendorTypeSlackOauth2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"atlassian_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"custom_oauth2_provider_config":     knownvalue.ListSizeExact(0),
						"github_oauth2_provider_config":     knownvalue.ListSizeExact(0),
						"google_oauth2_provider_config":     knownvalue.ListSizeExact(0),
						"included_oauth2_provider_config":   knownvalue.ListSizeExact(0),
						"linkedin_oauth2_provider_config":   knownvalue.ListSizeExact(0),
						"microsoft_oauth2_provider_config":  knownvalue.ListSizeExact(0),
						"salesforce_oauth2_provider_config": knownvalue.ListSizeExact(0),
						"slack_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Null(),
							names.AttrClientID:              knownvalue.StringExact("test-client-id"),
							"client_id_wo":                  knownvalue.Null(),
							names.AttrClientSecret:          knownvalue.StringExact("test-client-secret"),
							"client_secret_config":          knownvalue.ListSizeExact(0),
							"client_secret_source":          knownvalue.Null(),
							"client_secret_wo":              knownvalue.Null(),
							"oauth_discovery":               knownvalue.ListSizeExact(1),
						})}),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/slack_oauth2_provider_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.slack_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.slack_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.slack_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.slack_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.slack_oauth2_provider_config.0.client_secret_source",
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/custom_oauth2_provider_config.oauth_discovery.discovery_url/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:              config.StringVariable(rName),
					"client_id":                  config.StringVariable("auth0-client-id"),
					"client_secret":              config.StringVariable("auth0-client-secret"),
					"client_credentials_version": config.IntegerVariable(1),
					"discovery_url":              config.StringVariable("https://dev-example.auth0.com/.well-known/openid-configuration"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"custom_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Int64Exact(1),
							"oauth_discovery": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"discovery_url": knownvalue.StringExact("https://dev-example.auth0.com/.well-known/openid-configuration"),
							})}),
						})}),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/custom_oauth2_provider_config.oauth_discovery.discovery_url/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:              config.StringVariable(rName),
					"client_id":                  config.StringVariable("auth0-client-id"),
					"client_secret":              config.StringVariable("auth0-client-secret"),
					"client_credentials_version": config.IntegerVariable(1),
					"discovery_url":              config.StringVariable("https://dev-example.auth0.com/.well-known/openid-configuration"),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_secret_source",
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/custom_oauth2_provider_config.oauth_discovery.discovery_url/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:              config.StringVariable(rName),
					"client_id":                  config.StringVariable("updated-client-id"),
					"client_secret":              config.StringVariable("updated-client-secret"),
					"client_credentials_version": config.IntegerVariable(2),
					"discovery_url":              config.StringVariable("https://company.okta.com/.well-known/openid-configuration"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"custom_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Int64Exact(2),
							"oauth_discovery": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"discovery_url": knownvalue.StringExact("https://company.okta.com/.well-known/openid-configuration"),
							})}),
						})}),
					})})),
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/custom_oauth2_provider_config.oauth_discovery.authorization_server_metadata/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"custom_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"oauth_discovery": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"authorization_server_metadata": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"authorization_endpoint": knownvalue.StringExact("https://auth.company.com/realms/production/protocol/openid-connect/auth"),
									names.AttrIssuer:         knownvalue.StringExact("https://auth.company.com/realms/production"),
									"response_types": knownvalue.SetExact([]knownvalue.Check{
										knownvalue.StringExact("code"),
										knownvalue.StringExact("id_token"),
									}),
									"token_endpoint": knownvalue.StringExact("https://auth.company.com/realms/production/protocol/openid-connect/token"),
								})}),
							})}),
						})}),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/custom_oauth2_provider_config.oauth_discovery.authorization_server_metadata/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_secret_source",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_clientSecretSourceExternal(t *testing.T) {
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
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/client_secret_source.EXTERNAL/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, t, resourceName, &oauth2credentialprovider),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("oauth2_provider_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"github_oauth2_provider_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"client_credentials_wo_version": knownvalue.Null(),
							names.AttrClientID:              knownvalue.StringExact("test-client-id"),
							"client_id_wo":                  knownvalue.Null(),
							names.AttrClientSecret:          knownvalue.Null(),
							"client_secret_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
								"json_key":  knownvalue.StringExact("clientSecret"),
								"secret_id": knownvalue.NotNull(),
							})}),
							"client_secret_source": tfknownvalue.StringExact(awstypes.SecretSourceTypeExternal),
							"client_secret_wo":     knownvalue.Null(),
							"oauth_discovery":      knownvalue.ListSizeExact(1),
						})}),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OAuth2CredentialProvider/client_secret_source.EXTERNAL/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_credentials_wo_version",
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_id",
					"oauth2_provider_config.0.github_oauth2_provider_config.0.client_secret",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_secret_config",
					"oauth2_provider_config.0.custom_oauth2_provider_config.0.client_secret_source",
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
