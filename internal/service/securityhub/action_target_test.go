// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccActionTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActionTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActionTargetConfig_identifier("testaction"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "securityhub", "action/custom/testaction"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is a test custom action"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, "testaction"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "Test action"),
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

func testAccActionTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActionTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActionTargetConfig_identifier("testaction"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecurityhub.ResourceActionTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccActionTarget_Description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActionTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActionTargetConfig_description("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccActionTargetConfig_description("description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func testAccActionTarget_Name(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActionTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActionTargetConfig_name("name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccActionTargetConfig_name("name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "name2"),
				),
			},
		},
	})
}

func testAccCheckActionTargetExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		_, err := tfsecurityhub.FindActionTargetByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckActionTargetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_action_target" {
				continue
			}

			_, err := tfsecurityhub.FindActionTargetByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Action Target (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccActionTargetConfig_description(description string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_action_target" "test" {
  depends_on  = [aws_securityhub_account.test]
  description = %[1]q
  identifier  = "testaction"
  name        = "Test action"
}
`, description)
}

func testAccActionTargetConfig_identifier(identifier string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_action_target" "test" {
  depends_on  = [aws_securityhub_account.test]
  description = "This is a test custom action"
  identifier  = %[1]q
  name        = "Test action"
}
`, identifier)
}

func testAccActionTargetConfig_name(name string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_action_target" "test" {
  depends_on  = [aws_securityhub_account.test]
  description = "This is a test custom action"
  identifier  = "testaction"
  name        = %[1]q
}
`, name)
}
