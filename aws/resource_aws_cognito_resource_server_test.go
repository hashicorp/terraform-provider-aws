package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCognitoResourceServer_basic(t *testing.T) {
	identifier := fmt.Sprintf("tf-acc-test-resource-server-id-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	name := fmt.Sprintf("tf-acc-test-resource-server-name-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoResourceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoResourceServerConfig_basic(identifier, name, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoResourceServerExists("aws_cognito_resource_server.main"),
					resource.TestCheckResourceAttr("aws_cognito_resource_server.main", "identifier", identifier),
					resource.TestCheckResourceAttr("aws_cognito_resource_server.main", "name", name),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "name", poolName),
				),
			},
		},
	})
}

func TestAccAWSCognitoResourceServer_full(t *testing.T) {
	identifier := fmt.Sprintf("tf-acc-test-resource-server-id-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	name := fmt.Sprintf("tf-acc-test-resource-server-name-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoResourceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoResourceServerConfig_full(identifier, name, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoResourceServerExists("aws_cognito_resource_server.main"),
					resource.TestCheckResourceAttr("aws_cognito_resource_server.main", "identifier", identifier),
					resource.TestCheckResourceAttr("aws_cognito_resource_server.main", "name", name),
					resource.TestCheckResourceAttrSet("aws_cognito_resource_server.main", "scope_identifiers"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "name", poolName),
				),
			},
		},
	})
}

func testAccCheckAWSCognitoResourceServerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito Resource Server ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		_, err := conn.DescribeResourceServer(&cognitoidentityprovider.DescribeResourceServerInput{
			Identifier: aws.String(rs.Primary.ID),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAWSCognitoResourceServerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_resource_server" {
			continue
		}

		_, err := conn.DescribeResourceServer(&cognitoidentityprovider.DescribeResourceServerInput{
			Identifier: aws.String(rs.Primary.ID),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		})

		if err != nil {
			if isAWSErr(err, "ResourceNotFoundException", "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccAWSCognitoResourceServerConfig_basic(identifier string, name string, poolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_resource_server" "main" {
  identifier = "%s"
  name = "%s"
  user_pool_id = "${aws_cognito_user_pool.main.id}"
}

resource "aws_cognito_user_pool" "main" {
  name = "%s"
}
`, identifier, name, poolName)
}

func testAccAWSCognitoResourceServerConfig_full(identifier string, name string, poolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_resource_server" "main" {
  identifier = "%s"
  name = "%s"

  scope = {
	scope_name = "scope_1_name"
    scope_description = "scope_1_description"
  }

  scope = {
	scope_name = "scope_2_name"
    scope_description = "scope_2_description"
  }

  user_pool_id = "${aws_cognito_user_pool.main.id}"
}

resource "aws_cognito_user_pool" "main" {
  name = "%s"
}
`, identifier, name, poolName)
}
