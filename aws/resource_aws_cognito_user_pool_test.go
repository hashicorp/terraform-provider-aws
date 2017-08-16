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
	name := acctest.RandString(5)

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

func TestAccAWSCognitoUserPool_withEmailVerificationMessage(t *testing.T) {
	name := acctest.RandString(5)
	subject := acctest.RandString(10)
	updatedSubject := acctest.RandString(10)
	message := fmt.Sprintf("%s {####}", acctest.RandString(10))
	upatedMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withEmailVerificationMessage(name, subject, message),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_subject", subject),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_message", message),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withEmailVerificationMessage(name, updatedSubject, upatedMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_subject", updatedSubject),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_message", upatedMessage),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withSmsVerificationMessage(t *testing.T) {
	name := acctest.RandString(5)
	authenticationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))
	updatedAuthenticationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))
	verificationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))
	upatedVerificationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withSmsVerificationMessage(name, authenticationMessage, verificationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_authentication_message", authenticationMessage),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_verification_message", verificationMessage),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withSmsVerificationMessage(name, updatedAuthenticationMessage, upatedVerificationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_authentication_message", updatedAuthenticationMessage),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_verification_message", upatedVerificationMessage),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withTags(t *testing.T) {
	name := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withTags(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "tags.Name", "Foo"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withTagsUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "tags.Name", "FooBar"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "tags.Project", "Terraform"),
				),
			},
		},
	})
}

//func TestAccAWSCognitoUserPool_attributes(t *testing.T) {
//	name := acctest.RandString(5)
//	subject := acctest.RandString(10)
//	message := fmt.Sprintf("%s {####}", acctest.RandString(10))
//	authenticationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))
//	verificationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		Providers:    testAccProviders,
//		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccAWSCognitoUserPoolConfig_attributes(name, authenticationMessage, verificationMessage, subject, message),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.#", "3"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.0", "email"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.1", "phone_number"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.2", "preferred_username"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.#", "2"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.0", "email"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.1", "phone_number"),
//				),
//			},
//			{
//				Config: testAccAWSCognitoUserPoolConfig_attributesUpdated(name, subject, message),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.#", "2"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.0", "email"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.1", "preferred_username"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.#", "1"),
//					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.0", "email"),
//				),
//			},
//		},
//	})
//}

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

func testAccAWSCognitoUserPoolConfig_withEmailVerificationMessage(name, subject, message string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  email_verification_subject = "%s"
  email_verification_message = "%s"
}`, name, subject, message)
}

func testAccAWSCognitoUserPoolConfig_withSmsVerificationMessage(name, authenticationMessage, verificationMessage string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  sms_authentication_message = "%s"
  sms_verification_message   = "%s"
}`, name, authenticationMessage, verificationMessage)
}

func testAccAWSCognitoUserPoolConfig_withTags(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  tags {
    "Name" = "Foo"
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withTagsUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  tags {
    "Name"    = "FooBar"
    "Project" = "Terraform"
  }
}`, name)
}

//func testAccAWSCognitoUserPoolConfig_attributes(name, authenticationMessage, verificationMessage, subject, message string) string {
//	return fmt.Sprintf(`
//	resource "aws_cognito_user_pool" "pool" {
//		name = "terraform-test-pool-%s"
//
//		alias_attributes         = ["preferred_username"]
//	}`, name, subject, message, authenticationMessage, verificationMessage)
//}
//
//func testAccAWSCognitoUserPoolConfig_attributesUpdated(name, subject, message string) string {
//	return fmt.Sprintf(`
//	resource "aws_cognito_user_pool" "pool" {
//		name = "terraform-test-pool-%s"
//
//		alias_attributes         = ["email", "preferred_username"]
//		auto_verified_attributes = ["email", "phone_number"]
//	}`, name, subject, message)
//}
