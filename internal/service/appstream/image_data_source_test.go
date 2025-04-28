// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamImageDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
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
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrARN), // nosemgrep:ci.semgrep.acctest.checks.arn-resourceattrset
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrDisplayName),
					resource.TestCheckResourceAttrSet(dataSourceName, "image_builder_supported"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "name_regex", "^AppStream-WinServer.*$"),
					resource.TestCheckResourceAttrSet(dataSourceName, "platform"),
					resource.TestCheckResourceAttrSet(dataSourceName, "public_base_image_released_date"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrType, string(awstypes.VisibilityTypePublic)),
				),
			},
		},
	})
}

func testAccImageDataSourceConfig_basic() string {
	return `
data "aws_appstream_image" "test" {
  name_regex  = "^AppStream-WinServer.*$"
  type        = "PUBLIC"
  most_recent = true
}
`
}
