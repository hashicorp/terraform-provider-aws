// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resourcegroupstaggingapi_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResourceGroupsTaggingAPIRequiredTagsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_resourcegroupstaggingapi_required_tags.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ResourceGroupsTaggingAPIEndpointID)
			acctest.PreCheckResourceGroupsTaggingAPIRequiredTags(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceGroupsTaggingAPIServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRequiredTagsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "required_tags.#"),
				),
			},
		},
	})
}

func testAccRequiredTagsDataSourceConfig_basic() string {
	return `
data "aws_resourcegroupstaggingapi_required_tags" "test" {}
`
}
