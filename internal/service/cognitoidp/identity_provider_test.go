// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPIdentityProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var identityProvider awstypes.IdentityProviderType
	resourceName := "aws_cognito_identity_provider.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.username", "sub"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.%", "9"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.authorize_scopes", names.AttrEmail),
					resource.TestCheckResourceAttr(resourceName, "provider_details.authorize_url", "https://accounts.google.com/o/oauth2/v2/auth"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.client_id", "test-url.apps.googleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.client_secret", names.AttrClientSecret),
					resource.TestCheckResourceAttr(resourceName, "provider_details.attributes_url", "https://people.googleapis.com/v1/people/me?personFields="),
					resource.TestCheckResourceAttr(resourceName, "provider_details.attributes_url_add_attributes", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "provider_details.token_request_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.token_url", "https://www.googleapis.com/oauth2/v4/token"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.oidc_issuer", "https://accounts.google.com"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, "Google"),
					resource.TestCheckResourceAttr(resourceName, "provider_type", "Google"),
				),
			},
			{
				Config: testAccIdentityProviderConfig_basicUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.username", "sub"),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.email", names.AttrEmail),
					resource.TestCheckResourceAttr(resourceName, "provider_details.%", "9"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.authorize_scopes", names.AttrEmail),
					resource.TestCheckResourceAttr(resourceName, "provider_details.authorize_url", "https://accounts.google.com/o/oauth2/v2/auth"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.client_id", "new-client-id-url.apps.googleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.client_secret", "updated_client_secret"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.attributes_url", "https://people.googleapis.com/v1/people/me?personFields="),
					resource.TestCheckResourceAttr(resourceName, "provider_details.attributes_url_add_attributes", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "provider_details.token_request_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.token_url", "https://www.googleapis.com/oauth2/v4/token"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.oidc_issuer", "https://accounts.google.com"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, "Google"),
					resource.TestCheckResourceAttr(resourceName, "provider_type", "Google"),
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

func TestAccCognitoIDPIdentityProvider_idpIdentifiers(t *testing.T) {
	ctx := acctest.Context(t)
	var identityProvider awstypes.IdentityProviderType
	resourceName := "aws_cognito_identity_provider.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityProviderConfig_identifier(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "idp_identifiers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "idp_identifiers.0", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIdentityProviderConfig_identifier(rName, "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "idp_identifiers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "idp_identifiers.0", "test2"),
				),
			},
		},
	})
}

func TestAccCognitoIDPIdentityProvider_saml(t *testing.T) {
	ctx := acctest.Context(t)
	var identityProvider awstypes.IdentityProviderType
	resourceName := "aws_cognito_identity_provider.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityProviderConfig_saml(rName, acctest.CtFalse),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.email", names.AttrEmail),
					resource.TestCheckNoResourceAttr(resourceName, "idp_identifiers.#"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.%", acctest.Ct4),
					resource.TestCheckResourceAttrSet(resourceName, "provider_details.ActiveEncryptionCertificate"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.EncryptedResponses", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "provider_details.MetadataFile"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.SSORedirectBindingURI", "https://terraform-dev-ed.my.salesforce.com/idp/endpoint/HttpRedirect"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, rName),
					resource.TestCheckResourceAttr(resourceName, "provider_type", "SAML"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIdentityProviderConfig_saml(rName, acctest.CtTrue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.email", names.AttrEmail),
					resource.TestCheckNoResourceAttr(resourceName, "idp_identifiers.#"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.%", acctest.Ct4),
					resource.TestCheckResourceAttrSet(resourceName, "provider_details.ActiveEncryptionCertificate"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.EncryptedResponses", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "provider_details.MetadataFile"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.SSORedirectBindingURI", "https://terraform-dev-ed.my.salesforce.com/idp/endpoint/HttpRedirect"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, rName),
					resource.TestCheckResourceAttr(resourceName, "provider_type", "SAML"),
				),
			},
		},
	})
}

func TestAccCognitoIDPIdentityProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var identityProvider awstypes.IdentityProviderType
	resourceName := "aws_cognito_identity_provider.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, resourceName, &identityProvider),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceIdentityProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPIdentityProvider_Disappears_userPool(t *testing.T) {
	ctx := acctest.Context(t)
	var identityProvider awstypes.IdentityProviderType
	resourceName := "aws_cognito_identity_provider.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, resourceName, &identityProvider),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceUserPool(), "aws_cognito_user_pool.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIdentityProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_identity_provider" {
				continue
			}

			_, err := tfcognitoidp.FindIdentityProviderByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrProviderName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito Identity Provider %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIdentityProviderExists(ctx context.Context, n string, v *awstypes.IdentityProviderType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindIdentityProviderByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrProviderName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccIdentityProviderConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = %[1]q
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_identity_provider" "test" {
  user_pool_id  = aws_cognito_user_pool.test.id
  provider_name = "Google"
  provider_type = "Google"

  provider_details = {
    attributes_url                = "https://people.googleapis.com/v1/people/me?personFields="
    attributes_url_add_attributes = "true"
    authorize_scopes              = "email"
    authorize_url                 = "https://accounts.google.com/o/oauth2/v2/auth"
    client_id                     = "test-url.apps.googleusercontent.com"
    client_secret                 = "client_secret"
    oidc_issuer                   = "https://accounts.google.com"
    token_request_method          = "POST"
    token_url                     = "https://www.googleapis.com/oauth2/v4/token"
  }
}
`, rName)
}

func testAccIdentityProviderConfig_basicUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = %[1]q
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_identity_provider" "test" {
  user_pool_id  = aws_cognito_user_pool.test.id
  provider_name = "Google"
  provider_type = "Google"

  provider_details = {
    attributes_url                = "https://people.googleapis.com/v1/people/me?personFields="
    attributes_url_add_attributes = "true"
    authorize_scopes              = "email"
    authorize_url                 = "https://accounts.google.com/o/oauth2/v2/auth"
    client_id                     = "new-client-id-url.apps.googleusercontent.com"
    client_secret                 = "updated_client_secret"
    oidc_issuer                   = "https://accounts.google.com"
    token_request_method          = "POST"
    token_url                     = "https://www.googleapis.com/oauth2/v4/token"
  }

  attribute_mapping = {
    email    = "email"
    username = "sub"
  }
}
`, rName)
}

func testAccIdentityProviderConfig_identifier(rName, attribute string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = %[1]q
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_identity_provider" "test" {
  user_pool_id  = aws_cognito_user_pool.test.id
  provider_name = "Google"
  provider_type = "Google"

  idp_identifiers = [%[2]q]

  provider_details = {
    attributes_url                = "https://people.googleapis.com/v1/people/me?personFields="
    attributes_url_add_attributes = "true"
    authorize_scopes              = "email"
    authorize_url                 = "https://accounts.google.com/o/oauth2/v2/auth"
    client_id                     = "test-url.apps.googleusercontent.com"
    client_secret                 = "client_secret"
    oidc_issuer                   = "https://accounts.google.com"
    token_request_method          = "POST"
    token_url                     = "https://www.googleapis.com/oauth2/v4/token"
  }
}
`, rName, attribute)
}

func testAccIdentityProviderConfig_saml(rName, encryptedResponses string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = %[1]q
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_identity_provider" "test" {
  user_pool_id  = aws_cognito_user_pool.test.id
  provider_name = %[1]q
  provider_type = "SAML"

  provider_details = {
    EncryptedResponses    = %[2]q
    MetadataFile          = file("./test-fixtures/saml-metadata.xml")
    SSORedirectBindingURI = "https://terraform-dev-ed.my.salesforce.com/idp/endpoint/HttpRedirect"
  }

  attribute_mapping = {
    email = "email"
  }

  lifecycle {
    ignore_changes = [provider_details["ActiveEncryptionCertificate"]]
  }
}
`, rName, encryptedResponses)
}
