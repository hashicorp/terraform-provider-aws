// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccImageBuilderImageRecipesDataSource_owner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceNameOwnerAmazon := "data.aws_imagebuilder_image_recipes.amazon"
	dataSourceNameOwnerSelf := "data.aws_imagebuilder_image_recipes.self"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipesDataSourceConfig_owner(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceNameOwnerAmazon, "arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceNameOwnerAmazon, "names.#", acctest.Ct0),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceNameOwnerSelf, "arns.#", 1),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceNameOwnerSelf, "names.#", 1),
				),
			},
		},
	})
}

func TestAccImageBuilderImageRecipesDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_image_recipes.test"
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipesDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccImageRecipeDataSourceConfig_base(rName string) string {
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
	return acctest.ConfigCompose(testAccImageRecipeDataSourceConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccImageRecipeDataSourceConfig_base(rName), fmt.Sprintf(`
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
