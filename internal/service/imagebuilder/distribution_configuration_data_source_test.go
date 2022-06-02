package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/imagebuilder"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccImageBuilderDistributionConfigurationDataSource_arn(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_distribution_configuration.test"
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDataSourceConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_created", resourceName, "date_created"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_updated", resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.#", resourceName, "distribution.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.#", resourceName, "distribution.0.container_distribution_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.0.container_tags.#", resourceName, "distribution.0.container_distribution_configuration.0.container_tags.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.0.description", resourceName, "distribution.0.container_distribution_configuration.0.description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.0.target_repository.#", resourceName, "distribution.0.container_distribution_configuration.0.target_repository.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.repository_name", resourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.repository_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.service", resourceName, "distribution.0.container_distribution_configuration.0.target_repository.0.service"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.launch_template_configuration.#", resourceName, "distribution.0.launch_template_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.launch_template_configuration.0.default", resourceName, "distribution.0.launch_template_configuration.0.default"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.launch_template_configuration.0.launch_template_id", resourceName, "distribution.0.launch_template_configuration.0.launch_template_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution.0.launch_template_configuration.0.account_id", resourceName, "distribution.0.launch_template_configuration.0.account_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
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

    region = data.aws_region.current.name
  }
}

data "aws_imagebuilder_distribution_configuration" "test" {
  arn = aws_imagebuilder_distribution_configuration.test.arn
}
`, rName)
}
