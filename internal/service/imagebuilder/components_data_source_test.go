package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/imagebuilder"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccImageBuilderComponentsDataSource_owner(t *testing.T) {
	dataSourceName := "data.aws_imagebuilder_components.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComponentsOwnerDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "0"),
				),
			},
		},
	})
}

func TestAccImageBuilderComponentsDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_components.test"
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComponentsFilterDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arns.0", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, "name"),
				),
			},
		},
	})
}

const testAccComponentsOwnerDataSourceConfig = `
data "aws_imagebuilder_components" "test" {
  owner = "Self"
}
`

func testAccComponentsFilterDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
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
  version  = "1.0.1"
}

data "aws_imagebuilder_components" "test" {
  filter {
    name   = "name"
    values = [aws_imagebuilder_component.test.name]
  }
}
`, rName)
}
