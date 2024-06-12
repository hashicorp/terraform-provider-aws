// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"

	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this data source's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

func TestAccGlueRegistryDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var registry glue.GetRegistryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_glue_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueServiceID)
			testAccPreCheckRegistry(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, dataSourceName, &registry),
					acctest.CheckResourceAttrRegionalARN(dataSourceName, names.AttrARN, "glue", fmt.Sprintf("registry/%s", rName)),
					resource.TestCheckResourceAttr(dataSourceName, "registry_name", rName),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      dataSourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRegistryDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q
}

data "aws_glue_registry" "test" {
  id = aws_glue_registry.test.id
}
`, rName)
}
