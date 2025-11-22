// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2AccountDataSource_default(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_sesv2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SESEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataSourceConfig_default(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "dedicated_ip_auto_warmup_enabled"),
					resource.TestCheckResourceAttrSet(dataSourceName, "enforcement_status"),
					resource.TestCheckResourceAttrSet(dataSourceName, "production_access_enabled"),
					resource.TestCheckResourceAttrSet(dataSourceName, "send_quota.max_24_hour_send"),
					resource.TestCheckResourceAttrSet(dataSourceName, "send_quota.max_send_rate"),
					resource.TestCheckResourceAttrSet(dataSourceName, "send_quota.sent_last_24_hours"),
					resource.TestCheckResourceAttrSet(dataSourceName, "sending_enabled"),
					resource.TestCheckResourceAttrSet(dataSourceName, "suppression_attributes.suppressed_reasons.#"),
				),
			},
		},
	})
}

func TestAccSESV2AccountDataSource_vdmAttributes(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_sesv2_account.test"
	resourceName := "aws_sesv2_account_vdm_attributes.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SESEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataSourceConfig_vdmAttributes(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "vdm_attributes.vdm_enabled", resourceName, "vdm_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vdm_attributes.dashboard_attributes.engagement_metrics", resourceName, "dashboard_attributes.0.engagement_metrics"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vdm_attributes.guardian_attributes.optimized_shared_delivery", resourceName, "guardian_attributes.0.optimized_shared_delivery"),
				),
			},
		},
	})
}

func TestAccSESV2AccountDataSource_suppressionAttributes(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_sesv2_account.test"
	resourceName := "aws_sesv2_account_suppression_attributes.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SESEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataSourceConfig_suppressionAttributes(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "suppression_attributes.suppressed_reasons", resourceName, "suppressed_reasons"),
				),
			},
		},
	})
}

func testAccAccountDataSourceConfig_default() string {
	return `
data "aws_sesv2_account" "test" {
}
`
}

func testAccAccountDataSourceConfig_vdmAttributes() string {
	return `
resource "aws_sesv2_account_vdm_attributes" "test" {
  vdm_enabled = "ENABLED"

  dashboard_attributes {
    engagement_metrics = "ENABLED"
  }

  guardian_attributes {
    optimized_shared_delivery = "ENABLED"
  }
}

data "aws_sesv2_account" "test" {
  depends_on = [ aws_sesv2_account_vdm_attributes.test ]
}
`
}

func testAccAccountDataSourceConfig_suppressionAttributes() string {
	return `
resource "aws_sesv2_account_suppression_attributes" "test" {
  suppressed_reasons = ["COMPLAINT"]
}

data "aws_sesv2_account" "test" {
  depends_on = [ aws_sesv2_account_suppression_attributes.test ]
}
`
}
