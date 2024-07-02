// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamImageDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	dataSourceName := "data.aws_appstream_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppStreamEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImageDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckResourceAttrSet(dataSourceName, "applications.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "appstream_agent_version"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrDisplayName),
					resource.TestCheckResourceAttrSet(dataSourceName, "image_builder_supported"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(dataSourceName, "name_regex"),
					resource.TestCheckResourceAttrSet(dataSourceName, "platform"),
					resource.TestCheckResourceAttrSet(dataSourceName, "public_base_image_released_date"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrType),
				),
			},
		},
	})
}

// name        = "AppStream-WinServer2019-06-17-2024"
func testAccImageDataSourceConfig_basic() string {
	return (`
data "aws_appstream_image" "test" {
  name_regex  = "^AppStream-WinServer.*$"
  type        = "PUBLIC"
  most_recent = true
}
`)
}
