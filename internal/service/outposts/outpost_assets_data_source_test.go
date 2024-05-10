// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package outposts_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOutpostsAssetsDataSource_id(t *testing.T) {
	key := "OUTPOST_AVAIL"
	outpostBool := os.Getenv(key)
	if outpostBool == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_outposts_assets.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostAssetsDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "outposts", regexache.MustCompile(`outpost/.+`)),
				),
			},
		},
	})
}

func TestAccOutpostsAssetsDataSource_statusFilter(t *testing.T) {
	key := "OUTPOST_AVAIL"
	outpostBool := os.Getenv(key)
	if outpostBool == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_outposts_assets.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostAssetsDataSourceConfig_statusFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "status_id_filter.0", "ACTIVE"),
				),
			},
		},
	})
}

func TestAccOutpostsAssetsDataSource_hostFilter(t *testing.T) {
	key := "OUTPOST_HOST_ID" // Ex. "h-x38g5n0yd2a0ueb61"
	outpostHostId := os.Getenv(key)
	if outpostHostId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_outposts_assets.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostAssetsDataSourceConfig_hostFilter(outpostHostId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(dataSourceName, "host_id_filter.*", outpostHostId),
				),
			},
		},
	})
}

func testAccOutpostAssetsDataSourceConfig_id() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_assets" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}
`
}

func testAccOutpostAssetsDataSourceConfig_statusFilter() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_assets" "source" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_outposts_assets" "test" {
  arn              = data.aws_outposts_assets.source.arn
  status_id_filter = ["ACTIVE"]
}
`
}

func testAccOutpostAssetsDataSourceConfig_hostFilter(outpostHostId string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_assets" "source" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_outposts_assets" "test" {
  arn            = data.aws_outposts_assets.source.arn
  host_id_filter = [%[1]q]
}
`, outpostHostId)
}
