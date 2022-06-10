package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/imagebuilder"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccImageBuilderImageRecipesDataSource_owner(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceNameOwnerAmazon := "data.aws_imagebuilder_image_recipes.amazon"
	dataSourceNameOwnerSelf := "data.aws_imagebuilder_image_recipes.self"
	resourceName := "aws_imagebuilder_image_recipe.test"

	// Not a good test since it is susceptible to fail with parallel tests or if anything else
	// ImageBuilder is going on in the account
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipesDataSourceConfig_owner(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceNameOwnerAmazon, "arns.#", "0"),
					resource.TestCheckResourceAttr(dataSourceNameOwnerAmazon, "names.#", "0"),
					resource.TestCheckResourceAttr(dataSourceNameOwnerSelf, "arns.#", "1"),
					resource.TestCheckResourceAttr(dataSourceNameOwnerSelf, "names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceNameOwnerSelf, "arns.0", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceNameOwnerSelf, "names.0", resourceName, "name"),
				),
			},
		},
	})
}

func TestAccImageBuilderImageRecipesDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_image_recipes.test"
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipesDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, "name"),
				),
			},
		},
	})
}

func testAccImageRecipeDataSourceBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"
}
`, rName)
}

func testAccImageRecipesDataSourceConfig_owner(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeDataSourceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}

data "aws_imagebuilder_image_recipes" "amazon" {
  owner = "Amazon"

  depends_on = [
    aws_imagebuilder_image_recipe.test
  ]
}

data "aws_imagebuilder_image_recipes" "self" {
  owner = "Self"

  depends_on = [
    aws_imagebuilder_image_recipe.test
  ]
}
`, rName))
}

func testAccImageRecipesDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeDataSourceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}

data "aws_imagebuilder_image_recipes" "test" {
  filter {
    name   = "name"
    values = [aws_imagebuilder_image_recipe.test.name]
  }
}
`, rName))
}
