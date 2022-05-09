package cognitoidp_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
)

func TestAccCognitoIDPUser_basic(t *testing.T) {
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigBasic(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
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
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigBasic(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, cognitoidp.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPUser_temporaryPassword(t *testing.T) {
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rClientName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserPassword := sdkacctest.RandString(16)
	rUserPasswordUpdated := sdkacctest.RandString(16)
	userResourceName := "aws_cognito_user.test"
	clientResourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigTemporaryPassword(rUserPoolName, rClientName, rUserName, rUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(userResourceName),
					testAccUserTemporaryPassword(userResourceName, clientResourceName),
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
				Config: testAccUserConfigTemporaryPassword(rUserPoolName, rClientName, rUserName, rUserPasswordUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(userResourceName),
					testAccUserTemporaryPassword(userResourceName, clientResourceName),
					resource.TestCheckResourceAttr(userResourceName, "status", cognitoidentityprovider.UserStatusTypeForceChangePassword),
				),
			},
			{
				Config: testAccUserConfigNoPassword(rUserPoolName, rClientName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(userResourceName),
					resource.TestCheckResourceAttr(userResourceName, "temporary_password", ""),
					resource.TestCheckResourceAttr(userResourceName, "status", cognitoidentityprovider.UserStatusTypeForceChangePassword),
				),
			},
		},
	})
}

func TestAccCognitoIDPUser_password(t *testing.T) {
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rClientName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserPassword := sdkacctest.RandString(16)
	rUserPasswordUpdated := sdkacctest.RandString(16)
	userResourceName := "aws_cognito_user.test"
	clientResourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigPassword(rUserPoolName, rClientName, rUserName, rUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(userResourceName),
					testAccUserPassword(userResourceName, clientResourceName),
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
				Config: testAccUserConfigPassword(rUserPoolName, rClientName, rUserName, rUserPasswordUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(userResourceName),
					testAccUserPassword(userResourceName, clientResourceName),
					resource.TestCheckResourceAttr(userResourceName, "status", cognitoidentityprovider.UserStatusTypeConfirmed),
				),
			},
			{
				Config: testAccUserConfigNoPassword(rUserPoolName, rClientName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(userResourceName),
					resource.TestCheckResourceAttr(userResourceName, "password", ""),
					resource.TestCheckResourceAttr(userResourceName, "status", cognitoidentityprovider.UserStatusTypeConfirmed),
				),
			},
		},
	})
}

func TestAccCognitoIDPUser_attributes(t *testing.T) {
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigAttributes(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
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
				Config: testAccUserConfigAttributesUpdated(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
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
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigEnable(rUserPoolName, rUserName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
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
				Config: testAccUserConfigEnable(rUserPoolName, rUserName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckUserExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		id := rs.Primary.ID
		userName := rs.Primary.Attributes["username"]
		userPoolId := rs.Primary.Attributes["user_pool_id"]

		if userName == "" {
			return errors.New("No Cognito User Name set")
		}

		if userPoolId == "" {
			return errors.New("No Cognito User Pool Id set")
		}

		if id != fmt.Sprintf("%s/%s", userPoolId, userName) {
			return fmt.Errorf(fmt.Sprintf("ID should be user_pool_id/name. ID was %s. name was %s, user_pool_id was %s", id, userName, userPoolId))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

		params := &cognitoidentityprovider.AdminGetUserInput{
			Username:   aws.String(rs.Primary.Attributes["username"]),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.AdminGetUser(params)
		return err
	}
}

func testAccCheckUserDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user" {
			continue
		}

		params := &cognitoidentityprovider.AdminGetUserInput{
			Username:   aws.String(rs.Primary.Attributes["username"]),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.AdminGetUser(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && (awsErr.Code() == cognitoidentityprovider.ErrCodeUserNotFoundException || awsErr.Code() == cognitoidentityprovider.ErrCodeResourceNotFoundException) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccUserTemporaryPassword(userResName string, clientResName string) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

		params := &cognitoidentityprovider.InitiateAuthInput{
			AuthFlow: aws.String(cognitoidentityprovider.AuthFlowTypeUserPasswordAuth),
			AuthParameters: map[string]*string{
				"USERNAME": aws.String(userName),
				"PASSWORD": aws.String(userPassword),
			},
			ClientId: aws.String(clientId),
		}

		resp, err := conn.InitiateAuth(params)
		if err != nil {
			return err
		}

		if aws.StringValue(resp.ChallengeName) != cognitoidentityprovider.ChallengeNameTypeNewPasswordRequired {
			return errors.New("The password is not a temporary password.")
		}

		return nil
	}
}

func testAccUserPassword(userResName string, clientResName string) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

		params := &cognitoidentityprovider.InitiateAuthInput{
			AuthFlow: aws.String(cognitoidentityprovider.AuthFlowTypeUserPasswordAuth),
			AuthParameters: map[string]*string{
				"USERNAME": aws.String(userName),
				"PASSWORD": aws.String(userPassword),
			},
			ClientId: aws.String(clientId),
		}

		resp, err := conn.InitiateAuth(params)
		if err != nil {
			return err
		}

		if resp.AuthenticationResult == nil {
			return errors.New("Authentication has failed.")
		}

		return nil
	}
}

func testAccUserConfigBasic(userPoolName string, userName string) string {
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

func testAccUserConfigTemporaryPassword(userPoolName string, clientName string, userName string, password string) string {
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

func testAccUserConfigPassword(userPoolName string, clientName string, userName string, password string) string {
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

func testAccUserConfigNoPassword(userPoolName string, clientName string, userName string) string {
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

func testAccUserConfigAttributes(userPoolName string, userName string) string {
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

func testAccUserConfigAttributesUpdated(userPoolName string, userName string) string {
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

func testAccUserConfigEnable(userPoolName string, userName string, enabled bool) string {
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
