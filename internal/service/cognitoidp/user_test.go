package cognitoidp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCognitoIDPUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrSet(resourceName, "sub"),
					resource.TestCheckResourceAttr(resourceName, "preferred_mfa_setting", ""),
					resource.TestCheckResourceAttr(resourceName, "mfa_setting_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", cognitoidentityprovider.UserStatusTypeForceChangePassword),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"temporary_password",
					"password",
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_temporaryPassword(rUserPoolName, rClientName, rUserName, rUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName),
					testAccUserTemporaryPassword(ctx, userResourceName, clientResourceName),
					resource.TestCheckResourceAttr(userResourceName, "status", cognitoidentityprovider.UserStatusTypeForceChangePassword),
				),
			},
			{
				ResourceName:      userResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"temporary_password",
					"password",
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
					resource.TestCheckResourceAttr(userResourceName, "status", cognitoidentityprovider.UserStatusTypeForceChangePassword),
				),
			},
			{
				Config: testAccUserConfig_noPassword(rUserPoolName, rClientName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName),
					resource.TestCheckResourceAttr(userResourceName, "temporary_password", ""),
					resource.TestCheckResourceAttr(userResourceName, "status", cognitoidentityprovider.UserStatusTypeForceChangePassword),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_password(rUserPoolName, rClientName, rUserName, rUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName),
					testAccUserPassword(ctx, userResourceName, clientResourceName),
					resource.TestCheckResourceAttr(userResourceName, "status", cognitoidentityprovider.UserStatusTypeConfirmed),
				),
			},
			{
				ResourceName:      userResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"temporary_password",
					"password",
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
					resource.TestCheckResourceAttr(userResourceName, "status", cognitoidentityprovider.UserStatusTypeConfirmed),
				),
			},
			{
				Config: testAccUserConfig_noPassword(rUserPoolName, rClientName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName),
					resource.TestCheckResourceAttr(userResourceName, "password", ""),
					resource.TestCheckResourceAttr(userResourceName, "status", cognitoidentityprovider.UserStatusTypeConfirmed),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_attributes(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "attributes.one", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.two", "2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.three", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"temporary_password",
					"password",
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
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "attributes.two", "2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.three", "three"),
					resource.TestCheckResourceAttr(resourceName, "attributes.four", "4"),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_enable(rUserPoolName, rUserName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"temporary_password",
					"password",
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
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
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

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cognito User ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn()

		_, err := tfcognitoidp.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes["user_pool_id"], rs.Primary.Attributes["username"])

		return err
	}
}

func testAccCheckUserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user" {
				continue
			}

			_, err := tfcognitoidp.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes["user_pool_id"], rs.Primary.Attributes["username"])

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

func testAccUserTemporaryPassword(ctx context.Context, userResName string, clientResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		userRs, ok := s.RootModule().Resources[userResName]
		if !ok {
			return fmt.Errorf("Not found: %s", userResName)
		}

		clientRs, ok := s.RootModule().Resources[clientResName]
		if !ok {
			return fmt.Errorf("Not found: %s", clientResName)
		}

		userName := userRs.Primary.Attributes["username"]
		userPassword := userRs.Primary.Attributes["temporary_password"]
		clientId := clientRs.Primary.Attributes["id"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn()

		params := &cognitoidentityprovider.InitiateAuthInput{
			AuthFlow: aws.String(cognitoidentityprovider.AuthFlowTypeUserPasswordAuth),
			AuthParameters: map[string]*string{
				"USERNAME": aws.String(userName),
				"PASSWORD": aws.String(userPassword),
			},
			ClientId: aws.String(clientId),
		}

		resp, err := conn.InitiateAuthWithContext(ctx, params)
		if err != nil {
			return err
		}

		if aws.StringValue(resp.ChallengeName) != cognitoidentityprovider.ChallengeNameTypeNewPasswordRequired {
			return errors.New("The password is not a temporary password.")
		}

		return nil
	}
}

func testAccUserPassword(ctx context.Context, userResName string, clientResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		userRs, ok := s.RootModule().Resources[userResName]
		if !ok {
			return fmt.Errorf("Not found: %s", userResName)
		}

		clientRs, ok := s.RootModule().Resources[clientResName]
		if !ok {
			return fmt.Errorf("Not found: %s", clientResName)
		}

		userName := userRs.Primary.Attributes["username"]
		userPassword := userRs.Primary.Attributes["password"]
		clientId := clientRs.Primary.Attributes["id"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn()

		params := &cognitoidentityprovider.InitiateAuthInput{
			AuthFlow: aws.String(cognitoidentityprovider.AuthFlowTypeUserPasswordAuth),
			AuthParameters: map[string]*string{
				"USERNAME": aws.String(userName),
				"PASSWORD": aws.String(userPassword),
			},
			ClientId: aws.String(clientId),
		}

		resp, err := conn.InitiateAuthWithContext(ctx, params)
		if err != nil {
			return err
		}

		if resp.AuthenticationResult == nil {
			return errors.New("Authentication has failed.")
		}

		return nil
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
