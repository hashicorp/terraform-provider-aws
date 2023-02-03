package resourceexplorer2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfresourceexplorer2 "github.com/hashicorp/terraform-provider-aws/internal/service/resourceexplorer2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccView_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v resourceexplorer2.GetViewOutput
	resourceName := "aws_resourceexplorer2_view.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(names.ResourceExplorer2EndpointID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckViewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccViewConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "resource-explorer-2", regexp.MustCompile(`view/+.`)),
					resource.TestCheckResourceAttr(resourceName, "default_view", "false"),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "included_property.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccView_defaultView(t *testing.T) {
	ctx := acctest.Context(t)
	var v resourceexplorer2.GetViewOutput
	resourceName := "aws_resourceexplorer2_view.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(names.ResourceExplorer2EndpointID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckViewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccViewConfig_defaultView(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_view", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccViewConfig_defaultView(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_view", "false"),
				),
			},
			{
				Config: testAccViewConfig_defaultView(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_view", "true"),
				),
			},
		},
	})
}

func testAccView_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v resourceexplorer2.GetViewOutput
	resourceName := "aws_resourceexplorer2_view.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(names.ResourceExplorer2EndpointID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckViewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccViewConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(acctest.Provider, tfresourceexplorer2.ResourceView, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccView_filter(t *testing.T) {
	ctx := acctest.Context(t)
	var v resourceexplorer2.GetViewOutput
	resourceName := "aws_resourceexplorer2_view.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(names.ResourceExplorer2EndpointID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckViewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccViewConfig_filter(rName, "resourcetype:ec2:instance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "resource-explorer-2", regexp.MustCompile(`view/+.`)),
					resource.TestCheckResourceAttr(resourceName, "default_view", "false"),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.filter_string", "resourcetype:ec2:instance"),
					resource.TestCheckResourceAttr(resourceName, "included_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "included_property.0.name", "tags"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccViewConfig_filter(rName, "region:global"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "resource-explorer-2", regexp.MustCompile(`view/+.`)),
					resource.TestCheckResourceAttr(resourceName, "default_view", "false"),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.filter_string", "region:global"),
					resource.TestCheckResourceAttr(resourceName, "included_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "included_property.0.name", "tags"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccView_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v resourceexplorer2.GetViewOutput
	resourceName := "aws_resourceexplorer2_view.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(names.ResourceExplorer2EndpointID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckViewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccViewConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &v),
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
				Config: testAccViewConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccViewConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckViewDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ResourceExplorer2Client()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resourceexplorer2_iview" {
				continue
			}

			_, err := tfresourceexplorer2.FindViewByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Resource Explorer View %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckViewExists(ctx context.Context, n string, v *resourceexplorer2.GetViewOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Resource Explorer View ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ResourceExplorer2Client()

		output, err := tfresourceexplorer2.FindViewByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return err
	}
}

func testAccViewConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resourceexplorer2_index" "test" {
  type = "LOCAL"

  tags = {
    Name = %[1]q
  }
}

resource "aws_resourceexplorer2_view" "test" {
  name = %[1]q

  depends_on = [aws_resourceexplorer2_index.test]
}
`, rName)
}

func testAccViewConfig_defaultView(rName string, defaultView bool) string {
	return fmt.Sprintf(`
resource "aws_resourceexplorer2_index" "test" {
  type = "LOCAL"

  tags = {
    Name = %[1]q
  }
}

resource "aws_resourceexplorer2_view" "test" {
  name         = %[1]q
  default_view = %[2]t

  depends_on = [aws_resourceexplorer2_index.test]
}
`, rName, defaultView)
}

func testAccViewConfig_filter(rName, filter string) string {
	return fmt.Sprintf(`
resource "aws_resourceexplorer2_index" "test" {
  type = "LOCAL"

  tags = {
    Name = %[1]q
  }
}

resource "aws_resourceexplorer2_view" "test" {
  name = %[1]q

  filters {
    filter_string = %[2]q
  }

  included_property {
    name = "tags"
  }

  depends_on = [aws_resourceexplorer2_index.test]
}
`, rName, filter)
}

func testAccViewConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_resourceexplorer2_index" "test" {
  type = "LOCAL"

  tags = {
    Name = %[1]q
  }
}

resource "aws_resourceexplorer2_view" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_resourceexplorer2_index.test]
}
`, rName, tagKey1, tagValue1)
}

func testAccViewConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_resourceexplorer2_index" "test" {
  type = "LOCAL"

  tags = {
    Name = %[1]q
  }
}

resource "aws_resourceexplorer2_view" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_resourceexplorer2_index.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
