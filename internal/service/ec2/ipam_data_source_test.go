// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_ipam.test"
	dataSourceName := "data.aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMDataSourceConfig_optionsBasic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_private_gua", resourceName, "enable_private_gua"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tier", resourceName, "tier"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccIPAMDataSource_withTier(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_ipam.test"
	dataSourceName := "data.aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMDataSourceConfig_optionsBasic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_private_gua", resourceName, "enable_private_gua"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tier", resourceName, "tier"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

var testAccIPAMDataSourceConfig_optionsBasic = acctest.ConfigCompose(testAccIPAMConfig_tier("free"), `
data "aws_vpc_ipam" "test" {
  ipam_id = aws_vpc_ipam.test.id
}
`)
