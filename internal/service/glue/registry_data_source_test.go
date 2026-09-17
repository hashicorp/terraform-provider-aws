// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/glue"
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_glue_registry.test"
	resourceName := "aws_glue_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRegistry(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName, &registry),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, "registry_name"),
				),
			},
		},
	})
}

func testAccRegistryDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q
  description   = %[1]q
}

data "aws_glue_registry" "test" {
  name = aws_glue_registry.test.registry_name
}
`, rName)
}
