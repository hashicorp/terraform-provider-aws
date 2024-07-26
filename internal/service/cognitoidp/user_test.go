// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
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

func TestAccCognitoIDPUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrSet(resourceName, "sub"),
					resource.TestCheckResourceAttr(resourceName, "preferred_mfa_setting", ""),
					resource.TestCheckResourceAttr(resourceName, "mfa_setting_list.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.UserStatusTypeForceChangePassword)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"temporary_password",
					names.AttrPassword,
					"client_metadata",
					"validation_data",
					"desired_delivery_mediums",
					"message_action",
				},
			},
		},
	})
}

func TestAccCognitoIDPUser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPUser_temporaryPassword(t *testing.T) {
	ctx := acctest.Context(t)
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rClientName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserPassword := sdkacctest.RandString(16)
	rUserPasswordUpdated := sdkacctest.RandString(16)
	userResourceName := "aws_cognito_user.test"
	clientResourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_temporaryPassword(rUserPoolName, rClientName, rUserName, rUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName),
					testAccUserTemporaryPassword(ctx, userResourceName, clientResourceName),
					resource.TestCheckResourceAttr(userResourceName, names.AttrStatus, string(awstypes.UserStatusTypeForceChangePassword)),
				),
			},
			{
				ResourceName:      userResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"temporary_password",
					names.AttrPassword,
					"client_metadata",
					"validation_data",
					"desired_delivery_mediums",
					"message_action",
				},
			},
			{
				Config: testAccUserConfig_temporaryPassword(rUserPoolName, rClientName, rUserName, rUserPasswordUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName),
					testAccUserTemporaryPassword(ctx, userResourceName, clientResourceName),
					resource.TestCheckResourceAttr(userResourceName, names.AttrStatus, string(awstypes.UserStatusTypeForceChangePassword)),
				),
			},
			{
				Config: testAccUserConfig_noPassword(rUserPoolName, rClientName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName),
					resource.TestCheckResourceAttr(userResourceName, "temporary_password", ""),
					resource.TestCheckResourceAttr(userResourceName, names.AttrStatus, string(awstypes.UserStatusTypeForceChangePassword)),
				),
			},
		},
	})
}

func TestAccCognitoIDPUser_password(t *testing.T) {
	ctx := acctest.Context(t)
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rClientName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserPassword := sdkacctest.RandString(16)
	rUserPasswordUpdated := sdkacctest.RandString(16)
	userResourceName := "aws_cognito_user.test"
	clientResourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_password(rUserPoolName, rClientName, rUserName, rUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName),
					testAccUserPassword(ctx, userResourceName, clientResourceName),
					resource.TestCheckResourceAttr(userResourceName, names.AttrStatus, string(awstypes.UserStatusTypeConfirmed)),
				),
			},
			{
				ResourceName:      userResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"temporary_password",
					names.AttrPassword,
					"client_metadata",
					"validation_data",
					"desired_delivery_mediums",
					"message_action",
				},
			},
			{
				Config: testAccUserConfig_password(rUserPoolName, rClientName, rUserName, rUserPasswordUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName),
					testAccUserPassword(ctx, userResourceName, clientResourceName),
					resource.TestCheckResourceAttr(userResourceName, names.AttrStatus, string(awstypes.UserStatusTypeConfirmed)),
				),
			},
			{
				Config: testAccUserConfig_noPassword(rUserPoolName, rClientName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName),
					resource.TestCheckResourceAttr(userResourceName, names.AttrPassword, ""),
					resource.TestCheckResourceAttr(userResourceName, names.AttrStatus, string(awstypes.UserStatusTypeConfirmed)),
				),
			},
		},
	})
}

func TestAccCognitoIDPUser_attributes(t *testing.T) {
	ctx := acctest.Context(t)
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_attributes(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "attributes.one", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "attributes.two", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "attributes.three", acctest.Ct3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"temporary_password",
					names.AttrPassword,
					"client_metadata",
					"validation_data",
					"desired_delivery_mediums",
					"message_action",
				},
			},
			{
				Config: testAccUserConfig_attributesUpdated(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "attributes.two", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "attributes.three", "three"),
					resource.TestCheckResourceAttr(resourceName, "attributes.four", acctest.Ct4),
				),
			},
		},
	})
}

func TestAccCognitoIDPUser_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_enable(rUserPoolName, rUserName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"temporary_password",
					names.AttrPassword,
					"client_metadata",
					"validation_data",
					"desired_delivery_mediums",
					"message_action",
				},
			},
			{
				Config: testAccUserConfig_enable(rUserPoolName, rUserName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/38175.
func TestAccCognitoIDPUser_v5560Regression(t *testing.T) {
	ctx := acctest.Context(t)
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := acctest.RandomEmailAddress(domain)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		CheckDestroy: testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.53.0",
					},
				},
				Config: testAccUserConfig_v5560Regression(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrSet(resourceName, "sub"),
					resource.TestCheckResourceAttr(resourceName, "preferred_mfa_setting", ""),
					resource.TestCheckResourceAttr(resourceName, "mfa_setting_list.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.UserStatusTypeForceChangePassword)),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccUserConfig_v5560Regression(rUserPoolName, rUserName),
				PlanOnly:                 true,
			},
		},
	})
}

func testAccCheckUserExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		_, err := tfcognitoidp.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrUsername])

		return err
	}
}

func testAccCheckUserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user" {
				continue
			}

			_, err := tfcognitoidp.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrUsername])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito User %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccUserTemporaryPassword(ctx context.Context, userRsName string, clientRsName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		userRs, ok := s.RootModule().Resources[userRsName]
		if !ok {
			return fmt.Errorf("Not found: %s", userRsName)
		}

		clientRs, ok := s.RootModule().Resources[clientRsName]
		if !ok {
			return fmt.Errorf("Not found: %s", clientRsName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		input := &cognitoidentityprovider.InitiateAuthInput{
			AuthFlow: awstypes.AuthFlowTypeUserPasswordAuth,
			AuthParameters: map[string]string{
				"USERNAME": userRs.Primary.Attributes[names.AttrUsername],
				"PASSWORD": userRs.Primary.Attributes["temporary_password"],
			},
			ClientId: aws.String(clientRs.Primary.Attributes[names.AttrID]),
		}

		output, err := conn.InitiateAuth(ctx, input)

		if err != nil {
			return err
		}

		if output.ChallengeName != awstypes.ChallengeNameTypeNewPasswordRequired {
			return errors.New("The password is not a temporary password")
		}

		return nil
	}
}

func testAccUserPassword(ctx context.Context, userRsName string, clientRsName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		userRs, ok := s.RootModule().Resources[userRsName]
		if !ok {
			return fmt.Errorf("Not found: %s", userRsName)
		}

		clientRs, ok := s.RootModule().Resources[clientRsName]
		if !ok {
			return fmt.Errorf("Not found: %s", clientRsName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		input := &cognitoidentityprovider.InitiateAuthInput{
			AuthFlow: awstypes.AuthFlowTypeUserPasswordAuth,
			AuthParameters: map[string]string{
				"USERNAME": userRs.Primary.Attributes[names.AttrUsername],
				"PASSWORD": userRs.Primary.Attributes[names.AttrPassword],
			},
			ClientId: aws.String(clientRs.Primary.Attributes[names.AttrID]),
		}

		_, err := conn.InitiateAuth(ctx, input)

		return err
	}
}

func testAccUserConfig_basic(userPoolName string, userName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  username     = %[2]q
}
`, userPoolName, userName)
}

func testAccUserConfig_temporaryPassword(userPoolName string, clientName string, userName string, password string) string {
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

resource "aws_cognito_user_pool_client" "test" {
  name                = %[2]q
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ALLOW_USER_PASSWORD_AUTH", "ALLOW_REFRESH_TOKEN_AUTH"]
}

resource "aws_cognito_user" "test" {
  user_pool_id       = aws_cognito_user_pool.test.id
  username           = %[3]q
  temporary_password = %[4]q
}
`, userPoolName, clientName, userName, password)
}

func testAccUserConfig_password(userPoolName string, clientName string, userName string, password string) string {
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

resource "aws_cognito_user_pool_client" "test" {
  name                = %[2]q
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ALLOW_USER_PASSWORD_AUTH", "ALLOW_REFRESH_TOKEN_AUTH"]
}

resource "aws_cognito_user" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  username     = %[3]q
  password     = %[4]q
}
`, userPoolName, clientName, userName, password)
}

func testAccUserConfig_noPassword(userPoolName string, clientName string, userName string) string {
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

resource "aws_cognito_user_pool_client" "test" {
  name                = %[2]q
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ALLOW_USER_PASSWORD_AUTH", "ALLOW_REFRESH_TOKEN_AUTH"]
}

resource "aws_cognito_user" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  username     = %[3]q
}
`, userPoolName, clientName, userName)
}

func testAccUserConfig_attributes(userPoolName string, userName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  schema {
    name                     = "one"
    attribute_data_type      = "String"
    mutable                  = true
    required                 = false
    developer_only_attribute = false
    string_attribute_constraints {}
  }
  schema {
    name                     = "two"
    attribute_data_type      = "String"
    mutable                  = true
    required                 = false
    developer_only_attribute = false
    string_attribute_constraints {}
  }
  schema {
    name                     = "three"
    attribute_data_type      = "String"
    mutable                  = true
    required                 = false
    developer_only_attribute = false
    string_attribute_constraints {}
  }
  schema {
    name                     = "four"
    attribute_data_type      = "String"
    mutable                  = true
    required                 = false
    developer_only_attribute = false
    string_attribute_constraints {}
  }
}

resource "aws_cognito_user" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  username     = %[2]q

  attributes = {
    one   = "1"
    two   = "2"
    three = "3"
  }
}
`, userPoolName, userName)
}

func testAccUserConfig_attributesUpdated(userPoolName string, userName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  schema {
    name                     = "one"
    attribute_data_type      = "String"
    mutable                  = true
    required                 = false
    developer_only_attribute = false
    string_attribute_constraints {}
  }
  schema {
    name                     = "two"
    attribute_data_type      = "String"
    mutable                  = true
    required                 = false
    developer_only_attribute = false
    string_attribute_constraints {}
  }
  schema {
    name                     = "three"
    attribute_data_type      = "String"
    mutable                  = true
    required                 = false
    developer_only_attribute = false
    string_attribute_constraints {}
  }
  schema {
    name                     = "four"
    attribute_data_type      = "String"
    mutable                  = true
    required                 = false
    developer_only_attribute = false
    string_attribute_constraints {}
  }
}

resource "aws_cognito_user" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  username     = %[2]q

  attributes = {
    two   = "2"
    three = "three"
    four  = "4"
  }
}
`, userPoolName, userName)
}

func testAccUserConfig_enable(userPoolName string, userName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  username     = %[2]q
  enabled      = %[3]t
}
`, userPoolName, userName, enabled)
}

func testAccUserConfig_v5560Regression(userPoolName string, userName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  username     = %[2]q

  attributes = {
    "name"           = "test"
    "email"          = %[2]q
    "email_verified" = "true"
  }
}
`, userPoolName, userName)
}
