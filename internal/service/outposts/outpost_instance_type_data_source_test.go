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

func TestAccOutpostsOutpostInstanceTypeDataSource_instanceType(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_outposts_outpost_instance_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostInstanceTypeDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrInstanceType, regexache.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

func TestAccOutpostsOutpostInstanceTypeDataSource_preferredInstanceTypes(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_outposts_outpost_instance_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostInstanceTypeDataSourceConfig_preferreds(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrInstanceType, regexache.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

func testAccOutpostInstanceTypeDataSourceConfig_basic() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost_instance_types" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_outposts_outpost_instance_type" "test" {
  arn           = tolist(data.aws_outposts_outposts.test.arns)[0]
  instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
}
`
}

func testAccOutpostInstanceTypeDataSourceConfig_preferreds() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost_instance_types" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_outposts_outpost_instance_type" "test" {
  arn                      = tolist(data.aws_outposts_outposts.test.arns)[0]
  preferred_instance_types = data.aws_outposts_outpost_instance_types.test.instance_types
}
`
}
