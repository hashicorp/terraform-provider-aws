// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreConsentPortal_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var portal bedrockagentcorecontrol.GetConsentPortalOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_consent_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckConsentPortals(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConsentPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConsentPortalConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConsentPortalExists(ctx, t, resourceName, &portal),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "consent_portal_id"),
					resource.TestCheckResourceAttrSet(resourceName, "portal_url"),
					resource.TestCheckResourceAttr(resourceName, "idp_config.0.scopes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sources.0.type", "agentcore-gateway"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "consent_portal_arn", "bedrock-agentcore", regexache.MustCompile(`consent-portal/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "consent_portal_id"),
				ImportStateVerifyIdentifierAttribute: "consent_portal_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreConsentPortal_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var portal bedrockagentcorecontrol.GetConsentPortalOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_consent_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckConsentPortals(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConsentPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConsentPortalConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConsentPortalExists(ctx, t, resourceName, &portal),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceConsentPortal, resourceName),
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

func testAccCheckConsentPortalDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_consent_portal" {
				continue
			}

			_, err := tfbedrockagentcore.FindConsentPortalByID(ctx, conn, rs.Primary.Attributes["consent_portal_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, "Consent Portal", rs.Primary.Attributes["consent_portal_id"], err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, "Consent Portal", rs.Primary.Attributes["consent_portal_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckConsentPortalExists(ctx context.Context, t *testing.T, name string, v *bedrockagentcorecontrol.GetConsentPortalOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, "Consent Portal", name, errors.New("not found"))
		}

		if rs.Primary.Attributes["consent_portal_id"] == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, "Consent Portal", name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		out, err := tfbedrockagentcore.FindConsentPortalByID(ctx, conn, rs.Primary.Attributes["consent_portal_id"])
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, "Consent Portal", rs.Primary.Attributes["consent_portal_id"], err)
		}
		*v = *out
		return nil
	}
}

func testAccPreCheckConsentPortals(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := &bedrockagentcorecontrol.ListConsentPortalsInput{}

	_, err := conn.ListConsentPortals(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccConsentPortalConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_consent_portal" "test" {

  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  idp_config {
    credential_provider_arn = aws_bedrockagentcore_oauth2_credential_provider.test.credential_provider_arn
    scopes                  = ["openid"]
  }

  sources {
    identifier = aws_bedrockagentcore_gateway.test.gateway_id
    type       = "agentcore-gateway"
  }

  depends_on = [aws_iam_role_policy_attachment.test]

}

resource "aws_cognito_user_pool" "test" {

  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_client" "test" {

  name                                 = %[1]q
  user_pool_id                         = aws_cognito_user_pool.test.id
  generate_secret                      = true
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_scopes                 = ["openid", "email", "profile"]
  callback_urls                        = ["https://example.com/callback"]
  supported_identity_providers         = ["COGNITO"]
}

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {

  name                       = %[1]q
  credential_provider_vendor = "CustomOauth2"
  depends_on                 = [aws_cognito_user_pool_domain.test]

  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id     = aws_cognito_user_pool_client.test.id
      client_secret = aws_cognito_user_pool_client.test.client_secret
      oauth_discovery {
        discovery_url = "https://${aws_cognito_user_pool.test.endpoint}/.well-known/openid-configuration"
      }
    }
  }
}

resource "aws_bedrockagentcore_gateway" "test" {

  name            = %[1]q
  role_arn        = aws_iam_role.test.arn
  authorizer_type = "CUSTOM_JWT"
  protocol_type   = "MCP"

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url   = "https://${aws_cognito_user_pool.test.endpoint}/.well-known/openid-configuration"
      allowed_clients = [aws_cognito_user_pool_client.test.id]
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/BedrockAgentCoreFullAccess"
}
`, rName)
}

func TestAccBedrockAgentCoreConsentPortal_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_consent_portal.test"
	var portal bedrockagentcorecontrol.GetConsentPortalOutput
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckConsentPortals(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConsentPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConsentPortalConfig_update(rName, "First description", `"openid", "email"`, `"example-audience"`, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConsentPortalExists(ctx, t, resourceName, &portal),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "First description"),
					resource.TestCheckResourceAttr(resourceName, "idp_config.0.scopes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "idp_config.0.audience", "example-audience"),
				),
			},
			{
				Config:           testAccConsentPortalConfig_update(rName, "Updated description", `"openid", "email", "profile"`, `"updated-audience"`, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)}},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConsentPortalExists(ctx, t, resourceName, &portal),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated description"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.alternate", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "idp_config.0.scopes.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "idp_config.0.audience", "updated-audience"),
				),
			},
			{
				Config:           testAccConsentPortalConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("idp_config").AtSliceIndex(0).AtMapKey("audience"), knownvalue.Null()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConsentPortalExists(ctx, t, resourceName, &portal),
					resource.TestCheckResourceAttr(resourceName, "idp_config.0.scopes.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "consent_portal_id"),
				ImportStateVerifyIdentifierAttribute: "consent_portal_id",
			},
		},
	})
}

func testAccConsentPortalConfig_update(rName, description, scopes, audience string, alternateRole bool) string {
	s := testAccConsentPortalConfig_basic(rName)
	s = strings.Replace(s, `execution_role_arn = aws_iam_role.test.arn`, fmt.Sprintf("execution_role_arn = aws_iam_role.test.arn\n description = %q", description), 1)
	s = strings.Replace(s, `scopes                  = ["openid"]`, fmt.Sprintf("scopes = [%s]\n audience = %s", scopes, audience), 1)
	if alternateRole {
		s = strings.Replace(s, "execution_role_arn = aws_iam_role.test.arn", "execution_role_arn = aws_iam_role.alternate.arn", 1)
		s = strings.Replace(s, "depends_on = [aws_iam_role_policy_attachment.test]", "depends_on = [aws_iam_role_policy_attachment.alternate]", 1)
	}
	return acctest.ConfigCompose(s, fmt.Sprintf(`
resource "aws_iam_role" "alternate" {
  name               = "%s-alternate"
  assume_role_policy = data.aws_iam_policy_document.test.json
}
resource "aws_iam_role_policy_attachment" "alternate" {
  role       = aws_iam_role.alternate.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/BedrockAgentCoreFullAccess"
}
`, rName))
}
