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

func TestAccImageBuilderContainerRecipeDataSource_arn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_container_recipe.test"
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeDataSourceConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.#", resourceName, "component.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.0.component_arn", resourceName, "component.0.component_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "component.0.parameter.#", resourceName, "component.0.parameter.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "container_type", resourceName, "container_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_created", resourceName, "date_created"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "dockerfile_template_data", resourceName, "dockerfile_template_data"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEncrypted, resourceName, names.AttrEncrypted),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_configuration.#", resourceName, "instance_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwner, resourceName, names.AttrOwner),
					resource.TestCheckResourceAttrPair(dataSourceName, "parent_image", resourceName, "parent_image"),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform", resourceName, "platform"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "target_repository.#", resourceName, "target_repository.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "target_repository.0.repository_name", resourceName, "target_repository.0.repository_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "target_repository.0.service", resourceName, "target_repository.0.service"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
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
