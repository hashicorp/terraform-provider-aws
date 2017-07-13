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
	name := fmt.Sprintf("%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

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
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_client.client", "secret"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_noSecret(t *testing.T) {
	name := fmt.Sprintf("%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_noSecret(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists("aws_cognito_user_pool_client.client"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "name", name),
					resource.TestCheckNoResourceAttr("aws_cognito_user_pool_client.client", "secret"),
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
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool"]),
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
			return errors.New("No Cognito User Pool ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		params := &cognitoidentityprovider.DescribeUserPoolClientInput{
			ClientId:   aws.String(rs.Primary.ID),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool"]),
		}

		_, err := conn.DescribeUserPoolClient(params)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSCognitoUserPoolClientConfig_basic(name string) string {
	return fmt.Sprintf(`
	resource "aws_cognito_user_pool" "pool" {
		name = "pool"
	}

	resource "aws_cognito_user_pool_client" "client" {
		name = "%s",
		user_pool = "${aws_cognito_user_pool.pool.id}"
	}`, name)
}

func testAccAWSCognitoUserPoolClientConfig_noSecret(name string) string {
	return fmt.Sprintf(`
	resource "aws_cognito_user_pool" "pool" {
		name = "pool"
	}

	resource "aws_cognito_user_pool_client" "client" {
		name = "%s",
		user_pool = "${aws_cognito_user_pool.pool.id}"
		generate_secret = false
	}`, name)
}
