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

func TestAccImageBuilderDistributionConfigurationDataSource_arn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_distribution_configuration.test"
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDataSourceConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_created", resourceName, "date_created"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_updated", resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.#", resourceName, "distribution.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.#", resourceName, "distribution.0.container_distribution_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.0.container_tags.#", resourceName, "distribution.0.container_distribution_configuration.0.container_tags.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.0.description", resourceName, "distribution.0.container_distribution_configuration.0.description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.0.target_repository.#", resourceName, "distribution.0.container_distribution_configuration.0.target_repository.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.repository_name", resourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.repository_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.service", resourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.service"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.fast_launch_configuration.#", resourceName, "distribution.0.fast_launch_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.fast_launch_configuration.0.account_id", resourceName, "distribution.0.fast_launch_configuration.0.account_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.fast_launch_configuration.0.enabled", resourceName, "distribution.0.fast_launch_configuration.0.enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.fast_launch_configuration.0.launch_template.#", resourceName, "distribution.0.fast_launch_configuration.0.launch_template.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.fast_launch_configuration.0.launch_template.0.launch_template_id", resourceName, "distribution.0.fast_launch_configuration.0.launch_template.0.launch_template_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.fast_launch_configuration.0.launch_template.0.launch_template_name", resourceName, "distribution.0.fast_launch_configuration.0.launch_template.0.launch_template_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.fast_launch_configuration.0.launch_template.0.launch_template_version", resourceName, "distribution.0.fast_launch_configuration.0.launch_template.0.launch_template_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.fast_launch_configuration.0.max_parallel_launches", resourceName, "distribution.0.fast_launch_configuration.0.max_parallel_launches"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.fast_launch_configuration.0.snapshot_configuration.#", resourceName, "distribution.0.fast_launch_configuration.0.snapshot_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.fast_launch_configuration.0.snapshot_configuration.0.target_resource_count", resourceName, "distribution.0.fast_launch_configuration.0.snapshot_configuration.0.target_resource_count"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.launch_template_configuration.#", resourceName, "distribution.0.launch_template_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.launch_template_configuration.0.default", resourceName, "distribution.0.launch_template_configuration.0.default"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.launch_template_configuration.0.launch_template_id", resourceName, "distribution.0.launch_template_configuration.0.launch_template_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.launch_template_configuration.0.account_id", resourceName, "distribution.0.launch_template_configuration.0.account_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccDistributionConfigurationDataSourceConfig_arn(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_launch_template" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    container_distribution_configuration {
      target_repository {
        repository_name = "repository-name"
        service         = "ECR"
      }
    }

    launch_template_configuration {
      account_id         = data.aws_caller_identity.current.account_id
      default            = false
      launch_template_id = aws_launch_template.test.id
    }

    fast_launch_configuration {
      account_id = data.aws_caller_identity.current.account_id
      enabled    = true

      launch_template {
        launch_template_id      = aws_launch_template.test.id
        launch_template_version = "1"
      }

      max_parallel_launches = 6

      snapshot_configuration {
        target_resource_count = 1
      }
    }

    region = data.aws_region.current.name
  }
}

data "aws_imagebuilder_distribution_configuration" "test" {
  arn = aws_imagebuilder_distribution_configuration.test.arn
}
`, rName)
}
