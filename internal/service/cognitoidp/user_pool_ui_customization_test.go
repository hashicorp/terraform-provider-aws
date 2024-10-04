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

func TestAccCognitoIDPUserPoolUICustomization_AllClients_CSS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	css := ".label-customizable {font-weight: 400;}"
	cssUpdated := ".label-customizable {font-weight: 100;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolUICustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolUICustomizationConfig_allClientsCSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientID, "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolUICustomizationConfig_allClientsCSS(rName, cssUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", cssUpdated),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientID, "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
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

func TestAccCognitoIDPUserPoolUICustomization_AllClients_disappears(t *testing.T) { // nosemgrep:ci.acceptance-test-naming-parent-disappears
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_ui_customization.test"

	css := ".label-customizable {font-weight: 400;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolUICustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolUICustomizationConfig_allClientsCSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceUserPoolUICustomization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolUICustomization_AllClients_imageFile(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	filename := "test-fixtures/logo.png"
	updatedFilename := "test-fixtures/logo_modified.png"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolUICustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolUICustomizationConfig_allClientsImage(rName, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientID, "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
			{
				Config: testAccUserPoolUICustomizationConfig_allClientsImage(rName, updatedFilename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientID, "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolUICustomization_AllClients_CSSAndImageFile(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	css := ".label-customizable {font-weight: 400;}"
	filename := "test-fixtures/logo.png"
	updatedFilename := "test-fixtures/logo_modified.png"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolUICustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolUICustomizationConfig_allClientsCSSAndImage(rName, css, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientID, "ALL"),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
			{
				Config: testAccUserPoolUICustomizationConfig_allClientsCSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientID, "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				Config: testAccUserPoolUICustomizationConfig_allClientsImage(rName, updatedFilename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientID, "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolUICustomization_Client_CSS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	clientResourceName := "aws_cognito_user_pool_client.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	css := ".label-customizable {font-weight: 400;}"
	cssUpdated := ".label-customizable {font-weight: 100;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolUICustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolUICustomizationConfig_clientCSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClientID, clientResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolUICustomizationConfig_clientCSS(rName, cssUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", cssUpdated),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClientID, clientResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
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

func TestAccCognitoIDPUserPoolUICustomization_Client_disappears(t *testing.T) { // nosemgrep:ci.acceptance-test-naming-parent-disappears
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_ui_customization.test"

	css := ".label-customizable {font-weight: 400;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolUICustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolUICustomizationConfig_clientCSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceUserPoolUICustomization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolUICustomization_Client_image(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	clientResourceName := "aws_cognito_user_pool_client.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	filename := "test-fixtures/logo.png"
	updatedFilename := "test-fixtures/logo_modified.png"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolUICustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolUICustomizationConfig_clientImage(rName, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClientID, clientResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
			{
				Config: testAccUserPoolUICustomizationConfig_clientImage(rName, updatedFilename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClientID, clientResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolUICustomization_ClientAndAll_cSS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_ui_customization.ui_all"
	clientUIResourceName := "aws_cognito_user_pool_ui_customization.ui_client"

	clientResourceName := "aws_cognito_user_pool_client.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	allCSS := ".label-customizable {font-weight: 400;}"
	clientCSS := ".label-customizable {font-weight: 100;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolUICustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Test UI Customization settings shared by ALL and a specific client
				Config: testAccUserPoolUICustomizationConfig_clientAndAllCSS(rName, allCSS, allCSS),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					testAccCheckUserPoolUICustomizationExists(ctx, clientUIResourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(resourceName, "css", allCSS),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(clientUIResourceName, names.AttrClientID, clientResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(clientUIResourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(clientUIResourceName, "css", allCSS),
					resource.TestCheckResourceAttrSet(clientUIResourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(clientUIResourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      clientUIResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test UI Customization settings overridden for the client
				Config: testAccUserPoolUICustomizationConfig_clientAndAllCSS(rName, allCSS, clientCSS),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					testAccCheckUserPoolUICustomizationExists(ctx, clientUIResourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(resourceName, "css", allCSS),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(clientUIResourceName, names.AttrClientID, clientResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(clientUIResourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(clientUIResourceName, "css", clientCSS),
					resource.TestCheckResourceAttrSet(clientUIResourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(clientUIResourceName, names.AttrUserPoolID, userPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      clientUIResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolUICustomization_UpdateClientToAll_cSS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	clientResourceName := "aws_cognito_user_pool_client.test"

	css := ".label-customizable {font-weight: 100;}"
	cssUpdated := ".label-customizable {font-weight: 400;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolUICustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolUICustomizationConfig_clientCSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClientID, clientResourceName, names.AttrID),
				),
			},
			{
				Config: testAccUserPoolUICustomizationConfig_allClientsCSS(rName, cssUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", cssUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientID, "ALL"),
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

func TestAccCognitoIDPUserPoolUICustomization_UpdateAllToClient_cSS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	clientResourceName := "aws_cognito_user_pool_client.test"

	css := ".label-customizable {font-weight: 100;}"
	cssUpdated := ".label-customizable {font-weight: 400;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolUICustomizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolUICustomizationConfig_allClientsCSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientID, "ALL"),
				),
			},
			{
				Config: testAccUserPoolUICustomizationConfig_clientCSS(rName, cssUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolUICustomizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", cssUpdated),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClientID, clientResourceName, names.AttrID),
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

func testAccCheckUserPoolUICustomizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_pool_ui_customization" {
				continue
			}

			_, err := tfcognitoidp.FindUserPoolUICustomizationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrClientID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito User Pool UI Customization %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserPoolUICustomizationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		_, err := tfcognitoidp.FindUserPoolUICustomizationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrClientID])

		return err
	}
}

func testAccUserPoolUICustomizationConfig_allClientsCSS(rName, css string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "test" {
  css = %[2]q

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state 
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, css)
}

func testAccUserPoolUICustomizationConfig_allClientsImage(rName, filename string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "test" {
  image_file = filebase64(%[2]q)

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state 
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, filename)
}

func testAccUserPoolUICustomizationConfig_allClientsCSSAndImage(rName, css, filename string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "test" {
  css        = %[2]q
  image_file = filebase64(%[3]q)

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state 
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, css, filename)
}

func testAccUserPoolUICustomizationConfig_clientCSS(rName, css string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "test" {
  client_id = aws_cognito_user_pool_client.test.id
  css       = %[2]q

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state 
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, css)
}

func testAccUserPoolUICustomizationConfig_clientImage(rName, filename string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "test" {
  client_id  = aws_cognito_user_pool_client.test.id
  image_file = filebase64(%[2]q)

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, filename)
}

func testAccUserPoolUICustomizationConfig_clientAndAllCSS(rName, allCSS, clientCSS string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "ui_all" {
  css = %[2]q

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}

resource "aws_cognito_user_pool_ui_customization" "ui_client" {
  client_id = aws_cognito_user_pool_client.test.id
  css       = %[3]q

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, allCSS, clientCSS)
}
