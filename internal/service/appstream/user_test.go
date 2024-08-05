// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var userOutput awstypes.User
	resourceName := "aws_appstream_user.test"
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &userOutput),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, rEmail),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"send_email_notification"},
			},
		},
	})
}

func TestAccAppStreamUser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var userOutput awstypes.User
	resourceName := "aws_appstream_user.test"
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &userOutput),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamUser_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var userOutput awstypes.User
	resourceName := "aws_appstream_user.test"
	authType := "USERPOOL"
	firstName := "John"
	lastName := "Doe"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_complete(authType, rEmail, firstName, lastName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &userOutput),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, rEmail),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"send_email_notification"},
			},
			{
				Config: testAccUserConfig_complete(authType, rEmail, firstName, lastName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &userOutput),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, rEmail),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				Config: testAccUserConfig_complete(authType, rEmail, firstName, lastName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &userOutput),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, rEmail),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckUserExists(ctx context.Context, resourceName string, appStreamUser *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		userName, authType, err := tfappstream.DecodeUserID(rs.Primary.ID)
		if err != nil {
			return err
		}

		user, err := tfappstream.FindUserByTwoPartKey(ctx, conn, userName, authType)
		if tfresource.NotFound(err) {
			return fmt.Errorf("AppStream User %q does not exist", rs.Primary.ID)
		}
		if err != nil {
			return err
		}

		*appStreamUser = *user

		return nil
	}
}

func testAccCheckUserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_user" {
				continue
			}

			userName, authType, err := tfappstream.DecodeUserID(rs.Primary.ID)
			if err != nil {
				return err
			}

			resp, err := conn.DescribeUsers(ctx, &appstream.DescribeUsersInput{AuthenticationType: awstypes.AuthenticationType(authType)})

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			found := false

			for _, out := range resp.Users {
				if aws.ToString(out.UserName) == userName {
					found = true
				}
			}

			if found {
				return fmt.Errorf("AppStream User %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccUserConfig_basic(authType, userName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_user" "test" {
  authentication_type = %[1]q
  user_name           = %[2]q
}
`, authType, userName)
}

func testAccUserConfig_complete(authType, userName, firstName, lastName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_appstream_user" "test" {
  authentication_type = %[1]q
  user_name           = %[2]q
  first_name          = %[3]q
  last_name           = %[4]q
  enabled             = %[5]t
}
`, authType, userName, firstName, lastName, enabled)
}
