package resourceexplorer2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfresourceexplorer2 "github.com/hashicorp/terraform-provider-aws/internal/service/resourceexplorer2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_resourceexplorer2_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(names.ResourceExplorer2EndpointID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "resource-explorer-2", regexp.MustCompile(`index/+.`)),
					resource.TestCheckResourceAttr(resourceName, "type", "LOCAL"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(names.ResourceExplorer2EndpointID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(acctest.Provider, tfresourceexplorer2.ResourceIndex, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccIndex_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_resourceexplorer2_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(names.ResourceExplorer2EndpointID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName),
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
				Config: testAccIndexConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIndexConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccIndex_type(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_resourceexplorer2_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(names.ResourceExplorer2EndpointID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_type("AGGREGATOR"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "type", "AGGREGATOR"),
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
					testAccCheckIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "type", "LOCAL"),
				),
			},
		},
	})
}

func testAccCheckIndexDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ResourceExplorer2Client()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resourceexplorer2_index" {
				continue
			}

			_, err := tfresourceexplorer2.FindIndex(ctx, conn)

			if tfresource.NotFound(err) {
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

func testAccCheckIndexExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Resource Explorer Index ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ResourceExplorer2Client()

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
