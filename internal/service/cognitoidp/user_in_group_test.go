// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserInGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_in_group.test"
	userPoolResourceName := "aws_cognito_user_pool.test"
	userGroupResourceName := "aws_cognito_user_group.test"
	userResourceName := "aws_cognito_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserInGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserInGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserInGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, userGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUsername, userResourceName, names.AttrUsername),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserInGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_in_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserInGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserInGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserInGroupExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcognitoidp.ResourceUserInGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccCognitoIDPUserInGroup_upgrade_v6_0_0(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_in_group.test"
	userPoolResourceName := "aws_cognito_user_pool.test"
	userGroupResourceName := "aws_cognito_user_group.test"
	userResourceName := "aws_cognito_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		CheckDestroy: testAccCheckUserInGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.96.0",
					},
				},
				Config: testAccUserInGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserInGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, userGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUsername, userResourceName, names.AttrUsername),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccUserInGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserInGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, userGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUsername, userResourceName, names.AttrUsername),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckUserInGroupExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		return tfcognitoidp.FindGroupUserByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrGroupName], rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrUsername])
	}
}

func testAccCheckUserInGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_in_group" {
				continue
			}

			err := tfcognitoidp.FindGroupUserByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrGroupName], rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrUsername])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito Group User %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccUserInGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
  password_policy {
    temporary_password_validity_days = 7
    minimum_length                   = 6
    require_uppercase                = false
    require_symbols                  = false
    require_numbers                  = false
  }
}

resource "aws_cognito_user" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  username     = %[1]q
}

resource "aws_cognito_user_group" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  name         = %[1]q
}

resource "aws_cognito_user_in_group" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  group_name   = aws_cognito_user_group.test.name
  username     = aws_cognito_user.test.username
}
`, rName)
}
