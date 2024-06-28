// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogAppRegistryAttributeGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var attributegroup servicecatalogappregistry.GetAttributeGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDesc := "Simple Description"
	expectJsonV1 := `{"a":"1","b":"2"}`
	dataSourceName := "data.aws_servicecatalogappregistry_attribute_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttributeGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAttributeGroupDataSourceConfig_basic(rName, rDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttributeGroupExists(ctx, dataSourceName, &attributegroup),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, rDesc),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrAttributes, expectJsonV1),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "servicecatalog", regexache.MustCompile(`/attribute-groups/+.`)),
				),
			},
		},
	})
}

func testAccAttributeGroupDataSourceConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
data "aws_servicecatalogappregistry_attribute_group" "test" {
	name = aws_servicecatalogappregistry_attribute_group.test.name
}

resource "aws_servicecatalogappregistry_attribute_group" "test" {
  name        = %[1]q
  description = %[2]q
  attributes = jsonencode({
    a = "1"
    b = "2"
  })
  tags = {
    tag1 = "v1"
    tag2 = "v2"
  }
}
`, rName, description)
}
