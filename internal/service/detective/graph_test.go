// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package detective_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/detective/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdetective "github.com/hashicorp/terraform-provider-aws/internal/service/detective"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccGraph_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var graph awstypes.Graph
	resourceName := "aws_detective_graph.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "graph_arn", "detective", regexache.MustCompile(`graph:[a-z0-9]{32}$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, "graph_arn"),
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

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdetective.ResourceGraph(), resourceName),
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

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
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
				Config: testAccGraphConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGraphConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(ctx, t, resourceName, &graph),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckGraphDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DetectiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_detective_graph" {
				continue
			}

			_, err := tfdetective.FindGraphByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckGraphExists(ctx context.Context, t *testing.T, n string, v *awstypes.Graph) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DetectiveClient(ctx)

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
