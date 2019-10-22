package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSCognitoUserPoolIdentityProvider_basic(t *testing.T) {
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolIdentityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolIdentityProvider_basic(userPoolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolIdentityProviderExists("aws_cognito_user_pool_identity_provider.basic"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_name", "Google"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_type", "Google"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.authorize_scopes", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.attributes_url_add_attributes", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.authorize_url", "https://accounts.google.com/o/oauth2/v2/auth"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.token_url", "https://www.googleapis.com/oauth2/v4/token"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.oidc_issuer", "https://accounts.google.com"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.client_id", "123456789012-a1b2c3d4f5g6h7i8j9k0l1m2n3o4p5q6.apps.googleusercontent.com"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.attributes_url", "https://people.googleapis.com/v1/people/me?personFields="),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.client_secret", "rAnDoMly_GeNeRaTeD_sEcReT"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.token_request_method", "POST"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "attribute_mapping.username", "sub"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolIdentityProvider_basicUpdated(userPoolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolIdentityProviderExists("aws_cognito_user_pool_identity_provider.basic"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_name", "Google"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_type", "Google"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.authorize_scopes", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.attributes_url_add_attributes", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.authorize_url", "https://accounts.google.com/o/oauth2/v2/auth"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.token_url", "https://www.googleapis.com/oauth2/v4/token"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.oidc_issuer", "https://accounts.google.com"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.client_id", "123456789012-updatedclientid.apps.googleusercontent.com"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.attributes_url", "https://people.googleapis.com/v1/people/me?personFields="),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.client_secret", "aDifferentRandomlyGeneratedSecret"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "provider_details.token_request_method", "POST"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "attribute_mapping.username", "sub"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_identity_provider.basic", "attribute_mapping.email", "email"),
				),
			},
			{
				ResourceName:      "aws_cognito_user_pool.pool",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSCognitoUserPoolIdentityProviderDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool_identity_provider" {
			continue
		}

		params := &cognitoidentityprovider.DescribeIdentityProviderInput{
			ProviderName: aws.String(rs.Primary.ID),
			UserPoolId:   aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.DescribeIdentityProvider(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSCognitoUserPoolIdentityProviderExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool Identity Provider ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		params := &cognitoidentityprovider.DescribeIdentityProviderInput{
			ProviderName: aws.String(rs.Primary.ID),
			UserPoolId:   aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.DescribeIdentityProvider(params)

		return err
	}
}

func testAccAWSCognitoUserPoolIdentityProvider_basic(userPoolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "%s"
}

resource "aws_cognito_user_pool_identity_provider" "basic" {
  provider_name    = "Google"
  user_pool_id     = "${aws_cognito_user_pool.pool.id}"
  provider_type    = "Google"

  provider_details = {
    authorize_scopes              = "email"
	attributes_url_add_attributes = "true"
	authorize_url                 = "https://accounts.google.com/o/oauth2/v2/auth"
	token_url                     = "https://www.googleapis.com/oauth2/v4/token"
	oidc_issuer                   = "https://accounts.google.com"
    client_id                     = "123456789012-a1b2c3d4f5g6h7i8j9k0l1m2n3o4p5q6.apps.googleusercontent.com"
    attributes_url                = "https://people.googleapis.com/v1/people/me?personFields="
    client_secret                 = "rAnDoMly_GeNeRaTeD_sEcReT"
    token_request_method          = "POST"
  }
}
`, userPoolName)
}

func testAccAWSCognitoUserPoolIdentityProvider_basicUpdated(userPoolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "%s"
}

resource "aws_cognito_user_pool_identity_provider" "basic" {
  provider_name    = "Google"
  user_pool_id     = "${aws_cognito_user_pool.pool.id}"
  provider_type    = "Google"

  provider_details = {
    authorize_scopes              = "email"
    attributes_url_add_attributes = "true"
    authorize_url                 = "https://accounts.google.com/o/oauth2/v2/auth"
    token_url                     = "https://www.googleapis.com/oauth2/v4/token"
    oidc_issuer                   = "https://accounts.google.com"
    client_id                     = "123456789012-updatedclientid.apps.googleusercontent.com"
    attributes_url                = "https://people.googleapis.com/v1/people/me?personFields="
    client_secret                 = "aDifferentRandomlyGeneratedSecret"
    token_request_method          = "POST"
  }

  attribute_mapping = {
    username = "sub"
    email    = "email"
  }
}
`, userPoolName)
}
