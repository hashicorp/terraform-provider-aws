// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_vpc_ipams.test"
	dataSourceNameTwo := "data.aws_vpc_ipams.testtwo"
	resourceName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ipams.#", 0),
				),
			},
			{
				Config: testAccIPAMsDataSourceConfig_basicWithFilter,
				Check: resource.ComposeAggregateTestCheckFunc(
					// DS 1 finds all 2 IPAMs
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "ipams.#", 1),

					// DS 2 filters on 1 specific IPAM to validate attributes
					resource.TestCheckResourceAttr(dataSourceNameTwo, "ipams.#", "1"),
					// resource.TestCheckResourceAttr(dataSourceNameTwo, "ipams.0.tags.test", "two"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipams.0.enable_private_gua", resourceName, "enable_private_gua"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipams.0.tier", resourceName, "tier"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipams.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipams.0.description", resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipams.0.id", resourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccIPAMsDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_vpc_ipams.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMsDataSourceConfig_empty,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipams.#", "0"),
				),
			},
		},
	})
}

var testAccIPAMsDataSourceConfig_basic = acctest.ConfigCompose(testAccIPAMConfig_basic, `
data "aws_vpc_ipams" "test" {
  depends_on = [
    aws_vpc_ipam.test,
  ]
}
`)

var testAccIPAMsDataSourceConfig_basicWithFilter = acctest.ConfigCompose(testAccIPAMConfig_description("Some desc"), `
data "aws_vpc_ipams" "test" {
  depends_on = [
    aws_vpc_ipam.test,
  ]
}
  
data "aws_vpc_ipams" "testtwo" {
  filter {
    name   = "description"
    values = ["*Some*"]
  }
  depends_on = [
	aws_vpc_ipam.test,
  ]
}
`)

const testAccIPAMsDataSourceConfig_empty = `
data "aws_vpc_ipams" "test" {
  filter {
    name   = "description"
    values = ["*none*"]
  }
}
`
