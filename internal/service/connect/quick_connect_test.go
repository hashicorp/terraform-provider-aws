// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccQuickConnect_phoneNumber(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.QuickConnect
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	rName2 := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_connect_quick_connect.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickConnectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, "Created", "+12345678912"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/transfer-destination/{quick_connect_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.quick_connect_type", "PHONE_NUMBER"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.0.phone_number", "+12345678912"),

					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// update description
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, "Updated", "+12345678912"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/transfer-destination/{quick_connect_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.quick_connect_type", "PHONE_NUMBER"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.0.phone_number", "+12345678912"),

					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// update phone number
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, "Updated", "+12345678913"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/transfer-destination/{quick_connect_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.quick_connect_type", "PHONE_NUMBER"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.0.phone_number", "+12345678913"),

					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
				),
			},
		},
	})
}

func testAccQuickConnect_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.QuickConnect
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	rName2 := acctest.RandomWithPrefix(t, "resource-test-terraform")
	phone_number := "+12345678912"

	resourceName := "aws_connect_quick_connect.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickConnectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, names.AttrTags, phone_number),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Quick Connect"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQuickConnectConfig_tags(rName, rName2, names.AttrTags, phone_number),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Quick Connect"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccQuickConnectConfig_tagsUpdated(rName, rName2, names.AttrTags, phone_number),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Quick Connect"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func testAccQuickConnect_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.QuickConnect
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	rName2 := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_connect_quick_connect.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickConnectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, "Disappear", "+12345678912"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfconnect.ResourceQuickConnect(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQuickConnectExists(ctx context.Context, t *testing.T, n string, v *awstypes.QuickConnect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ConnectClient(ctx)

		output, err := tfconnect.FindQuickConnectByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["quick_connect_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckQuickConnectDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_quick_connect" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).ConnectClient(ctx)

			_, err := tfconnect.FindQuickConnectByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["quick_connect_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Quick Connect %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccQuickConnectConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccQuickConnectConfig_phoneNumber(rName, rName2, label string, phoneNumber string) string {
	return acctest.ConfigCompose(
		testAccQuickConnectConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_quick_connect" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  quick_connect_config {
    quick_connect_type = "PHONE_NUMBER"

    phone_config {
      phone_number = %[3]q
    }
  }

  tags = {
    "Name" = "Test Quick Connect"
  }
}
`, rName2, label, phoneNumber))
}

func testAccQuickConnectConfig_tags(rName, rName2, label string, phoneNumber string) string {
	return acctest.ConfigCompose(
		testAccQuickConnectConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_quick_connect" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  quick_connect_config {
    quick_connect_type = "PHONE_NUMBER"

    phone_config {
      phone_number = %[3]q
    }
  }

  tags = {
    "Name" = "Test Quick Connect"
    "Key2" = "Value2a"
  }
}
`, rName2, label, phoneNumber))
}

func testAccQuickConnectConfig_tagsUpdated(rName, rName2, label string, phoneNumber string) string {
	return acctest.ConfigCompose(
		testAccQuickConnectConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_quick_connect" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  quick_connect_config {
    quick_connect_type = "PHONE_NUMBER"

    phone_config {
      phone_number = %[3]q
    }
  }

  tags = {
    "Name" = "Test Quick Connect"
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName2, label, phoneNumber))
}
