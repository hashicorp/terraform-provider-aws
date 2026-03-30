// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfresourceexplorer2 "github.com/hashicorp/terraform-provider-aws/internal/service/resourceexplorer2"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_resourceexplorer2_index.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ResourceExplorer2EndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "resource-explorer-2", regexache.MustCompile(`index/`+verify.UUIDRegexPattern+`$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "LOCAL"),
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

func testAccIndex_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_resourceexplorer2_index.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ResourceExplorer2EndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfresourceexplorer2.ResourceIndex, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccIndex_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_resourceexplorer2_index.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ResourceExplorer2EndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIndexConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccIndexConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccIndex_type(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_resourceexplorer2_index.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ResourceExplorer2EndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_type("AGGREGATOR"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AGGREGATOR"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIndexConfig_type("LOCAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "LOCAL"),
				),
			},
			{
				Config:      testAccIndexConfig_type("AGGREGATOR"),
				ExpectError: regexache.MustCompile("cool down period has expired"),
				Check:       testAccCheckIndexDestroy(ctx, t),
			},
		},
	})
}

func testAccCheckIndexDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResourceExplorer2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resourceexplorer2_index" {
				continue
			}

			_, err := tfresourceexplorer2.FindIndex(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Resource Explorer Index %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIndexExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Resource Explorer Index ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ResourceExplorer2Client(ctx)

		_, err := tfresourceexplorer2.FindIndex(ctx, conn)

		return err
	}
}

var testAccIndexConfig_basic = testAccIndexConfig_type("LOCAL")

func testAccIndexConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_resourceexplorer2_index" "test" {
  type = "LOCAL"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccIndexConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_resourceexplorer2_index" "test" {
  type = "LOCAL"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccIndexConfig_type(typ string) string {
	return fmt.Sprintf(`
resource "aws_resourceexplorer2_index" "test" {
  type = %[1]q
}
`, typ)
}
