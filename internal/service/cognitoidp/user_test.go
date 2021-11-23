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

func TestAccCognitoUser_basic(t *testing.T) {
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigBasic(rUserPoolName, rUserName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", cognitoidentityprovider.UserStatusTypeForceChangePassword),
					resource.TestCheckResourceAttr(resourceName, "mfa_preference.0.sms_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "mfa_preference.0.software_token_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "mfa_preference.0.preferred_mfa", ""),
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

func TestAccCognitoUser_disappears(t *testing.T) {
	rUserPoolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
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
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccUserConfigBasic(userPoolName string, userName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
	name = %[1]q
}

resource "aws_cognito_user" "test" {
	user_pool_id = aws_cognito_user_pool.test.id
	username = %[2]q
}
`, userPoolName, userName)
}
