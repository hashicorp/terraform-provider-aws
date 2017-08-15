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

func TestAccAWSCognitoUserPool_basic(t *testing.T) {
	name := fmt.Sprintf("%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "name", name),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withVerificationMessage(t *testing.T) {
	subject := fmt.Sprintf("%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	message := fmt.Sprintf("%s {####}", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withVerificationMessage(subject, message),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_subject", subject),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_message", message),
				),
			},
		},
	})
}

func testAccCheckAWSCognitoUserPoolDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool" {
			continue
		}

		params := &cognitoidentityprovider.DescribeUserPoolInput{
			UserPoolId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeUserPool(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSCognitoUserPoolExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		params := &cognitoidentityprovider.DescribeUserPoolInput{
			UserPoolId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeUserPool(params)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSCognitoUserPoolConfig_basic(name string) string {
	return fmt.Sprintf(`
	resource "aws_cognito_user_pool" "pool" {
		name = "%s"
	}`, name)
}

func testAccAWSCognitoUserPoolConfig_withVerificationMessage(subject, message string) string {
	return fmt.Sprintf(`
	resource "aws_cognito_user_pool" "pool" {
		name = "terraform-test-pool"

		email_verification_subject = "%s"
		email_verification_message = "%s"
	}`, subject, message)
}
