// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

func TestAccBedrockAgentCoreOAuth2CredentialProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, resourceName, &oauth2credentialprovider),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`token-vault/default/oauth2credentialprovider/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "client_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "config.0.github.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.github.0.client_id", "test-client-id"),
					resource.TestCheckResourceAttr(resourceName, "config.0.github.0.client_secret", "test-client-secret"),
					resource.TestCheckResourceAttrSet(resourceName, "config.0.github.0.oauth_discovery.0.authorization_server_metadata.0.issuer"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`token-vault/default/oauth2credentialprovider/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore: []string{
					"config.0.github.0.client_secret",
					"config.0.github.0.client_id"},
			},
		},
	})
}

func TestAccBedrockAgentCoreOAuth2CredentialProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, resourceName, &oauth2credentialprovider),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceOAuth2CredentialProvider, resourceName),
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

func TestAccBedrockAgentCoreOAuth2CredentialProvider_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var oauth2credentialprovider bedrockagentcorecontrol.GetOauth2CredentialProviderOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_oauth2_credential_provider.test"

	// Test data variables
	clientId1 := "auth0-write-only-client-id"
	clientSecret1 := "auth0-write-only-client-secret"
	version1 := 1
	discoveryURL1 := "https://dev-example.auth0.com/.well-known/openid-configuration"

	clientId2 := "updated-auth0-write-only-client-id"
	clientSecret2 := "updated-auth0-write-only-client-secret"
	version2 := 2
	discoveryURL2 := "https://company.okta.com/.well-known/openid-configuration"

	clientId3 := "keycloak-write-only-client-id"
	clientSecret3 := "keycloak-write-only-client-secret"
	version3 := 3
	issuer := "https://auth.company.com/realms/production"
	authEndpoint := "https://auth.company.com/realms/production/protocol/openid-connect/auth"
	tokenEndpoint := "https://auth.company.com/realms/production/protocol/openid-connect/token"
	responseTypes := []string{"code", "id_token"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckOAuth2CredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOAuth2CredentialProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOAuth2CredentialProviderConfig_customWithDiscoveryURL(rName, clientId1, clientSecret1, version1, discoveryURL1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, resourceName, &oauth2credentialprovider),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`token-vault/default/oauth2credentialprovider/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "client_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "vendor", "CustomOauth2"),
					// WriteOnly attributes should not appear in state
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_id_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_secret_wo"),
					// Regular attributes should be null when WriteOnly is used
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_id"),
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_secret"),
					// Version attribute should be present
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.0.client_credentials_wo_version", "1"),
					// OAuth discovery should be configured
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.0.oauth_discovery.0.discovery_url", discoveryURL1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`token-vault/default/oauth2credentialprovider/.+$`)),
				),
			},
			{
				Config: testAccOAuth2CredentialProviderConfig_customWithDiscoveryURL(rName, clientId2, clientSecret2, version2, discoveryURL2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, resourceName, &oauth2credentialprovider),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`token-vault/default/oauth2credentialprovider/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "client_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "vendor", "CustomOauth2"),
					resource.TestCheckResourceAttr(resourceName, "config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.#", "1"),
					// WriteOnly attributes should not appear in state
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_id_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_secret_wo"),
					// Regular attributes should be null when WriteOnly is used
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_id"),
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_secret"),
					// Version attribute should be updated to 2
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.0.client_credentials_wo_version", "2"),
					// OAuth discovery should still be configured with discovery_url
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.0.oauth_discovery.0.discovery_url", discoveryURL2),
				),
			},
			{
				Config: testAccOAuth2CredentialProviderConfig_customWithAuthServerMetadata(rName, clientId3, clientSecret3, version3, issuer, authEndpoint, tokenEndpoint, responseTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOAuth2CredentialProviderExists(ctx, resourceName, &oauth2credentialprovider),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`token-vault/default/oauth2credentialprovider/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "client_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "vendor", "CustomOauth2"),
					resource.TestCheckResourceAttr(resourceName, "config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.#", "1"),
					// WriteOnly attributes should not appear in state
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_id_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_secret_wo"),
					// Regular attributes should be null when WriteOnly is used
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_id"),
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.client_secret"),
					// Version attribute should be updated to 3
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.0.client_credentials_wo_version", "3"),
					// OAuth discovery should now use full authorization_server_metadata instead of discovery_url
					resource.TestCheckNoResourceAttr(resourceName, "config.0.custom.0.oauth_discovery.0.discovery_url"),
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.0.oauth_discovery.0.authorization_server_metadata.0.issuer", issuer),
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.0.oauth_discovery.0.authorization_server_metadata.0.authorization_endpoint", authEndpoint),
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.0.oauth_discovery.0.authorization_server_metadata.0.token_endpoint", tokenEndpoint),
					resource.TestCheckResourceAttr(resourceName, "config.0.custom.0.oauth_discovery.0.authorization_server_metadata.0.response_types.#", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				// WriteOnly attributes won't be in state after import
				ImportStateVerifyIgnore: []string{
					"config.0.custom.0.client_id_wo",
					"config.0.custom.0.client_secret_wo",
					"config.0.custom.0.client_credentials_wo_version",
				},
			},
		},
	})
}

func testAccCheckOAuth2CredentialProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_oauth2_credential_provider" {
				continue
			}

			_, err := tfbedrockagentcore.FindOAuth2CredentialProviderByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameOAuth2CredentialProvider, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameOAuth2CredentialProvider, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOAuth2CredentialProviderExists(ctx context.Context, name string, oauth2credentialprovider *bedrockagentcorecontrol.GetOauth2CredentialProviderOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameOAuth2CredentialProvider, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameOAuth2CredentialProvider, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindOAuth2CredentialProviderByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameOAuth2CredentialProvider, rs.Primary.ID, err)
		}

		*oauth2credentialprovider = *resp

		return nil
	}
}

func testAccPreCheckOAuth2CredentialProviders(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

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

  config {
    github {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"
    }
  }
}
`, rName)
}

func testAccOAuth2CredentialProviderConfig_customWithDiscoveryURL(rName, clientId, clientSecret string, version int, discoveryURL string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  config {
    custom {
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

func testAccOAuth2CredentialProviderConfig_customWithAuthServerMetadata(rName, clientId, clientSecret string, version int, issuer, authEndpoint, tokenEndpoint string, responseTypes []string) string {
	responseTypesStr := `"` + strings.Join(responseTypes, `", "`) + `"`
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = %[1]q

  config {
    custom {
      client_id_wo                  = %[2]q
      client_secret_wo              = %[3]q
      client_credentials_wo_version = %[4]d

      oauth_discovery {
        authorization_server_metadata {
          issuer                 = %[5]q
          authorization_endpoint = %[6]q
          token_endpoint         = %[7]q
          response_types         = [%[8]s]
        }
      }
    }
  }
}
`, rName, clientId, clientSecret, version, issuer, authEndpoint, tokenEndpoint, responseTypesStr)
}
