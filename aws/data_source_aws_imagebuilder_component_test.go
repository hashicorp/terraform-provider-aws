package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceAwsImageBuilderComponent_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-ib-comp-test-%s", acctest.RandString(8))
	resourceName := "aws_builder_component.test"
	dataSourceName := "data.aws_builder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsImageBuilderComponentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "semantic_version", resourceName, "semantic_version"),
				),
			},
		},
	})
}

func testAccDataSourceAwsImageBuilderComponentConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  name             = "%s"
  platform         = "Linux"
  semantic_version = "1.0.1"

  data = <<EOD
name: HelloWorldTestingDocument
description: This is hello world testing document.
schemaVersion: 1.0

phases:
  - name: test
    steps:
      - name: HelloWorldStep
        action: ExecuteBash
        inputs:
          commands:
            - echo "Hello World! Test."
EOD
}

data "aws_builder_component" "test" {
  arn = "arn:aws:imagebuilder:eu-west-1:116147290797:component/test/1.0.1/1"
}
`, name)
}
