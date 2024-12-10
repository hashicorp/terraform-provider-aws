// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogAppRegistryAttributeGroupAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_servicecatalogappregistry_attribute_group_associations.test"
	resourceName := "aws_servicecatalogappregistry_attribute_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttributeGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAttributeGroupAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttributeGroupAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(dataSourceName, "attribute_group_ids.#", "1"),
				),
			},
		},
	})
}

func testAccApplicationAttributeGroupAssociationsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAttributeGroupAssociationConfig_basic(rName),
		`
data "aws_servicecatalogappregistry_attribute_group_associations" "test" {
  id = aws_servicecatalogappregistry_application.test.id

  depends_on = [aws_servicecatalogappregistry_attribute_group_association.test]
}
`)
}
