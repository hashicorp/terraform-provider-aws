// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTag_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_transfer_tag.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1),
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

func testAccTag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_transfer_tag.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tftransfer.ResourceTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccTag_value(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_transfer_tag.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1Updated),
				),
			},
		},
	})
}

func testAccTag_system(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_transfer_tag.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, "aws:transfer:customHostname", "abc.example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, "aws:transfer:customHostname"),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "abc.example.com"),
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

func testAccTagConfig_basic(rName string, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"

  tags = {
    Name = %[1]q
  }

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_transfer_tag" "test" {
  resource_arn = aws_transfer_server.test.arn
  key          = %[2]q
  value        = %[3]q
}
`, rName, key, value)
}
