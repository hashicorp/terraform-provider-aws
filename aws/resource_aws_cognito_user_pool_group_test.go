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

func TestAccAWSCognitoUserPoolGroup_basic(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolGroupConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolGroupExists("aws_cognito_user_pool_group.group"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_group.group", "name", name),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolGroup_allFields(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolGroupConfig_allFields(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolGroupExists("aws_cognito_user_pool_client.client"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "name", name),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "description", "test_description"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "precedence", "8"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_client.client", "role_arn"),
				),
			},
		},
	})
}

func testAccCheckAWSCognitoUserPoolGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool_group" {
			continue
		}

		params := &cognitoidentityprovider.GetGroupInput{
			GroupName:  aws.String(rs.Primary.Attributes["name"]),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.GetGroup(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSCognitoUserPoolGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.Attributes["name"] == "" {
			return errors.New("No Cognito User Pool Group name set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		params := &cognitoidentityprovider.GetGroupInput{
			GroupName:  aws.String(rs.Primary.Attributes["name"]),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.GetGroup(params)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSCognitoUserPoolGroupConfig_basic(groupName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool_group" "group" {
  name = "%s"

  user_pool_id = "${aws_cognito_user_pool.pool.id}"
}

resource "aws_cognito_user_pool" "pool" {
  name = "test-pool"
}
`, groupName)
}

func testAccAWSCognitoUserPoolGroupConfig_allFields(groupName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool_group" "group" {
  name = "%s"

  user_pool_id = "${aws_cognito_user_pool.pool.id}"

  description = "test_description"
  precedence = 8
  role_arn = "${aws_iam_role.authenticated.arn}"
}

resource "aws_cognito_user_pool" "pool" {
  name = "test-pool"
}

# Authenticated Role
resource "aws_iam_role" "authenticated" {
  name = "cognito_authenticated"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "cognito-identity.amazonaws.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "authenticated" {
  name = "authenticated_policy"
  role = "${aws_iam_role.authenticated.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cognito-sync:*",
        "cognito-identity:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, groupName)
}
