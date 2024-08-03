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
					resource.TestCheckNoResourceAttr(dataSourceName, names.AttrDescription),
				),
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
