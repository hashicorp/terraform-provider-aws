// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/synthetics"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsynthetics "github.com/hashicorp/terraform-provider-aws/internal/service/synthetics"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSyntheticsGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group synthetics.Group
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", synthetics.ServiceName, regexp.MustCompile(`group:.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccSyntheticsGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var group synthetics.Group
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsynthetics.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSyntheticsGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var group synthetics.Group
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckGroupExists(ctx context.Context, name string, group *synthetics.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Synthetics Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SyntheticsConn(ctx)
		output, err := tfsynthetics.FindGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*group = *output

		return nil
	}
}

func testAccCheckGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SyntheticsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_synthetics_group" {
				continue
			}

			_, err := tfsynthetics.FindGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("synthetics Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_synthetics_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_synthetics_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_synthetics_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
