// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccImageBuilderComponentsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_components.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentsDataSourceConfig_component(rName),
			},
			{
				Config: testAccComponentsDataSourceConfig_component2(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
				),
			},
		},
	})
}

func testAccComponentsDataSourceConfig_component(rName string) string {
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
  version  = "1.0.0"
}
`, rName)
}

func testAccComponentsDataSourceConfig_component2(rName string) string {
	return acctest.ConfigCompose(
		testAccComponentsDataSourceConfig_component(rName),
		`
data "aws_imagebuilder_components" "test" {
  filter {
    name   = "name"
    values = [aws_imagebuilder_component.test.name]
  }
}
`)
}
