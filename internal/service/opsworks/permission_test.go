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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpsWorksPermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_permission.test"
	var opsperm opsworks.Permission

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_create(rName, true, true, "iam_only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &opsperm),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "level", "iam_only"),
				),
			},
			{
				Config: testAccPermissionConfig_create(rName, true, false, "iam_only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &opsperm),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "level", "iam_only"),
				),
			},
			{
				Config: testAccPermissionConfig_create(rName, false, false, "deny"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &opsperm),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "level", "deny"),
				),
			},
			{
				Config: testAccPermissionConfig_create(rName, false, false, "show"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &opsperm),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "level", "show"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/4804
func TestAccOpsWorksPermission_self(t *testing.T) {
	ctx := acctest.Context(t)
	var opsperm opsworks.Permission
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_self(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &opsperm),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", acctest.CtTrue),
				),
			},
			{
				Config: testAccPermissionConfig_self(rName, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &opsperm),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckPermissionExists(ctx context.Context, n string, v *opsworks.Permission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No OpsWorks Layer ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn(ctx)

		output, err := tfopsworks.FindPermissionByTwoPartKey(ctx, conn, rs.Primary.Attributes["user_arn"], rs.Primary.Attributes["stack_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPermissionConfig_create(rName string, allowSSH, allowSudo bool, level string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		fmt.Sprintf(`
resource "aws_opsworks_permission" "test" {
  stack_id = aws_opsworks_stack.test.id

  allow_ssh  = %[1]t
  allow_sudo = %[2]t
  user_arn   = aws_opsworks_user_profile.user.user_arn
  level      = %[3]q
}

resource "aws_opsworks_user_profile" "user" {
  user_arn     = aws_iam_user.user.arn
  ssh_username = aws_iam_user.user.name
}

resource "aws_iam_user" "user" {
  name = %[4]q
  path = "/"
}
`, allowSSH, allowSudo, level, rName))
}

func testAccPermissionConfig_self(rName string, allowSSH bool, allowSudo bool) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_opsworks_permission" "test" {
  allow_ssh  = %[1]t
  allow_sudo = %[2]t
  stack_id   = aws_opsworks_stack.test.id
  user_arn   = data.aws_caller_identity.current.arn
}
`, allowSSH, allowSudo))
}
