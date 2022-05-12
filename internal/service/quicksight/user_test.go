package quicksight_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
)

func TestAccQuickSightUser_basic(t *testing.T) {
	var user quicksight.User
	rName1 := "tfacctest" + sdkacctest.RandString(10)
	resourceName1 := "aws_quicksight_user." + rName1
	rName2 := "tfacctest" + sdkacctest.RandString(10)
	resourceName2 := "aws_quicksight_user." + rName2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName1, &user),
					resource.TestCheckResourceAttr(resourceName1, "user_name", rName1),
					acctest.CheckResourceAttrRegionalARN(resourceName1, "arn", "quicksight", fmt.Sprintf("user/default/%s", rName1)),
				),
			},
			{
				Config: testAccUserConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName2, &user),
					resource.TestCheckResourceAttr(resourceName2, "user_name", rName2),
					acctest.CheckResourceAttrRegionalARN(resourceName2, "arn", "quicksight", fmt.Sprintf("user/default/%s", rName2)),
				),
			},
		},
	})
}

func TestAccQuickSightUser_withInvalidFormattedEmailStillWorks(t *testing.T) {
	var user quicksight.User
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user." + rName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWithEmailConfig(rName, "nottarealemailbutworks"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "email", "nottarealemailbutworks"),
				),
			},
			{
				Config: testAccUserWithEmailConfig(rName, "nottarealemailbutworks2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "email", "nottarealemailbutworks2"),
				),
			},
		},
	})
}

func TestAccQuickSightUser_withNamespace(t *testing.T) {
	key := "QUICKSIGHT_NAMESPACE"
	namespace := os.Getenv(key)
	if namespace == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var user quicksight.User
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user." + rName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWithNamespaceConfig(rName, namespace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "namespace", namespace),
				),
			},
		},
	})
}

func TestAccQuickSightUser_disappears(t *testing.T) {
	var user quicksight.User
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user." + rName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserDisappears(&user),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserExists(resourceName string, user *quicksight.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, namespace, userName, err := tfquicksight.UserParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn

		input := &quicksight.DescribeUserInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			UserName:     aws.String(userName),
		}

		output, err := conn.DescribeUser(input)

		if err != nil {
			return err
		}

		if output == nil || output.User == nil {
			return fmt.Errorf("QuickSight User (%s) not found", rs.Primary.ID)
		}

		*user = *output.User

		return nil
	}
}

func testAccCheckUserDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_quicksight_user" {
			continue
		}

		awsAccountID, namespace, userName, err := tfquicksight.UserParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeUser(&quicksight.DescribeUserInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			UserName:     aws.String(userName),
		})
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("QuickSight User '%s' was not deleted properly", rs.Primary.ID)
	}

	return nil
}

func testAccCheckUserDisappears(v *quicksight.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn

		arn, err := arn.Parse(aws.StringValue(v.Arn))
		if err != nil {
			return err
		}

		parts := strings.SplitN(arn.Resource, "/", 3)

		input := &quicksight.DeleteUserInput{
			AwsAccountId: aws.String(arn.AccountID),
			Namespace:    aws.String(parts[1]),
			UserName:     v.UserName,
		}

		if _, err := conn.DeleteUser(input); err != nil {
			return err
		}

		return nil
	}
}

func testAccUserWithEmailConfig(rName, email string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_user" %[1]q {
  aws_account_id = data.aws_caller_identity.current.account_id
  user_name      = %[1]q
  email          = %[2]q
  identity_type  = "QUICKSIGHT"
  user_role      = "READER"
}
`, rName, email)
}

func testAccUserWithNamespaceConfig(rName, namespace string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_user" %[1]q {
  aws_account_id = data.aws_caller_identity.current.account_id
  user_name      = %[1]q
  email          = %[2]q
  namespace      = %[3]q
  identity_type  = "QUICKSIGHT"
  user_role      = "READER"
}
`, rName, acctest.DefaultEmailAddress, namespace)
}

func testAccUserConfig(rName string) string {
	return testAccUserWithEmailConfig(rName, acctest.DefaultEmailAddress)
}
