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

func testAccQuickConnect_phoneNumber(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeQuickConnectOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_quick_connect.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickConnectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, "Created", "+12345678912"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.quick_connect_type", "PHONE_NUMBER"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.0.phone_number", "+12345678912"),

					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckQuickConnectExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.quick_connect_type", "PHONE_NUMBER"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.0.phone_number", "+12345678912"),

					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckQuickConnectExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.quick_connect_type", "PHONE_NUMBER"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.0.phone_number", "+12345678913"),

					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
		},
	})
}

func testAccQuickConnect_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeQuickConnectOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	phone_number := "+12345678912"

	resourceName := "aws_connect_quick_connect.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickConnectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, names.AttrTags, phone_number),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckQuickConnectExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Quick Connect"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccQuickConnectConfig_tagsUpdated(rName, rName2, names.AttrTags, phone_number),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
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
	var v connect.DescribeQuickConnectOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_quick_connect.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickConnectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, "Disappear", "+12345678912"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickConnectExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconnect.ResourceQuickConnect(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQuickConnectExists(ctx context.Context, resourceName string, function *connect.DescribeQuickConnectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Quick Connect not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Quick Connect ID not set")
		}
		instanceID, quickConnectID, err := tfconnect.QuickConnectParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

		params := &connect.DescribeQuickConnectInput{
			QuickConnectId: aws.String(quickConnectID),
			InstanceId:     aws.String(instanceID),
		}

		getFunction, err := conn.DescribeQuickConnectWithContext(ctx, params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckQuickConnectDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_quick_connect" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

			instanceID, quickConnectID, err := tfconnect.QuickConnectParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			params := &connect.DescribeQuickConnectInput{
				QuickConnectId: aws.String(quickConnectID),
				InstanceId:     aws.String(instanceID),
			}

			_, err = conn.DescribeQuickConnectWithContext(ctx, params)

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
