// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogAppRegistryAttributeGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var attributegroup servicecatalogappregistry.GetAttributeGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description := "Simple Description"
	expectJsonV1 := `{"a":"1","b":"2"}`
	dataSourceName := "data.aws_servicecatalogappregistry_attribute_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttributeGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttributeGroupDataSourceConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttributeGroupExists(ctx, t, dataSourceName, &attributegroup),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrAttributes, expectJsonV1),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "servicecatalog", regexache.MustCompile(`/attribute-groups/+.`)),
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
}
`, rName, description)
}
