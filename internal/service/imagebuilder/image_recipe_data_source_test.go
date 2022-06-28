package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/imagebuilder"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccImageBuilderImageRecipeDataSource_arn(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_image_recipe.test"
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeDataSourceConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "block_device_mapping.#", resourceName, "block_device_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.#", resourceName, "component.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.0.component_arn", resourceName, "component.0.component_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.0.parameter.#", resourceName, "component.0.parameter.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.0.parameter.0.name", resourceName, "component.0.parameter.0.name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.0.parameter.0.value", resourceName, "component.0.parameter.0.value"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_created", resourceName, "date_created"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner", resourceName, "owner"),
					resource.TestCheckResourceAttrPair(dataSourceName, "parent_image", resourceName, "parent_image"),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform", resourceName, "platform"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_data_base64", resourceName, "user_data_base64"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "working_directory", resourceName, "working_directory"),
				),
			},
		},
	})
}

func testAccImageRecipeDataSourceConfig_arn(rName string) string {
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
    parameters = [{
      Parameter1 = {
        type = "string"
      }
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"
}

resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn

    parameter {
      name  = "Parameter1"
      value = "Value1"
    }
  }

  name             = %[1]q
  parent_image     = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version          = "1.0.0"
  user_data_base64 = base64encode("helloworld")
}

data "aws_imagebuilder_image_recipe" "test" {
  arn = aws_imagebuilder_image_recipe.test.arn
}
`, rName)
}
