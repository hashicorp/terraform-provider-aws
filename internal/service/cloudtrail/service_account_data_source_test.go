// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudTrailServiceAccountDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	expectedAccountID := tfcloudtrail.ServiceAccountPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_cloudtrail_service_account.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, names.AttrARN, expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

func TestAccCloudTrailServiceAccountDataSource_region(t *testing.T) {
	ctx := acctest.Context(t)
	expectedAccountID := tfcloudtrail.ServiceAccountPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_cloudtrail_service_account.regional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountDataSourceConfig_region,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, names.AttrARN, expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

const testAccServiceAccountDataSourceConfig_basic = `
data "aws_cloudtrail_service_account" "main" {}
`

const testAccServiceAccountDataSourceConfig_region = `
data "aws_region" "current" {}

data "aws_cloudtrail_service_account" "regional" {
  region = data.aws_region.current.name
}
`
