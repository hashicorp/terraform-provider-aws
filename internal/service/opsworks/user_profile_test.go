// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopsworks "github.com/hashicorp/terraform-provider-aws/internal/service/opsworks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpsWorksUserProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_create(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ssh_public_key", ""),
					resource.TestCheckResourceAttr(resourceName, "ssh_username", rName1),
					resource.TestCheckResourceAttr(resourceName, "allow_self_management", acctest.CtFalse),
				),
			},
			{
				Config: testAccUserProfileConfig_update(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ssh_public_key", ""),
					resource.TestCheckResourceAttr(resourceName, "ssh_username", rName2),
					resource.TestCheckResourceAttr(resourceName, "allow_self_management", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccOpsWorksUserProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfopsworks.ResourceUserProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserProfileExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No OpsWorks User Profile ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn(ctx)

		_, err := tfopsworks.FindUserProfileByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckUserProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opsworks_user_profile" {
				continue
			}

			_, err := tfopsworks.FindUserProfileByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpsWorks User Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccUserProfileConfig_create(rName string) string {
	return fmt.Sprintf(`
resource "aws_opsworks_user_profile" "test" {
  user_arn     = aws_iam_user.test1.arn
  ssh_username = aws_iam_user.test1.name
}

resource "aws_iam_user" "test1" {
  name = %[1]q
  path = "/"
}
`, rName)
}

func testAccUserProfileConfig_update(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_opsworks_user_profile" "test" {
  user_arn     = aws_iam_user.test2.arn
  ssh_username = aws_iam_user.test2.name
}

resource "aws_iam_user" "test1" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_user" "test2" {
  name = %[2]q
  path = "/"
}
`, rName1, rName2)
}
