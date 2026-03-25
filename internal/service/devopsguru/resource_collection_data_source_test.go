// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsguru_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccResourceCollectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_devopsguru_resource_collection.test"
	resourceName := "aws_devopsguru_resource_collection.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCollectionDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, resourceName, names.AttrType),
				),
			},
		},
	})
}

func testAccResourceCollectionDataSourceConfig_basic() string {
	return `
resource "aws_devopsguru_resource_collection" "test" {
  type = "AWS_SERVICE"
  cloudformation {
    stack_names = ["*"]
  }
}

data "aws_devopsguru_resource_collection" "test" {
  type = aws_devopsguru_resource_collection.test.type
}
`
}
