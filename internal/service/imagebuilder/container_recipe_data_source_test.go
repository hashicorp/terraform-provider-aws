package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/imagebuilder"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccImageBuilderContainerRecipeDataSource_arn(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_container_recipe.test"
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeDataSourceConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.#", resourceName, "component.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.0.component_arn", resourceName, "component.0.component_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.0.parameter.#", resourceName, "component.0.parameter.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "container_type", resourceName, "container_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_created", resourceName, "date_created"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dockerfile_template_data", resourceName, "dockerfile_template_data"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encrypted", resourceName, "encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_configuration.#", resourceName, "instance_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner", resourceName, "owner"),
					resource.TestCheckResourceAttrPair(dataSourceName, "parent_image", resourceName, "parent_image"),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform", resourceName, "platform"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "target_repository.#", resourceName, "target_repository.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "target_repository.0.repository_name", resourceName, "target_repository.0.repository_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "target_repository.0.service", resourceName, "target_repository.0.service"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "working_directory", resourceName, "working_directory"),
				),
			},
		},
	})
}

func testAccContainerRecipeDataSourceConfig_arn(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_imagebuilder_container_recipe" "test" {
  name           = %[1]q
  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}

data "aws_imagebuilder_container_recipe" "test" {
  arn = aws_imagebuilder_container_recipe.test.arn
}
`, rName)
}
