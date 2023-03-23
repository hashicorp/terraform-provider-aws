package datapipeline_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/datapipeline"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataPipelinePipelineDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_datapipeline_pipeline.test"
	resourceName := "aws_datapipeline_pipeline.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPipelineDefinitionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "pipeline_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
				),
			},
		},
	})
}

func testAccPipelineDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "test" {
  name = %[1]q
}

data "aws_datapipeline_pipeline" "test" {
  pipeline_id = aws_datapipeline_pipeline.test.id
}
`, name)
}
