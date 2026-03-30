// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var userOutput awstypes.User
	resourceName := "aws_appstream_user.test"
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, resourceName, &userOutput),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, resourceName, &userOutput),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappstream.ResourceUser(), resourceName),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_complete(authType, rEmail, firstName, lastName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, resourceName, &userOutput),
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
					testAccCheckUserExists(ctx, t, resourceName, &userOutput),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, rEmail),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				Config: testAccUserConfig_complete(authType, rEmail, firstName, lastName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, resourceName, &userOutput),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, rEmail),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_user" {
				continue
			}

			_, err := tfappstream.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserName], awstypes.AuthenticationType(rs.Primary.Attributes["authentication_type"]))

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppStream User %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserExists(ctx context.Context, t *testing.T, n string, v *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppStreamClient(ctx)

		output, err := tfappstream.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserName], awstypes.AuthenticationType(rs.Primary.Attributes["authentication_type"]))

		if err != nil {
			return err
		}

		*v = *output

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
