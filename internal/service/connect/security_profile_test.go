// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSecurityProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeSecurityProfileOutput
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Created"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
		},
	})
}

func testAccSecurityProfile_updatePermissions(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeSecurityProfileOutput
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "TestPermissionsUpdate"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "TestPermissionsUpdate"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
		},
	})
}

func testAccSecurityProfile_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeSecurityProfileOutput
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
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Security Profile"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccSecurityProfileConfig_tagsUpdated(rName, rName2, names.AttrTags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityProfileExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
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
	var v connect.DescribeSecurityProfileOutput
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

func testAccCheckSecurityProfileExists(ctx context.Context, resourceName string, function *connect.DescribeSecurityProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Security Profile not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Security Profile ID not set")
		}
		instanceID, securityProfileID, err := tfconnect.SecurityProfileParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

		params := &connect.DescribeSecurityProfileInput{
			InstanceId:        aws.String(instanceID),
			SecurityProfileId: aws.String(securityProfileID),
		}

		getFunction, err := conn.DescribeSecurityProfileWithContext(ctx, params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckSecurityProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_security_profile" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

			instanceID, securityProfileID, err := tfconnect.SecurityProfileParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			params := &connect.DescribeSecurityProfileInput{
				InstanceId:        aws.String(instanceID),
				SecurityProfileId: aws.String(securityProfileID),
			}

			_, err = conn.DescribeSecurityProfileWithContext(ctx, params)

			if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}
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
