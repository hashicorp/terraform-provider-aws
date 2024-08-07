// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/detective/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdetective "github.com/hashicorp/terraform-provider-aws/internal/service/detective"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccGraph_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var graph awstypes.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
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

func testAccGraph_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var graph awstypes.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdetective.ResourceGraph(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGraph_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var graph awstypes.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGraphConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGraphConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckGraphDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_detective_graph" {
				continue
			}

			_, err := tfdetective.FindGraphByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Detective Graph %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGraphExists(ctx context.Context, n string, v *awstypes.Graph) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveClient(ctx)

		output, err := tfdetective.FindGraphByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccGraphConfig_basic() string {
	return `
resource "aws_detective_graph" "test" {}
`
}

func testAccGraphConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_detective_graph" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccGraphConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_detective_graph" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
