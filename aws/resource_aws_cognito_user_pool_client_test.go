package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCognitoUserPoolClient_basic(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists("aws_cognito_user_pool_client.client"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "name", name),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.245201344", "ADMIN_NO_SRP_AUTH"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_allFields(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_allFields(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists("aws_cognito_user_pool_client.client"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "name", name),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.245201344", "ADMIN_NO_SRP_AUTH"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "generate_secret", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "read_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "read_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "write_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "write_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "refresh_token_validity", "300"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows.#", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows.2645166319", "code"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows.3465961881", "implicit"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows_user_pool_client", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.#", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.2517049750", "openid"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "callback_urls.#", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "callback_urls.0", "https://www.example.com/callback"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "callback_urls.1", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "default_redirect_uri", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "logout_urls.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "logout_urls.0", "https://www.example.com/login"),
				),
			},
		},
	})
}

func testAccCheckAWSCognitoUserPoolClientDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool_client" {
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

func testAccCheckAWSCognitoUserPoolClientExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool Client ID set")
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

func testAccAWSCognitoUserPoolClientConfig_basic(clientName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "client" {
  name = "%s"

  user_pool_id = "${aws_cognito_user_pool.pool.id}"
  explicit_auth_flows = [ "ADMIN_NO_SRP_AUTH" ]
}

resource "aws_cognito_user_pool" "pool" {
  name = "test-pool"
}
`, clientName)
}

func testAccAWSCognitoUserPoolClientConfig_allFields(clientName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "client" {
  name = "%s"

  user_pool_id = "${aws_cognito_user_pool.pool.id}"
  explicit_auth_flows = [ "ADMIN_NO_SRP_AUTH" ]

  generate_secret = "true"

  read_attributes = ["email"]
  write_attributes = ["email"]

  refresh_token_validity = 300

  allowed_oauth_flows = ["code", "implicit"]
  allowed_oauth_flows_user_pool_client = "true"
  allowed_oauth_scopes = ["openid", "email"]
  
  callback_urls = ["https://www.example.com/callback", "https://www.example.com/redirect"]
  default_redirect_uri = "https://www.example.com/redirect"
  logout_urls = ["https://www.example.com/login"]
}

resource "aws_cognito_user_pool" "pool" {
  name = "test-pool"
}
`, clientName)
}
