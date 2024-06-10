// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfcognitoidentity "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidentity"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIdentityPoolRolesAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cognito_identity_pool_roles_attachment.test"
	name := sdkacctest.RandString(10)
	updatedName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolRolesAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolRolesAttachmentConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolRolesAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "roles.authenticated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolRolesAttachmentConfig_basic(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolRolesAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "roles.authenticated"),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPoolRolesAttachment_roleMappings(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cognito_identity_pool_roles_attachment.test"
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolRolesAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolRolesAttachmentConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolRolesAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "role_mapping.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "roles.authenticated"),
				),
			},
			{
				Config: testAccPoolRolesAttachmentConfig_roleMappings(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolRolesAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "role_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "roles.authenticated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolRolesAttachmentConfig_roleMappingsUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolRolesAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "role_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "roles.authenticated"),
				),
			},
			{
				Config: testAccPoolRolesAttachmentConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolRolesAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "role_mapping.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "roles.authenticated"),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPoolRolesAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cognito_identity_pool_roles_attachment.test"
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolRolesAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolRolesAttachmentConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolRolesAttachmentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidentity.ResourcePoolRolesAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIdentityPoolRolesAttachment_roleMappingsWithAmbiguousRoleResolutionError(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolRolesAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccPoolRolesAttachmentConfig_roleMappingsWithAmbiguousRoleResolutionError(name),
				ExpectError: regexache.MustCompile(`validating ambiguous role resolution`),
			},
		},
	})
}

func TestAccCognitoIdentityPoolRolesAttachment_roleMappingsWithRulesTypeError(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolRolesAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccPoolRolesAttachmentConfig_roleMappingsWithRulesTypeError(name),
				ExpectError: regexache.MustCompile(`mapping_rule is required for Rules`),
			},
		},
	})
}

func TestAccCognitoIdentityPoolRolesAttachment_roleMappingsWithTokenTypeError(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolRolesAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccPoolRolesAttachmentConfig_roleMappingsWithTokenTypeError(name),
				ExpectError: regexache.MustCompile(`mapping_rule must not be set for Token based role mapping`),
			},
		},
	})
}

func testAccCheckPoolRolesAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito Identity Pool Roles Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIdentityClient(ctx)

		_, err := conn.GetIdentityPoolRoles(ctx, &cognitoidentity.GetIdentityPoolRolesInput{
			IdentityPoolId: aws.String(rs.Primary.Attributes["identity_pool_id"]),
		})

		return err
	}
}

func testAccCheckPoolRolesAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIdentityClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_identity_pool_roles_attachment" {
				continue
			}

			_, err := conn.GetIdentityPoolRoles(ctx, &cognitoidentity.GetIdentityPoolRolesInput{
				IdentityPoolId: aws.String(rs.Primary.Attributes["identity_pool_id"]),
			})

			if err != nil {
				if errs.IsA[*awstypes.ResourceNotFoundException](err) {
					return nil
				}
				return err
			}
		}

		return nil
	}
}

func testAccPoolRolesAttachmentConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %[1]s"
  allow_unauthenticated_identities = false

  supported_login_providers = {
    "graph.facebook.com" = "7346241598935555"
  }
}

# Unauthenticated Role
resource "aws_iam_role" "unauthenticated" {
  name = "cognito_unauthenticated_%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "cognito-identity.amazonaws.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "cognito-identity.amazonaws.com:aud": "${aws_cognito_identity_pool.main.id}"
        },
        "ForAnyValue:StringLike": {
          "cognito-identity.amazonaws.com:amr": "unauthenticated"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "unauthenticated" {
  name = "unauthenticated_policy_%[1]s"
  role = aws_iam_role.unauthenticated.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "mobileanalytics:PutEvents",
        "cognito-sync:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

# Authenticated Role
resource "aws_iam_role" "authenticated" {
  name = "cognito_authenticated_%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "cognito-identity.amazonaws.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "cognito-identity.amazonaws.com:aud": "${aws_cognito_identity_pool.main.id}"
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

resource "aws_iam_role_policy" "authenticated" {
  name = "authenticated_policy_%[1]s"
  role = aws_iam_role.authenticated.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "mobileanalytics:PutEvents",
        "cognito-sync:*",
        "cognito-identity:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, name)
}

func testAccPoolRolesAttachmentConfig_basic(name string) string {
	return fmt.Sprintf(testAccPoolRolesAttachmentConfig(name) + `
resource "aws_cognito_identity_pool_roles_attachment" "test" {
  identity_pool_id = aws_cognito_identity_pool.main.id

  roles = {
    "authenticated" = aws_iam_role.authenticated.arn
  }
}
`)
}

func testAccPoolRolesAttachmentConfig_roleMappings(name string) string {
	return fmt.Sprintf(testAccPoolRolesAttachmentConfig(name) + `
resource "aws_cognito_identity_pool_roles_attachment" "test" {
  identity_pool_id = aws_cognito_identity_pool.main.id

  role_mapping {
    identity_provider         = "graph.facebook.com"
    ambiguous_role_resolution = "AuthenticatedRole"
    type                      = "Rules"

    mapping_rule {
      claim      = "isAdmin"
      match_type = "Equals"
      role_arn   = aws_iam_role.authenticated.arn
      value      = "paid"
    }
  }

  roles = {
    "authenticated" = aws_iam_role.authenticated.arn
  }
}
`)
}

func testAccPoolRolesAttachmentConfig_roleMappingsUpdated(name string) string {
	return fmt.Sprintf(testAccPoolRolesAttachmentConfig(name) + `
resource "aws_cognito_identity_pool_roles_attachment" "test" {
  identity_pool_id = aws_cognito_identity_pool.main.id

  role_mapping {
    identity_provider         = "graph.facebook.com"
    ambiguous_role_resolution = "AuthenticatedRole"
    type                      = "Rules"

    mapping_rule {
      claim      = "isPaid"
      match_type = "Equals"
      role_arn   = aws_iam_role.authenticated.arn
      value      = "unpaid"
    }

    mapping_rule {
      claim      = "isFoo"
      match_type = "Equals"
      role_arn   = aws_iam_role.authenticated.arn
      value      = "bar"
    }
  }

  roles = {
    "authenticated" = aws_iam_role.authenticated.arn
  }
}
`)
}

func testAccPoolRolesAttachmentConfig_roleMappingsWithAmbiguousRoleResolutionError(name string) string {
	return fmt.Sprintf(testAccPoolRolesAttachmentConfig(name) + `
resource "aws_cognito_identity_pool_roles_attachment" "test" {
  identity_pool_id = aws_cognito_identity_pool.main.id

  role_mapping {
    identity_provider = "graph.facebook.com"
    type              = "Rules"

    mapping_rule {
      claim      = "isAdmin"
      match_type = "Equals"
      role_arn   = aws_iam_role.authenticated.arn
      value      = "paid"
    }
  }

  roles = {
    "authenticated" = aws_iam_role.authenticated.arn
  }
}
`)
}

func testAccPoolRolesAttachmentConfig_roleMappingsWithRulesTypeError(name string) string {
	return fmt.Sprintf(testAccPoolRolesAttachmentConfig(name) + `
resource "aws_cognito_identity_pool_roles_attachment" "test" {
  identity_pool_id = aws_cognito_identity_pool.main.id

  role_mapping {
    identity_provider         = "graph.facebook.com"
    ambiguous_role_resolution = "AuthenticatedRole"
    type                      = "Rules"
  }

  roles = {
    "authenticated" = aws_iam_role.authenticated.arn
  }
}
`)
}

func testAccPoolRolesAttachmentConfig_roleMappingsWithTokenTypeError(name string) string {
	return fmt.Sprintf(testAccPoolRolesAttachmentConfig(name) + `
resource "aws_cognito_identity_pool_roles_attachment" "test" {
  identity_pool_id = aws_cognito_identity_pool.main.id

  role_mapping {
    identity_provider         = "graph.facebook.com"
    ambiguous_role_resolution = "AuthenticatedRole"
    type                      = "Token"

    mapping_rule {
      claim      = "isAdmin"
      match_type = "Equals"
      role_arn   = aws_iam_role.authenticated.arn
      value      = "paid"
    }
  }

  roles = {
    "authenticated" = aws_iam_role.authenticated.arn
  }
}
`)
}
