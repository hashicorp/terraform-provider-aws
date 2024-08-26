// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package outposts_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOutpostsSiteDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_outposts_site.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSites(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrAccountID(dataSourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrDescription),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrID, regexache.MustCompile(`^os-.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrName, regexache.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

func TestAccOutpostsSiteDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	sourceDataSourceName := "data.aws_outposts_site.source"
	dataSourceName := "data.aws_outposts_site.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSites(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteDataSourceConfig_name(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAccountID, sourceDataSourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, sourceDataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, sourceDataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, sourceDataSourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccSiteDataSourceConfig_id() string {
	return `
data "aws_outposts_sites" "test" {}

data "aws_outposts_site" "test" {
  id = tolist(data.aws_outposts_sites.test.ids)[0]
}
`
}

func testAccSiteDataSourceConfig_name() string {
	return `
data "aws_outposts_sites" "test" {}

data "aws_outposts_site" "source" {
  id = tolist(data.aws_outposts_sites.test.ids)[0]
}

data "aws_outposts_site" "test" {
  name = data.aws_outposts_site.source.name
}
`
}
