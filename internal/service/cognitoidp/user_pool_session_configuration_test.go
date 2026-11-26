// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserPoolSessionConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v cognitoidentityprovider.GetUserPoolSessionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_session_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolSessionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolSessionConfigurationConfig_basic(rName, "ENABLED", 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolSessionConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "user_session_tracking", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "token_session.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "token_session.0.max_per_user", "20"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrUserPoolID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserPoolID,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolSessionConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v cognitoidentityprovider.GetUserPoolSessionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_session_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolSessionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolSessionConfigurationConfig_basic(rName, "ENABLED", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolSessionConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_session_tracking", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "token_session.0.max_per_user", "1"),
				),
			},
			{
				Config: testAccUserPoolSessionConfigurationConfig_basic(rName, "ENABLED", 100),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolSessionConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_session_tracking", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "token_session.0.max_per_user", "100"),
				),
			},
			{
				Config: testAccUserPoolSessionConfigurationConfig_basic(rName, "DISABLED", 100),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolSessionConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_session_tracking", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "token_session.0.max_per_user", "100"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrUserPoolID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserPoolID,
			},
		},
	})
}

// The API has no delete operation, so destroying the resource restores the defaults.
// Terraform must then plan an in-place update rather than a create.
func TestAccCognitoIDPUserPoolSessionConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v cognitoidentityprovider.GetUserPoolSessionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_session_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolSessionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolSessionConfigurationConfig_basic(rName, "ENABLED", 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolSessionConfigurationExists(ctx, t, resourceName, &v),
					// nosemgrep:ci.semgrep.acctest.disappears-expect-resource-action -- Session configuration is a sub-resource of the user pool; "disappearing" it resets the pool to the default configuration, so the plan is an update, not a create.
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcognitoidp.ResourceUserPoolSessionConfiguration, resourceName),
					testAccCheckUserPoolSessionConfigurationIsDefault(ctx, t, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolSessionConfiguration_disappears_userPool(t *testing.T) {
	ctx := acctest.Context(t)
	var v cognitoidentityprovider.GetUserPoolSessionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_session_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolSessionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolSessionConfigurationConfig_basic(rName, "ENABLED", 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolSessionConfigurationExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcognitoidp.ResourceUserPool(), "aws_cognito_user_pool.test"),
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

func TestAccCognitoIDPUserPoolSessionConfiguration_liteTier(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolSessionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserPoolSessionConfigurationConfig_tier(rName, "LITE"),
				ExpectError: regexache.MustCompile(`FeatureUnavailableInTierException`),
			},
		},
	})
}

func testAccCheckUserPoolSessionConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_pool_session_configuration" {
				continue
			}

			userPoolID := rs.Primary.Attributes[names.AttrUserPoolID]
			out, err := tfcognitoidp.FindUserPoolSessionConfigurationByUserPoolID(ctx, conn, userPoolID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.CognitoIDP, create.ErrActionCheckingDestroyed, "User Pool Session Configuration", userPoolID, err)
			}

			if !isDefaultUserPoolSessionConfiguration(out) {
				return create.Error(names.CognitoIDP, create.ErrActionCheckingDestroyed, "User Pool Session Configuration", userPoolID, errors.New("not reset to defaults"))
			}
		}

		return nil
	}
}

func testAccCheckUserPoolSessionConfigurationExists(ctx context.Context, t *testing.T, n string, v *cognitoidentityprovider.GetUserPoolSessionConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindUserPoolSessionConfigurationByUserPoolID(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckUserPoolSessionConfigurationIsDefault(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindUserPoolSessionConfigurationByUserPoolID(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID])
		if err != nil {
			return err
		}

		if !isDefaultUserPoolSessionConfiguration(output) {
			return fmt.Errorf("User Pool Session Configuration (%s) not reset to defaults: tracking %s, max per user %d",
				rs.Primary.Attributes[names.AttrUserPoolID], output.UserSessionTracking, aws.ToInt32(output.TokenSession.MaxPerUser))
		}

		return nil
	}
}

func isDefaultUserPoolSessionConfiguration(v *cognitoidentityprovider.GetUserPoolSessionConfigurationOutput) bool {
	return v.UserSessionTracking == awstypes.UserSessionTrackingTypeDisabled && aws.ToInt32(v.TokenSession.MaxPerUser) == 20
}

func testAccUserPoolSessionConfigurationConfig_basic(rName, tracking string, maxPerUser int) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name           = %[1]q
  user_pool_tier = "ESSENTIALS"
}

resource "aws_cognito_user_pool_session_configuration" "test" {
  user_pool_id          = aws_cognito_user_pool.test.id
  user_session_tracking = %[2]q

  token_session {
    max_per_user = %[3]d
  }
}
`, rName, tracking, maxPerUser)
}

func testAccUserPoolSessionConfigurationConfig_tier(rName, tier string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name           = %[1]q
  user_pool_tier = %[2]q
}

resource "aws_cognito_user_pool_session_configuration" "test" {
  user_pool_id          = aws_cognito_user_pool.test.id
  user_session_tracking = "ENABLED"

  token_session {
    max_per_user = 20
  }
}
`, rName, tier)
}
