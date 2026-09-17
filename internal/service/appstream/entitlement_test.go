// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamEntitlement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var entitlementOutput awstypes.Entitlement
	resourceName := "aws_appstream_entitlement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntitlementExists(ctx, resourceName, &entitlementOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "app_visibility", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified_time"),
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

func TestAccAppStreamEntitlement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var entitlementOutput awstypes.Entitlement
	resourceName := "aws_appstream_entitlement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntitlementExists(ctx, resourceName, &entitlementOutput),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceEntitlement(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamEntitlement_update(t *testing.T) {
	ctx := acctest.Context(t)
	var entitlementOutput awstypes.Entitlement
	resourceName := "aws_appstream_entitlement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Test description"
	descriptionUpdated := "Updated test description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementConfig_description(rName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntitlementExists(ctx, resourceName, &entitlementOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "app_visibility", "ALL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEntitlementConfig_description(rName, descriptionUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntitlementExists(ctx, resourceName, &entitlementOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
				),
			},
		},
	})
}

func TestAccAppStreamEntitlement_attributes(t *testing.T) {
	ctx := acctest.Context(t)
	var entitlementOutput awstypes.Entitlement
	resourceName := "aws_appstream_entitlement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementConfig_attributes(rName, "roles", "engineering"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntitlementExists(ctx, resourceName, &entitlementOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attributes.*", map[string]string{
						names.AttrName:  "roles",
						names.AttrValue: "engineering",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEntitlementConfig_attributes(rName, "department", "development"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntitlementExists(ctx, resourceName, &entitlementOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attributes.*", map[string]string{
						names.AttrName:  "department",
						names.AttrValue: "development",
					}),
				),
			},
		},
	})
}

func testAccCheckEntitlementDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_entitlement" {
				continue
			}

			stackName := rs.Primary.Attributes["stack_name"]
			_, err := tfappstream.FindEntitlementByTwoPartKey(ctx, conn, stackName, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppStream Entitlement %s/%s still exists", stackName, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEntitlementExists(ctx context.Context, n string, v *awstypes.Entitlement) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		stackName := rs.Primary.Attributes["stack_name"]
		output, err := tfappstream.FindEntitlementByTwoPartKey(ctx, conn, stackName, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEntitlementConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q
}

resource "aws_appstream_entitlement" "test" {
  name           = %[1]q
  stack_name     = aws_appstream_stack.test.name
  app_visibility = "ALL"

  attributes {
    name  = "domain"
    value = "test.example.com"
  }
}
`, rName)
}

func testAccEntitlementConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q
}

resource "aws_appstream_entitlement" "test" {
  name           = %[1]q
  stack_name     = aws_appstream_stack.test.name
  description    = %[2]q
  app_visibility = "ALL"

  attributes {
    name  = "domain"
    value = "test.example.com"
  }
}
`, rName, description)
}

func testAccEntitlementConfig_attributes(rName, attrName, attrValue string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q
}

resource "aws_appstream_entitlement" "test" {
  name           = %[1]q
  stack_name     = aws_appstream_stack.test.name
  app_visibility = "ALL"

  attributes {
    name  = %[2]q
    value = %[3]q
  }
}
`, rName, attrName, attrValue)
}
