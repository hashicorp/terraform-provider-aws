// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rolesanywhere_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfrolesanywhere "github.com/hashicorp/terraform-provider-aws/internal/service/rolesanywhere"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRolesAnywhereProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rolesanywhere_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_basic(rName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "role_arns.#", "1"),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, "role_arns.0", "iam", fmt.Sprintf("role/%s", roleName)),
					resource.TestCheckResourceAttr(resourceName, "duration_seconds", "3600"),
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

func TestAccRolesAnywhereProfile_noRoleARNs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rolesanywhere_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_noRoleARNs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "duration_seconds", "3600"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProfileConfig_basic(rName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "role_arns.#", "1"),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, "role_arns.0", "iam", fmt.Sprintf("role/%s", roleName)),
					resource.TestCheckResourceAttr(resourceName, "duration_seconds", "3600"),
				),
			},
		},
	})
}

func TestAccRolesAnywhereProfile_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rolesanywhere_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_tags1(rName, roleName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProfileConfig_tags2(rName, roleName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccProfileConfig_tags1(rName, roleName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRolesAnywhereProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rolesanywhere_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_basic(rName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrolesanywhere.ResourceProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRolesAnywhereProfile_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rolesanywhere_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_enabled(rName, roleName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProfileConfig_enabled(rName, roleName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				Config: testAccProfileConfig_enabled(rName, roleName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRolesAnywhereProfile_acceptRoleSessionName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rolesanywhere_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RolesAnywhereServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_acceptRoleSessionName(rName, roleName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_role_session_name", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProfileConfig_acceptRoleSessionName(rName, roleName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_role_session_name", acctest.CtFalse),
				),
			},
			{
				Config: testAccProfileConfig_acceptRoleSessionName(rName, roleName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_role_session_name", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RolesAnywhereClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rolesanywhere_profile" {
				continue
			}

			_, err := tfrolesanywhere.FindProfileByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RolesAnywhere Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProfileExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RolesAnywhereClient(ctx)

		_, err := tfrolesanywhere.FindProfileByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccProfileConfig_base(roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, roleName)
}

func testAccProfileConfig_basic(rName, roleName string) string {
	return acctest.ConfigCompose(
		testAccProfileConfig_base(roleName),
		fmt.Sprintf(`
resource "aws_rolesanywhere_profile" "test" {
  name      = %[1]q
  role_arns = [aws_iam_role.test.arn]
}
`, rName))
}

func testAccProfileConfig_noRoleARNs(rName string) string {
	return fmt.Sprintf(`
resource "aws_rolesanywhere_profile" "test" {
  name = %[1]q
}
`, rName)
}

func testAccProfileConfig_tags1(rName, roleName, tag, value string) string {
	return acctest.ConfigCompose(
		testAccProfileConfig_base(roleName),
		fmt.Sprintf(`
resource "aws_rolesanywhere_profile" "test" {
  name      = %[1]q
  role_arns = [aws_iam_role.test.arn]
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag, value))
}

func testAccProfileConfig_tags2(rName, roleName, tag1, value1, tag2, value2 string) string {
	return acctest.ConfigCompose(
		testAccProfileConfig_base(roleName),
		fmt.Sprintf(`
resource "aws_rolesanywhere_profile" "test" {
  name      = %[1]q
  role_arns = [aws_iam_role.test.arn]
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1, value1, tag2, value2))
}

func testAccProfileConfig_enabled(rName, roleName string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccProfileConfig_base(roleName),
		fmt.Sprintf(`
resource "aws_rolesanywhere_profile" "test" {
  name      = %[1]q
  role_arns = [aws_iam_role.test.arn]
  enabled   = %[2]t
}
`, rName, enabled))
}

func testAccProfileConfig_acceptRoleSessionName(rName, roleName string, acceptRoleSessionName bool) string {
	return acctest.ConfigCompose(
		testAccProfileConfig_base(roleName),
		fmt.Sprintf(`
resource "aws_rolesanywhere_profile" "test" {
  name                     = %[1]q
  role_arns                = [aws_iam_role.test.arn]
  accept_role_session_name = %[2]t
}
`, rName, acceptRoleSessionName))
}
