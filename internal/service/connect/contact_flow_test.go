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

func testAccContactFlow_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ContactFlow
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	rName2 := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_connect_contact_flow.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactFlowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowConfig_basic(rName, rName2, "Created"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactFlowExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/contact-flow/{contact_flow_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrContent),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.ContactFlowTypeContactFlow)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactFlowConfig_basic(rName, rName2, "Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContactFlowExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/contact-flow/{contact_flow_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrContent),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.ContactFlowTypeContactFlow)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
				),
			},
		},
	})
}

func testAccContactFlow_filename(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ContactFlow
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	rName2 := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_connect_contact_flow.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactFlowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowConfig_filename(rName, rName2, "Created", "test-fixtures/connect_contact_flow.json"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactFlowExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/contact-flow/{contact_flow_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.ContactFlowTypeContactFlow)),
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
				Config: testAccContactFlowConfig_filename(rName, rName2, "Updated", "test-fixtures/connect_contact_flow_updated.json"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContactFlowExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/contact-flow/{contact_flow_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.ContactFlowTypeContactFlow)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
				),
			},
		},
	})
}

func testAccContactFlow_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ContactFlow
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	rName2 := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_connect_contact_flow.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactFlowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowConfig_basic(rName, rName2, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactFlowExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfconnect.ResourceContactFlow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContactFlowExists(ctx context.Context, t *testing.T, n string, v *awstypes.ContactFlow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ConnectClient(ctx)

		output, err := tfconnect.FindContactFlowByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["contact_flow_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckContactFlowDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_contact_flow" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).ConnectClient(ctx)

			_, err := tfconnect.FindContactFlowByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["contact_flow_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Contact Flow %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccContactFlowConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccContactFlowConfig_basic(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccContactFlowConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_contact_flow" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q
  type        = "CONTACT_FLOW"
  content     = <<JSON
    {
		"Version": "2019-10-30",
		"StartAction": "12345678-1234-1234-1234-123456789012",
		"Actions": [
			{
				"Identifier": "12345678-1234-1234-1234-123456789012",
				"Type": "MessageParticipant",
				"Transitions": {
					"NextAction": "abcdef-abcd-abcd-abcd-abcdefghijkl",
					"Errors": [],
					"Conditions": []
				},
				"Parameters": {
					"Text": %[2]q
				}
			},
			{
				"Identifier": "abcdef-abcd-abcd-abcd-abcdefghijkl",
				"Type": "DisconnectParticipant",
				"Transitions": {},
				"Parameters": {}
			}
		]
    }
    JSON
  tags = {
    "Name"   = "Test Contact Flow",
    "Method" = %[2]q
  }
}
`, rName2, label))
}

func testAccContactFlowConfig_filename(rName, rName2 string, label string, filepath string) string {
	return acctest.ConfigCompose(
		testAccContactFlowConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_contact_flow" "test" {
  instance_id  = aws_connect_instance.test.id
  name         = %[1]q
  description  = %[2]q
  type         = "CONTACT_FLOW"
  filename     = %[3]q
  content_hash = filebase64sha256(%[3]q)
  tags = {
    "Name"   = "Test Contact Flow",
    "Method" = %[2]q
  }
}
`, rName2, label, filepath))
}
