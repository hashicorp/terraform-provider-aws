// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserInGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_in_group.test"
	userPoolResourceName := "aws_cognito_user_pool.test"
	userGroupResourceName := "aws_cognito_user_group.test"
	userResourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserInGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserInGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserInGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, userGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUsername, userResourceName, names.AttrUsername),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserInGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_in_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserInGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserInGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserInGroupExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceUserInGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
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

func testAccCheckUserInGroupExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		return tfcognitoidp.FindGroupUserByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrGroupName], rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrUsername])
	}
}

func testAccCheckUserInGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_in_group" {
				continue
			}

			err := tfcognitoidp.FindGroupUserByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrGroupName], rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrUsername])

			if tfresource.NotFound(err) {
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
