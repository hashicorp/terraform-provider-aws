package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCognitoIdentityProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityProviderConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_name", "Google"),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_type", "Google"),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_details.%", "9"),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_details.authorize_scopes", "email"),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_details.authorize_url", "https://accounts.google.com/o/oauth2/v2/auth"),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_details.client_id", "test-url.apps.googleusercontent.com"),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_details.client_secret", "client_secret"),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_details.attributes_url", "https://people.googleapis.com/v1/people/me?personFields="),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_details.attributes_url_add_attributes", "true"),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_details.token_request_method", "POST"),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_details.token_url", "https://www.googleapis.com/oauth2/v4/token"),
					resource.TestCheckResourceAttr("aws_cognito_identity_provider.tf_test_provider", "provider_details.oidc_issuer", "https://accounts.google.com"),
				),
			},
		},
	})
}

func testAccCheckAWSCognitoIdentityProviderDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_identity_provider" {
			continue
		}

		_, err := conn.DescribeIdentityProvider(&cognitoidentityprovider.DescribeIdentityProviderInput{
			ProviderName: aws.String(rs.Primary.ID),
			UserPoolId:   aws.String(rs.Primary.Attributes["user_pool_id"]),
		})

		if err != nil {
			if wserr, ok := err.(awserr.Error); ok && wserr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccAWSCognitoIdentityProviderConfig_basic() string {
	return `

resource "aws_cognito_user_pool" "tf_test_pool" {
  name 						= "tfmytestpool"
  auto_verified_attributes  = ["email"]
}

resource "aws_cognito_identity_provider" "tf_test_provider" {
  user_pool_id  	= "${aws_cognito_user_pool.tf_test_pool.id}"
  provider_name 	= "Google"
  provider_type 	= "Google"

  provider_details {
  	authorize_scopes 		= "email"
	client_id				= "test-url.apps.googleusercontent.com"
	client_secret			= "client_secret"
	attributes_url			= "https://people.googleapis.com/v1/people/me?personFields="
	attributes_url_add_attributes = "true"
	authorize_url           = "https://accounts.google.com/o/oauth2/v2/auth"
	oidc_issuer             = "https://accounts.google.com"
	token_request_method 	= "POST"
	token_url               = "https://www.googleapis.com/oauth2/v4/token"
  }

  attribute_mapping {
  	email 		= "email"
    username 	= "sub"
  }
}
`
}
