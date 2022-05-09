package cognitoidp_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
)

func TestAccCognitoUserInGroup_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_in_group.test"
	userPoolResourceName := "aws_cognito_user_pool.test"
	userGroupResourceName := "aws_cognito_user_group.test"
	userResourceName := "aws_cognito_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserInGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigUserInGroup_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserInGroupExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "group_name", userGroupResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "username", userResourceName, "username"),
				),
			},
		},
	})
}

func TestAccCognitoUserInGroup_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_in_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserInGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigUserInGroup_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserInGroupExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcognitoidp.ResourceUserInGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigUserInGroup_basic(rName string) string {
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

func testAccCheckUserInGroupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

		groupName := rs.Primary.Attributes["group_name"]
		userPoolId := rs.Primary.Attributes["user_pool_id"]
		username := rs.Primary.Attributes["username"]

		found, err := tfcognitoidp.FindCognitoUserInGroup(conn, groupName, userPoolId, username)

		if err != nil {
			return err
		}

		if !found {
			return errors.New("user in group not found")
		}

		return nil
	}
}

func testAccCheckUserInGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_in_group" {
			continue
		}

		groupName := rs.Primary.Attributes["group_name"]
		userPoolId := rs.Primary.Attributes["user_pool_id"]
		username := rs.Primary.Attributes["username"]

		found, err := tfcognitoidp.FindCognitoUserInGroup(conn, groupName, userPoolId, username)

		if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if found {
			return fmt.Errorf("user in group still exists (%s)", rs.Primary.ID)
		}
	}

	return nil
}
