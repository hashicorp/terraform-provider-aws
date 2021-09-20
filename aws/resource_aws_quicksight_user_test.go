package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/quicksight"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccAWSQuickSightUser_basic(t *testing.T) {
	var user quicksight.User
	rName1 := "tfacctest" + sdkacctest.RandString(10)
	resourceName1 := "aws_quicksight_user." + rName1
	rName2 := "tfacctest" + sdkacctest.RandString(10)
	resourceName2 := "aws_quicksight_user." + rName2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, quicksight.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckQuickSightUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightUserConfig(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightUserExists(resourceName1, &user),
					resource.TestCheckResourceAttr(resourceName1, "user_name", rName1),
					acctest.CheckResourceAttrRegionalARN(resourceName1, "arn", "quicksight", fmt.Sprintf("user/default/%s", rName1)),
				),
			},
			{
				Config: testAccAWSQuickSightUserConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightUserExists(resourceName2, &user),
					resource.TestCheckResourceAttr(resourceName2, "user_name", rName2),
					acctest.CheckResourceAttrRegionalARN(resourceName2, "arn", "quicksight", fmt.Sprintf("user/default/%s", rName2)),
				),
			},
		},
	})
}

func TestAccAWSQuickSightUser_withInvalidFormattedEmailStillWorks(t *testing.T) {
	var user quicksight.User
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user." + rName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, quicksight.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckQuickSightUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightUserConfigWithEmail(rName, "nottarealemailbutworks"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "email", "nottarealemailbutworks"),
				),
			},
			{
				Config: testAccAWSQuickSightUserConfigWithEmail(rName, "nottarealemailbutworks2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "email", "nottarealemailbutworks2"),
				),
			},
		},
	})
}

func TestAccAWSQuickSightUser_disappears(t *testing.T) {
	var user quicksight.User
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user." + rName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, quicksight.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckQuickSightUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightUserConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightUserExists(resourceName, &user),
					testAccCheckQuickSightUserDisappears(&user),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQuickSightUserExists(resourceName string, user *quicksight.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, namespace, userName, err := resourceAwsQuickSightUserParseID(rs.Primary.ID)
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

func testAccCheckQuickSightUserDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_quicksight_user" {
			continue
		}

		awsAccountID, namespace, userName, err := resourceAwsQuickSightUserParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeUser(&quicksight.DescribeUserInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			UserName:     aws.String(userName),
		})
		if tfawserr.ErrMessageContains(err, quicksight.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("QuickSight User '%s' was not deleted properly", rs.Primary.ID)
	}

	return nil
}

func testAccCheckQuickSightUserDisappears(v *quicksight.User) resource.TestCheckFunc {
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

func testAccAWSQuickSightUserConfigWithEmail(rName, email string) string {
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

func testAccAWSQuickSightUserConfig(rName string) string {
	return testAccAWSQuickSightUserConfigWithEmail(rName, acctest.DefaultEmailAddress)
}
