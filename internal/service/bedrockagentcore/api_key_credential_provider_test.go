// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
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

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var p1, p2 bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput
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
					testAccCheckAPIKeyCredentialProviderExists(ctx, resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "api_key", "secret-value-1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`token-vault/default/apikeycredentialprovider/.+`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "api_key_secret_arn", "secretsmanager", regexache.MustCompile(`secret:.+`)),
				),
			},
			{
				Config: testAccAPIKeyCredentialProviderConfig_basic(rName, "secret-value-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, resourceName, &p2),
					resource.TestCheckResourceAttr(resourceName, "api_key", "secret-value-2"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "api_key_secret_arn", "secretsmanager", regexache.MustCompile(`secret:.+`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{"api_key"},
			},
		},
	})
}

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_writeOnly(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var p1, p2 bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput
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
					testAccCheckAPIKeyCredentialProviderExists(ctx, resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "api_key_wo_version", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`token-vault/default/apikeycredentialprovider/.+`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "api_key_secret_arn", "secretsmanager", regexache.MustCompile(`secret:.+`)),
				),
			},
			{
				Config: testAccAPIKeyCredentialProviderConfig_writeOnly(rName, "updated-write-only-api-key-456", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyCredentialProviderExists(ctx, resourceName, &p2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "api_key_wo_version", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{"api_key_wo", "api_key_wo_version"},
			},
		},
	})
}

func TestAccBedrockAgentCoreAPIKeyCredentialProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceAPIKeyCredentialProvider, resourceName),
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

func testAccCheckAPIKeyCredentialProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_api_key_credential_provider" {
				continue
			}

			rName := rs.Primary.Attributes[names.AttrName]
			_, err := tfbedrockagentcore.FindAPIKeyCredentialProviderByName(ctx, conn, rName)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameAPIKeyCredentialProvider, rName, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameAPIKeyCredentialProvider, rName, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAPIKeyCredentialProviderExists(ctx context.Context, name string, apiKeyProvider *bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameAPIKeyCredentialProvider, name, errors.New("not found"))
		}

		rName := rs.Primary.Attributes[names.AttrName]

		if rName == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameAPIKeyCredentialProvider, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindAPIKeyCredentialProviderByName(ctx, conn, rName)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameAPIKeyCredentialProvider, rName, err)
		}

		*apiKeyProvider = *resp

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
