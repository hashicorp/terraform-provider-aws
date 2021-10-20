package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
)

func TestAccAppStreamUser_basic(t *testing.T) {
	var userOutput appstream.User
	resourceName := "aws_appstream_user.test"
	authType := "USERPOOL"
	rEmail := acctest.RandomEmailAddress("hashicorp.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &userOutput),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, "user_name", rEmail),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppStreamUser_disappears(t *testing.T) {
	var userOutput appstream.User
	resourceName := "aws_appstream_user.test"
	authType := "USERPOOL"
	rEmail := acctest.RandomEmailAddress("hashicorp.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &userOutput),
					acctest.CheckResourceDisappears(acctest.Provider, tfappstream.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamUser_complete(t *testing.T) {
	var userOutput appstream.User
	resourceName := "aws_appstream_user.test"
	authType := "USERPOOL"
	firstName := "John"
	lastName := "Doe"
	rEmail := acctest.RandomEmailAddress("hashicorp.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserCompleteConfig(authType, rEmail, firstName, lastName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &userOutput),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, "user_name", rEmail),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
				),
			},
			{
				Config: testAccUserCompleteSupressConfig(authType, rEmail, firstName, lastName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &userOutput),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, "user_name", rEmail),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"message_action"},
			},
		},
	})
}

func testAccCheckUserExists(resourceName string, appStreamUser *appstream.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

		userName, authType, err := tfappstream.DecodeUserID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.DescribeUsersWithContext(context.Background(), &appstream.DescribeUsersInput{AuthenticationType: aws.String(authType)})
		if err != nil {
			return err
		}

		if resp == nil && len(resp.Users) == 0 {
			return fmt.Errorf("appstream user %q does not exist", rs.Primary.ID)
		}

		var user *appstream.User

		for _, out := range resp.Users {
			if aws.StringValue(out.UserName) == userName {
				user = out
			}
		}

		if user == nil {
			return fmt.Errorf("appstream user %q does not exist", rs.Primary.ID)
		}

		appStreamUser = user

		return nil
	}
}

func testAccCheckUserDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_user" {
			continue
		}

		userName, authType, err := tfappstream.DecodeUserID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.DescribeUsersWithContext(context.Background(), &appstream.DescribeUsersInput{AuthenticationType: aws.String(authType)})

		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		found := false

		for _, out := range resp.Users {
			if aws.StringValue(out.UserName) == userName {
				found = true
			}
		}

		if found {
			return fmt.Errorf("appstream user %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccUserConfig(authType, userName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_user" "test" {
  authentication_type = %[1]q
  user_name           = %[2]q
}
`, authType, userName)
}

func testAccUserCompleteConfig(authType, userName, firstName, lastName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_user" "test" {
  authentication_type = %[1]q
  user_name           = %[2]q
  first_name          = %[3]q
  last_name           = %[4]q
}
`, authType, userName, firstName, lastName)
}

func testAccUserCompleteSupressConfig(authType, userName, firstName, lastName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_user" "test" {
  authentication_type = %[1]q
  user_name           = %[2]q
  first_name          = %[3]q
  last_name           = %[4]q
  message_action      = "SUPPRESS"
}
`, authType, userName, firstName, lastName)
}
