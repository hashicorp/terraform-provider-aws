// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

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

func TestAccCognitoIdentityPoolProviderPrincipalTags_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cognito_identity_pool_provider_principal_tag.test"
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolProviderPrincipalTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolProviderPrincipalTagsConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolProviderPrincipalTagsExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "principal_tags.test", names.AttrValue),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPoolProviderPrincipalTags_updated(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cognito_identity_pool_provider_principal_tag.test"
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolProviderPrincipalTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolProviderPrincipalTagsConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolProviderPrincipalTagsExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "principal_tags.test", names.AttrValue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolProviderPrincipalTagsConfig_tagsUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolProviderPrincipalTagsExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "principal_tags.test", names.AttrValue),
					resource.TestCheckResourceAttr(resourceName, "principal_tags.new", "map"),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPoolProviderPrincipalTags_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cognito_identity_pool_provider_principal_tag.test"
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolProviderPrincipalTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolProviderPrincipalTagsConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolProviderPrincipalTagsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidentity.ResourcePoolProviderPrincipalTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIdentityPoolProviderPrincipalTags_oidc(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cognito_identity_pool_provider_principal_tag.test"
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolProviderPrincipalTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolProviderPrincipalTagsConfig_oidc(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolProviderPrincipalTagsExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "principal_tags.test", names.AttrValue),
				),
			},
		},
	})
}

func testAccCheckPoolProviderPrincipalTagsExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito Identity Princpal Tags is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIdentityClient(ctx)

		_, err := conn.GetPrincipalTagAttributeMap(ctx, &cognitoidentity.GetPrincipalTagAttributeMapInput{
			IdentityPoolId:       aws.String(rs.Primary.Attributes["identity_pool_id"]),
			IdentityProviderName: aws.String(rs.Primary.Attributes["identity_provider_name"]),
		})

		return err
	}
}

func testAccCheckPoolProviderPrincipalTagsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIdentityClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_identity_pool_provider_principal_tag" {
				continue
			}

			_, err := conn.GetPrincipalTagAttributeMap(ctx, &cognitoidentity.GetPrincipalTagAttributeMapInput{
				IdentityPoolId:       aws.String(rs.Primary.Attributes["identity_pool_id"]),
				IdentityProviderName: aws.String(rs.Primary.Attributes["identity_provider_name"]),
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

func testAccPoolProviderPrincipalTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = %[1]q
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
  supported_identity_providers = compact([
    "COGNITO",
  ])
}

resource "aws_cognito_identity_pool" "test" {
  identity_pool_name               = %[1]q
  allow_unauthenticated_identities = false
  cognito_identity_providers {
    client_id               = aws_cognito_user_pool_client.test.id
    provider_name           = aws_cognito_user_pool.test.endpoint
    server_side_token_check = false
  }
  supported_login_providers = {
    "accounts.google.com" = "new-client-id-url.apps.googleusercontent.com"
  }
}
`, name)
}

func testAccPoolProviderPrincipalTagsConfig_basic(name string) string {
	return acctest.ConfigCompose(testAccPoolProviderPrincipalTagsConfig(name), `
resource "aws_cognito_identity_pool_provider_principal_tag" "test" {
  identity_pool_id       = aws_cognito_identity_pool.test.id
  identity_provider_name = aws_cognito_user_pool.test.endpoint
  use_defaults           = false
  principal_tags = {
    test = "value"
  }
}
`)
}

func testAccPoolProviderPrincipalTagsConfig_tagsUpdated(name string) string {
	return acctest.ConfigCompose(testAccPoolProviderPrincipalTagsConfig(name), `
resource "aws_cognito_identity_pool_provider_principal_tag" "test" {
  identity_pool_id       = aws_cognito_identity_pool.test.id
  identity_provider_name = aws_cognito_user_pool.test.endpoint
  use_defaults           = false
  principal_tags = {
    test = "value"
    new  = "map"
  }
}
`)
}

func testAccPoolProviderPrincipalTagsConfig_oidc(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "idp" {
  url = "https://accounts.example.com"
  client_id_list = [
    "sts.amazonaws.com"
  ]
  thumbprint_list = ["990f4193972f2becf12ddeda5237f9c952f20d9e"]
}

resource "aws_cognito_identity_pool" "pool" {
  identity_pool_name               = "%s"
  allow_unauthenticated_identities = false
  allow_classic_flow               = false

  openid_connect_provider_arns = [
    aws_iam_openid_connect_provider.idp.arn
  ]
}

resource "aws_cognito_identity_pool_provider_principal_tag" "test" {
  identity_pool_id       = aws_cognito_identity_pool.pool.id
  identity_provider_name = aws_iam_openid_connect_provider.idp.arn
  use_defaults           = false
  principal_tags = {
    test = "value"
  }
}`, name)
}
