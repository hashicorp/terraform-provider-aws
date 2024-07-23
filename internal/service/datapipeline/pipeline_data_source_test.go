// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datapipeline_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataPipelinePipelineDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_datapipeline_pipeline.test"
	resourceName := "aws_datapipeline_pipeline.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDefinitionDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.DataPipelineServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "pipeline_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
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
