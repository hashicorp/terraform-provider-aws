package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestValidateIamUserName(t *testing.T) {
	validNames := []string{
		"test-user",
		"test_user",
		"testuser123",
		"TestUser",
		"Test-User",
		"test.user",
		"test.123,user",
		"testuser@hashicorp",
		"test+user@hashicorp.com",
	}
	for _, v := range validNames {
		_, errors := validateAwsIamUserName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid IAM User name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"!",
		"/",
		" ",
		":",
		";",
		"test name",
		"/slash-at-the-beginning",
		"slash-at-the-end/",
	}
	for _, v := range invalidNames {
		_, errors := validateAwsIamUserName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid IAM User name", v)
		}
	}
}

func TestAccAWSUser_basic(t *testing.T) {
	var conf iam.GetUserOutput

	name1 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	path1 := "/"
	path2 := "/path2/"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfig(name1, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
					testAccCheckAWSUserAttributes(&conf, name1, "/"),
				),
			},
			{
				Config: testAccAWSUserConfig(name2, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
					testAccCheckAWSUserAttributes(&conf, name2, "/path2/"),
				),
			},
		},
	})
}

func TestAccAWSUser_disappears(t *testing.T) {
	var user iam.GetUserOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.user"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfig(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					testAccCheckAWSUserDisappears(&user),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSUser_nameChange(t *testing.T) {
	var conf iam.GetUserOutput

	name1 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	path := "/"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfig(name1, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
				),
			},
			{
				Config: testAccAWSUserConfig(name2, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
				),
			},
		},
	})
}

func TestAccAWSUser_pathChange(t *testing.T) {
	var conf iam.GetUserOutput

	name := fmt.Sprintf("test-user-%d", acctest.RandInt())
	path1 := "/"
	path2 := "/updated/"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfig(name, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
				),
			},
			{
				Config: testAccAWSUserConfig(name, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
				),
			},
		},
	})
}

func TestAccAWSUser_permissionsBoundary(t *testing.T) {
	var user iam.GetUserOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.user"

	permissionsBoundary1 := fmt.Sprintf("arn:%s:iam::aws:policy/AdministratorAccess", testAccGetPartition())
	permissionsBoundary2 := fmt.Sprintf("arn:%s:iam::aws:policy/ReadOnlyAccess", testAccGetPartition())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccAWSUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
				),
			},
			// Test update
			{
				Config: testAccAWSUserConfig_permissionsBoundary(rName, permissionsBoundary2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary2),
				),
			},
			// Test import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			// Test removal
			{
				Config: testAccAWSUserConfig(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
				),
			},
			// Test addition
			{
				Config: testAccAWSUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
				),
			},
		},
	})
}

func testAccCheckAWSUserDestroy(s *terraform.State) error {
	iamconn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_user" {
			continue
		}

		// Try to get user
		_, err := iamconn.GetUser(&iam.GetUserInput{
			UserName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("still exist.")
		}

		// Verify the error is what we want
		if !isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSUserExists(n string, res *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		resp, err := iamconn.GetUser(&iam.GetUserInput{
			UserName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckAWSUserAttributes(user *iam.GetUserOutput, name string, path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *user.User.UserName != name {
			return fmt.Errorf("Bad name: %s", *user.User.UserName)
		}

		if *user.User.Path != path {
			return fmt.Errorf("Bad path: %s", *user.User.Path)
		}

		return nil
	}
}

func testAccCheckAWSUserDisappears(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		userName := aws.StringValue(getUserOutput.User.UserName)

		_, err := iamconn.DeleteUser(&iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
		if err != nil {
			return fmt.Errorf("error deleting user %q: %s", userName, err)
		}

		return nil
	}
}

func testAccAWSUserConfig(rName, path string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name = %q
  path = %q
}
`, rName, path)
}

func testAccAWSUserConfig_permissionsBoundary(rName, permissionsBoundary string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name                 = %q
  permissions_boundary = %q
}
`, rName, permissionsBoundary)
}
