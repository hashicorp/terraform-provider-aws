// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserPoolClientTerms_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TermsType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client_terms.test"
	configDirectory := config.StaticDirectory("testdata/UserPoolClientTerms/basic")
	configVariables := config.Variables{acctest.CtRName: config.StringVariable(rName)}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientTermsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: configDirectory,
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientTermsExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClientID, "aws_cognito_user_pool_client.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "terms_id"),
					resource.TestCheckResourceAttr(resourceName, "terms_name", "terms-of-use"),
					resource.TestCheckResourceAttr(resourceName, "links.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "links.cognito:default", "https://example.com/terms"),
					resource.TestCheckResourceAttr(resourceName, "links.cognito:french", "https://example.com/fr/terms"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, names.AttrUserPoolID, "terms_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "terms_id",
			},
			{
				ConfigDirectory: configDirectory,
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"terms_name":    config.StringVariable("privacy-policy"),
					"links": config.MapVariable(map[string]config.Variable{
						"cognito:default": config.StringVariable("https://example.com/privacy"),
					}),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientTermsExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "terms_name", "privacy-policy"),
					resource.TestCheckResourceAttr(resourceName, "links.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "links.cognito:default", "https://example.com/privacy"),
				),
			},
			{
				ConfigDirectory: configDirectory,
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"terms_name":    config.StringVariable("privacy-policy"),
					"links":         config.MapVariable(map[string]config.Variable{}),
				},
				Check: resource.TestCheckResourceAttr(resourceName, "links.%", "0"),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClientTerms_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TermsType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client_terms.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientTermsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/UserPoolClientTerms/basic"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolClientTermsExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcognitoidp.ResourceUserPoolClientTerms, resourceName),
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

func testAccCheckUserPoolClientTermsExists(ctx context.Context, t *testing.T, name string, v *awstypes.TermsType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource %s not found in state", name)
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)
		out, err := tfcognitoidp.FindUserPoolClientTermsByID(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes["terms_id"])
		if err != nil {
			return err
		}

		*v = *out
		return nil
	}
}

func testAccCheckUserPoolClientTermsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_pool_client_terms" {
				continue
			}

			termsID := rs.Primary.Attributes["terms_id"]
			_, err := tfcognitoidp.FindUserPoolClientTermsByID(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], termsID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito user pool client terms %s still exist", termsID)
		}
		return nil
	}
}
