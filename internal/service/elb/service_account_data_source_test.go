// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
)

func TestAccELBServiceAccountDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	expectedAccountID := tfelb.AccountIdPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_elb_service_account.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

func TestAccELBServiceAccountDataSource_region(t *testing.T) {
	ctx := acctest.Context(t)
	expectedAccountID := tfelb.AccountIdPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_elb_service_account.regional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountDataSourceConfig_explicitRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

const testAccServiceAccountDataSourceConfig_basic = `
data "aws_elb_service_account" "main" {}
`

const testAccServiceAccountDataSourceConfig_explicitRegion = `
data "aws_region" "current" {}

data "aws_elb_service_account" "regional" {
  region = data.aws_region.current.name
}
`
