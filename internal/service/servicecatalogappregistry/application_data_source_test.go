// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogAppRegistryApplicationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_servicecatalogappregistry_application.test"
	applicationResourceName := "aws_servicecatalogappregistry_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, applicationResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "servicecatalog", regexache.MustCompile(`/applications/+.`)),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalogappregistry_application" "test" {
  name = %[1]q
}

data "aws_servicecatalogappregistry_application" "test" {
  id = aws_servicecatalogappregistry_application.test.id
}

resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "String"
  value = "test"
  tags  = data.aws_servicecatalogappregistry_application.test.application_tag
}
`, rName)
}
