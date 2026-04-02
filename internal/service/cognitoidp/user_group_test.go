// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	poolName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	groupName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	updatedGroupName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	updatedGroupNameUTF8 := strings.Repeat("あ", 128)
	resourceName := "aws_cognito_user_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupConfig_basic(poolName, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserGroupConfig_basic(poolName, updatedGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, updatedGroupName),
				),
			},
			{
				Config: testAccUserGroupConfig_basic(poolName, updatedGroupNameUTF8),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, updatedGroupNameUTF8),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	poolName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	groupName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupConfig_basic(poolName, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserGroupExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcognitoidp.ResourceUserGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPUserGroup_complex(t *testing.T) {
	ctx := acctest.Context(t)
	poolName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	groupName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	updatedGroupName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupConfig_complex(poolName, groupName, "This is the user group description", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is the user group description"),
					resource.TestCheckResourceAttr(resourceName, "precedence", "1"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserGroupConfig_complex(poolName, updatedGroupName, "This is the updated user group description", 42),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, updatedGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is the updated user group description"),
					resource.TestCheckResourceAttr(resourceName, "precedence", "42"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserGroup_roleARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupConfig_roleARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserGroupConfig_roleARNUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
				),
			},
		},
	})
}

func testAccCheckUserGroupExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		_, err := tfcognitoidp.FindGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccCheckUserGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_group" {
				continue
			}

			_, err := tfcognitoidp.FindGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito User Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccUserGroupConfig_basic(poolName, groupName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_group" "test" {
  name         = %[2]q
  user_pool_id = aws_cognito_user_pool.test.id
}
`, poolName, groupName)
}

func testAccUserGroupConfig_complex(poolName, groupName, groupDescription string, precedence int) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name = %[2]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Federated": "cognito-identity.amazonaws.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "cognito-identity.amazonaws.com:aud": "${data.aws_region.current.region}:12345678-dead-beef-cafe-123456790ab"
        },
        "ForAnyValue:StringLike": {
          "cognito-identity.amazonaws.com:amr": "authenticated"
        }
      }
    }
  ]
}
EOF
}

resource "aws_cognito_user_group" "test" {
  name         = %[2]q
  user_pool_id = aws_cognito_user_pool.test.id
  description  = %[3]q
  precedence   = %[4]d
  role_arn     = aws_iam_role.test.arn
}
`, poolName, groupName, groupDescription, precedence)
}

func testAccUserGroupConfig_roleARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_identity_pool" "test" {
  identity_pool_name               = %[1]q
  allow_unauthenticated_identities = false
}

resource "aws_iam_role" "test1" {
  name = "%[1]s-1"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Principal": {
      "Federated": "cognito-identity.amazonaws.com"
    },
    "Condition": {
      "StringEquals": {
        "cognito-identity.amazonaws.com:aud": [
            "${aws_cognito_identity_pool.test.identity_pool_name}"
        ]
      }
    }
  }]
}
EOF
}

resource "aws_cognito_user_group" "test" {
  name         = "%[1]s/a/b/c/test"
  user_pool_id = aws_cognito_user_pool.test.id
  role_arn     = aws_iam_role.test1.arn
}
`, rName)
}

func testAccUserGroupConfig_roleARNUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_identity_pool" "test" {
  identity_pool_name               = %[1]q
  allow_unauthenticated_identities = false
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Principal": {
      "Federated": "cognito-identity.amazonaws.com"
    },
    "Condition": {
      "StringEquals": {
        "cognito-identity.amazonaws.com:aud": [
            "${aws_cognito_identity_pool.test.identity_pool_name}"
        ]
      }
    }
  }]
}
EOF
}

resource "aws_cognito_user_group" "test" {
  name         = "%[1]s/a/b/c/test"
  user_pool_id = aws_cognito_user_pool.test.id
  role_arn     = aws_iam_role.test2.arn
}
`, rName)
}
