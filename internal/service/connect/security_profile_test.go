// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSecurityProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SecurityProfile
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileConfig_basic(rName, rName2, "Created"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityProfileExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/security-profile/{security_profile_id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Created"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSecurityProfileConfig_basic(rName, rName2, "Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityProfileExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/security-profile/{security_profile_id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
				),
			},
		},
	})
}

func testAccSecurityProfile_updatePermissions(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SecurityProfile
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileConfig_basic(rName, rName2, "TestPermissionsUpdate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityProfileExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/security-profile/{security_profile_id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "TestPermissionsUpdate"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test updating permissions
				Config: testAccSecurityProfileConfig_permissions(rName, rName2, "TestPermissionsUpdate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityProfileExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/security-profile/{security_profile_id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "TestPermissionsUpdate"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
				),
			},
		},
	})
}

func testAccSecurityProfile_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SecurityProfile
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileConfig_basic(rName, rName2, names.AttrTags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityProfileExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Security Profile"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSecurityProfileConfig_tags(rName, rName2, names.AttrTags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityProfileExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Security Profile"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccSecurityProfileConfig_tagsUpdated(rName, rName2, names.AttrTags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityProfileExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Security Profile"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func testAccSecurityProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SecurityProfile
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileConfig_basic(rName, rName2, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityProfileExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconnect.ResourceSecurityProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSecurityProfileExists(ctx context.Context, n string, v *awstypes.SecurityProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectClient(ctx)

		output, err := tfconnect.FindSecurityProfileByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["security_profile_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSecurityProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_security_profile" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectClient(ctx)

			_, err := tfconnect.FindSecurityProfileByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["security_profile_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Security Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSecurityProfileConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccSecurityProfileConfig_basic(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccSecurityProfileConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_security_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  tags = {
    "Name" = "Test Security Profile"
  }
}
`, rName2, label))
}

func testAccSecurityProfileConfig_permissions(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccSecurityProfileConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_security_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  permissions = [
    "BasicAgentAccess",
    "OutboundCallAccess",
  ]

  tags = {
    "Name" = "Test Security Profile"
  }
}
`, rName2, label))
}

func testAccSecurityProfileConfig_tags(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccSecurityProfileConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_security_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  tags = {
    "Name" = "Test Security Profile"
    "Key2" = "Value2a"
  }
}
`, rName2, label))
}

func testAccSecurityProfileConfig_tagsUpdated(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccSecurityProfileConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_security_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  tags = {
    "Name" = "Test Security Profile"
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName2, label))
}
