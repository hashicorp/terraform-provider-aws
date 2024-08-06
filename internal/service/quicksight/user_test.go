// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var user quicksight.User
	rName1 := "tfacctest" + sdkacctest.RandString(10)
	resourceName1 := "aws_quicksight_user." + rName1
	rName2 := "tfacctest" + sdkacctest.RandString(10)
	resourceName2 := "aws_quicksight_user." + rName2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName1, &user),
					resource.TestCheckResourceAttr(resourceName1, names.AttrUserName, rName1),
					acctest.CheckResourceAttrRegionalARN(resourceName1, names.AttrARN, "quicksight", fmt.Sprintf("user/default/%s", rName1)),
				),
			},
			{
				Config: testAccUserConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName2, &user),
					resource.TestCheckResourceAttr(resourceName2, names.AttrUserName, rName2),
					acctest.CheckResourceAttrRegionalARN(resourceName2, names.AttrARN, "quicksight", fmt.Sprintf("user/default/%s", rName2)),
				),
			},
		},
	})
}

func TestAccQuickSightUser_withInvalidFormattedEmailStillWorks(t *testing.T) {
	ctx := acctest.Context(t)
	var user quicksight.User
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user." + rName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_email(rName, "nottarealemailbutworks"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, "nottarealemailbutworks"),
				),
			},
			{
				Config: testAccUserConfig_email(rName, "nottarealemailbutworks2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, "nottarealemailbutworks2"),
				),
			},
		},
	})
}

func TestAccQuickSightUser_withNamespace(t *testing.T) {
	ctx := acctest.Context(t)
	key := "QUICKSIGHT_NAMESPACE"
	namespace := os.Getenv(key)
	if namespace == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var user quicksight.User
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user." + rName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_namespace(rName, namespace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, namespace),
				),
			},
		},
	})
}

func TestAccQuickSightUser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var user quicksight.User
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user." + rName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					testAccCheckUserDisappears(ctx, &user),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserExists(ctx context.Context, resourceName string, user *quicksight.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, namespace, userName, err := tfquicksight.UserParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)

		input := &quicksight.DescribeUserInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			UserName:     aws.String(userName),
		}

		output, err := conn.DescribeUserWithContext(ctx, input)

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

func testAccCheckUserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_user" {
				continue
			}

			awsAccountID, namespace, userName, err := tfquicksight.UserParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = conn.DescribeUserWithContext(ctx, &quicksight.DescribeUserInput{
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
}

func testAccCheckUserDisappears(ctx context.Context, v *quicksight.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)

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

		if _, err := conn.DeleteUserWithContext(ctx, input); err != nil {
			return err
		}

		return nil
	}
}

func testAccUserConfig_email(rName, email string) string {
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

func testAccUserConfig_namespace(rName, namespace string) string {
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

func testAccUserConfig_basic(rName string) string {
	return testAccUserConfig_email(rName, acctest.DefaultEmailAddress)
}
