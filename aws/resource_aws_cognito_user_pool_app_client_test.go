package aws

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestAccAWSCognitoUserPoolAppClient_basic(t *testing.T) {
	name := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolAppClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolAppClientConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolAppClientExists("aws_cognito_user_pool_app_client.basic"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.basic", "name", "terraform-test-pool-app-client-"+name),
					resource.TestCheckNoResourceAttr("aws_cognito_user_pool_app_client.basic", "client_secret"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolAppClient_generate_secret(t *testing.T) {
	name := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolAppClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolAppClientConfig_generate_secret(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolAppClientExists("aws_cognito_user_pool_app_client.secret"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.secret", "name", "terraform-test-pool-app-client-"+name),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.secret", "generate_secret", "true"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_app_client.secret", "client_secret"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolAppClient_complex(t *testing.T) {
	name := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolAppClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolAppClientConfig_complex(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolAppClientExists("aws_cognito_user_pool_app_client.complex"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "name", "terraform-test-pool-app-client-"+name),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "generate_secret", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "refresh_token_validity", "7"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.140932285", "email_verified"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.2318696674", "name"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.2135446866", "custom:foo"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.98075411", "custom:bar"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "write_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "write_attributes.2318696674", "name"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "write_attributes.98075411", "custom:bar"),
					resource.TestCheckNoResourceAttr("aws_cognito_user_pool_app_client.complex", "client_secret"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolAppClientConfig_complex_updated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolAppClientExists("aws_cognito_user_pool_app_client.complex"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "name", "terraform-test-pool-app-client-"+name),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "generate_secret", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "refresh_token_validity", "15"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.140932285", "email_verified"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.2318696674", "name"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.2090881135", "custom:foobar"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "write_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "write_attributes.2318696674", "name"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_app_client.complex", "write_attributes.2090881135", "custom:foobar"),
					resource.TestCheckNoResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.2135446866"),
					resource.TestCheckNoResourceAttr("aws_cognito_user_pool_app_client.complex", "read_attributes.98075411"),
					resource.TestCheckNoResourceAttr("aws_cognito_user_pool_app_client.complex", "write_attributes.98075411"),
					resource.TestCheckNoResourceAttr("aws_cognito_user_pool_app_client.complex", "client_secret"),
				),
			},
		},
	})
}

func testAccCheckAWSCognitoUserPoolAppClientDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool_app_client" {
			continue
		}

		params := &cognitoidentityprovider.DescribeUserPoolClientInput{
			ClientId:   aws.String(rs.Primary.ID),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.DescribeUserPoolClient(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSCognitoUserPoolAppClientExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool App Client ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		params := &cognitoidentityprovider.DescribeUserPoolClientInput{
			ClientId:   aws.String(rs.Primary.ID),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.DescribeUserPoolClient(params)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSCognitoUserPoolAppClientConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"
}

resource "aws_cognito_user_pool_app_client" "basic" {
  name         = "terraform-test-pool-app-client-%s"
  user_pool_id = "${aws_cognito_user_pool.pool.id}"
}`, name, name)
}

func testAccAWSCognitoUserPoolAppClientConfig_generate_secret(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"
}

resource "aws_cognito_user_pool_app_client" "secret" {
  name         = "terraform-test-pool-app-client-%s"
  user_pool_id = "${aws_cognito_user_pool.pool.id}"

  generate_secret = true
}`, name, name)
}

func testAccAWSCognitoUserPoolAppClientConfig_complex(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "foo"
    required                 = false
  }

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "bar"
    required                 = false
  }
}

resource "aws_cognito_user_pool_app_client" "complex" {
  name         = "terraform-test-pool-app-client-%s"
  user_pool_id = "${aws_cognito_user_pool.pool.id}"

  generate_secret        = false
  refresh_token_validity = 7

  read_attributes = [
    "email",
    "email_verified",
    "name",
    "custom:foo",
    "custom:bar",
  ]

  write_attributes = [
    "email",
    "name",
    "custom:bar",
  ]
}`, name, name)
}

func testAccAWSCognitoUserPoolAppClientConfig_complex_updated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "foo"
    required                 = false
  }

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "bar"
    required                 = false
  }

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "foobar"
    required                 = false
  }
}

resource "aws_cognito_user_pool_app_client" "complex" {
  name         = "terraform-test-pool-app-client-%s"
  user_pool_id = "${aws_cognito_user_pool.pool.id}"

  generate_secret        = false
  refresh_token_validity = 15

  read_attributes = [
    "email",
    "email_verified",
    "name",
    "custom:foobar",
  ]

  write_attributes = [
    "email",
    "name",
    "custom:foobar",
  ]
}`, name, name)
}
