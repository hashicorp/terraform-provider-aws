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

func testAccContactFlowModule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ContactFlowModule
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow_module.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactFlowModuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowModuleConfig_basic(rName, rName2, "Created"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactFlowModuleExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/flow-module/{contact_flow_module_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_module_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrContent),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Created"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactFlowModuleConfig_basic(rName, rName2, "Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContactFlowModuleExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/flow-module/{contact_flow_module_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_module_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrContent),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
				),
			},
		},
	})
}

func testAccContactFlowModule_filename(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ContactFlowModule
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow_module.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactFlowModuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowModuleConfig_filename(rName, rName2, "Created", "test-fixtures/connect_contact_flow_module.json"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactFlowModuleExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/flow-module/{contact_flow_module_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_module_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Created"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"content_hash",
					"filename",
				},
			},
			{
				Config: testAccContactFlowModuleConfig_filename(rName, rName2, "Updated", "test-fixtures/connect_contact_flow_module_updated.json"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContactFlowModuleExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/flow-module/{contact_flow_module_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_module_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
				),
			},
		},
	})
}

func testAccContactFlowModule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ContactFlowModule
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow_module.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactFlowModuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowModuleConfig_basic(rName, rName2, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactFlowModuleExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconnect.ResourceContactFlowModule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContactFlowModuleExists(ctx context.Context, n string, v *awstypes.ContactFlowModule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectClient(ctx)

		output, err := tfconnect.FindContactFlowModuleByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["contact_flow_module_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckContactFlowModuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_contact_flow_module" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectClient(ctx)

			_, err := tfconnect.FindContactFlowModuleByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["contact_flow_module_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Contact Flow Module %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccContactFlowModuleConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccContactFlowModuleConfig_basic(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccContactFlowModuleConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_contact_flow_module" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  content = <<JSON
    {
		"Version": "2019-10-30",
		"StartAction": "12345678-1234-1234-1234-123456789012",
		"Actions": [
			{
				"Identifier": "12345678-1234-1234-1234-123456789012",
				"Parameters": {
					"Text": "Hello contact flow module"
				},
				"Transitions": {
					"NextAction": "abcdef-abcd-abcd-abcd-abcdefghijkl",
					"Errors": [],
					"Conditions": []
				},
				"Type": "MessageParticipant"
			},
			{
				"Identifier": "abcdef-abcd-abcd-abcd-abcdefghijkl",
				"Type": "DisconnectParticipant",
				"Parameters": {},
				"Transitions": {}
			}
		],
		"Settings": {
			"InputParameters": [],
			"OutputParameters": [],
			"Transitions": [
				{
					"DisplayName": "Success",
					"ReferenceName": "Success",
					"Description": ""
				},
				{
					"DisplayName": "Error",
					"ReferenceName": "Error",
					"Description": ""
				}
			]
		}
	}
    JSON

  tags = {
    "Name"   = "Test Contact Flow Module",
    "Method" = %[2]q
  }
}
`, rName2, label))
}

func testAccContactFlowModuleConfig_filename(rName, rName2 string, label string, filepath string) string {
	return acctest.ConfigCompose(
		testAccContactFlowModuleConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_contact_flow_module" "test" {
  instance_id  = aws_connect_instance.test.id
  name         = %[1]q
  description  = %[2]q
  filename     = %[3]q
  content_hash = filebase64sha256(%[3]q)

  tags = {
    "Name"   = "Test Contact Flow Module",
    "Method" = %[2]q
  }
}
`, rName2, label, filepath))
}
