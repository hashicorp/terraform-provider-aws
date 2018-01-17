package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIAMUserPolicy_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserPolicyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(
						"aws_iam_user.user",
						"aws_iam_user_policy.foo",
					),
				),
			},
			{
				Config: testAccIAMUserPolicyConfigUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(
						"aws_iam_user.user",
						"aws_iam_user_policy.bar",
					),
				),
			},
		},
	})
}

func TestAccAWSIAMUserPolicy_namePrefix(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_iam_user_policy.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserPolicyConfig_namePrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(
						"aws_iam_user.test",
						"aws_iam_user_policy.test",
					),
				),
			},
		},
	})
}

func TestAccAWSIAMUserPolicy_generatedName(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_iam_user_policy.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserPolicyConfig_generatedName(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(
						"aws_iam_user.test",
						"aws_iam_user_policy.test",
					),
				),
			},
			{
				Config: testAccIAMUserPolicyConfig_generatedNameUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(
						"aws_iam_user.test",
						"aws_iam_user_policy.test",
					),
				),
			},
		},
	})
}

func TestAccAWSIAMUserPolicy_importBasic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserPolicyConfig(rInt),
			},
			{
				ResourceName:      "aws_iam_user_policy.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSIAMUserPolicy_invalidJSON(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccIAMUserPolicyConfig_invalidJSON(rInt),
				ExpectError: regexp.MustCompile("invalid JSON"),
			},
		},
	})
}

func testAccCheckIAMUserPolicyDestroy(s *terraform.State) error {
	iamconn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_user_policy" {
			continue
		}

		user, name, err := resourceAwsIamUserPolicyParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		request := &iam.GetUserPolicyInput{
			PolicyName: aws.String(name),
			UserName:   aws.String(user),
		}

		getResp, err := iamconn.GetUserPolicy(request)
		if err != nil {
			if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" {
				// none found, that's good
				return nil
			}
			return fmt.Errorf("Error reading IAM policy %s from user %s: %s", name, user, err)
		}

		if getResp != nil {
			return fmt.Errorf("Found IAM user policy, expected none: %s", getResp)
		}
	}

	return nil
}

func testAccCheckIAMUserPolicy(
	iamUserResource string,
	iamUserPolicyResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[iamUserResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamUserResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[iamUserPolicyResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamUserPolicyResource)
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn
		username, name, err := resourceAwsIamUserPolicyParseId(policy.Primary.ID)
		if err != nil {
			return err
		}

		_, err = iamconn.GetUserPolicy(&iam.GetUserPolicyInput{
			UserName:   aws.String(username),
			PolicyName: aws.String(name),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccIAMUserPolicyConfig(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_iam_user" "user" {
		name = "test_user_%d"
		path = "/"
	}

	resource "aws_iam_user_policy" "foo" {
		name = "foo_policy_%d"
		user = "${aws_iam_user.user.name}"
		policy = "{\"Version\":\"2012-10-17\",\"Statement\":{\"Effect\":\"Allow\",\"Action\":\"*\",\"Resource\":\"*\"}}"
	}`, rInt, rInt)
}

func testAccIAMUserPolicyConfig_invalidJSON(rInt int) string {
	return fmt.Sprintf(`
  resource "aws_iam_user" "user" {
    name = "test_user_%d"
    path = "/"
  }

  resource "aws_iam_user_policy" "foo" {
    name = "foo_policy_%d"
    user = "${aws_iam_user.user.name}"
    policy = "NonJSONString"
  }`, rInt, rInt)
}

func testAccIAMUserPolicyConfig_namePrefix(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_iam_user" "test" {
		name = "test_user_%d"
		path = "/"
	}

	resource "aws_iam_user_policy" "test" {
		name_prefix = "test-%d"
		user = "${aws_iam_user.test.name}"
		policy = "{\"Version\":\"2012-10-17\",\"Statement\":{\"Effect\":\"Allow\",\"Action\":\"*\",\"Resource\":\"*\"}}"
	}`, rInt, rInt)
}

func testAccIAMUserPolicyConfig_generatedName(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_iam_user" "test" {
		name = "test_user_%d"
		path = "/"
	}

	resource "aws_iam_user_policy" "test" {
		user = "${aws_iam_user.test.name}"
		policy = "{\"Version\":\"2012-10-17\",\"Statement\":{\"Effect\":\"Allow\",\"Action\":\"*\",\"Resource\":\"*\"}}"
	}`, rInt)
}

func testAccIAMUserPolicyConfig_generatedNameUpdate(rInt int) string {
	return fmt.Sprintf(`
  resource "aws_iam_user" "test" {
    name = "test_user_%d"
    path = "/"
  }

  resource "aws_iam_user_policy" "test" {
    user = "${aws_iam_user.test.name}"
    policy = "{\"Version\":\"2012-10-17\",\"Statement\":{\"Effect\":\"Allow\",\"Action\":\"iam:*\",\"Resource\":\"*\"}}"
  }`, rInt)
}

func testAccIAMUserPolicyConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_iam_user" "user" {
		name = "test_user_%d"
		path = "/"
	}

	resource "aws_iam_user_policy" "foo" {
		name = "foo_policy_%d"
		user = "${aws_iam_user.user.name}"
		policy = "{\"Version\":\"2012-10-17\",\"Statement\":{\"Effect\":\"Allow\",\"Action\":\"*\",\"Resource\":\"*\"}}"
	}

	resource "aws_iam_user_policy" "bar" {
		name = "bar_policy_%d"
		user = "${aws_iam_user.user.name}"
		policy = "{\"Version\":\"2012-10-17\",\"Statement\":{\"Effect\":\"Allow\",\"Action\":\"*\",\"Resource\":\"*\"}}"
	}`, rInt, rInt, rInt)
}
