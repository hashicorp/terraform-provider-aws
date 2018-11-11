package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAWSGlueScript_Language_Python(t *testing.T) {
	dataSourceName := "data.aws_glue_script.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSGlueScriptConfig_Language("PYTHON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "python_script"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSGlueScript_Language_Scala(t *testing.T) {
	dataSourceName := "data.aws_glue_script.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSGlueScriptConfig_Language("SCALA"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "scala_code"),
				),
			},
		},
	})
}

func testAccDataSourceAWSGlueScriptConfig_Language(language string) string {
	return fmt.Sprintf(`
data "aws_glue_script" "test" {
  dag_edge = []
  dag_node = []
  language = "%s"
}
`, language)
}
