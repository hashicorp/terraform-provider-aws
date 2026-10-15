// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTemplatesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_servicequotas_templates.test"
	regionDataSourceName := "data.aws_region.current"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplatesDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "aws_region", regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "templates.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "templates.0.quota_code", lambdaStorageQuotaCode),
					resource.TestCheckResourceAttr(dataSourceName, "templates.0.service_code", lambdaServiceCode),
					resource.TestCheckResourceAttr(dataSourceName, "templates.0.value", lambdaStorageValue),
				),
			},
		},
	})
}

func testAccTemplatesDataSource_region(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_servicequotas_templates.test"
	regionDataSourceName := "data.aws_region.current"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplatesDataSourceConfig_region(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrRegion, regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "templates.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "templates.0.quota_code", lambdaStorageQuotaCode),
					resource.TestCheckResourceAttr(dataSourceName, "templates.0.service_code", lambdaServiceCode),
					resource.TestCheckResourceAttr(dataSourceName, "templates.0.value", lambdaStorageValue),
				),
			},
		},
	})
}

func testAccTemplatesDataSourceConfig_basic() string {
	return acctest.ConfigCompose(testAccTemplateConfig_basic(lambdaStorageQuotaCode, lambdaServiceCode, lambdaStorageValue), `
data "aws_servicequotas_templates" "test" {
  aws_region = aws_servicequotas_template.test.region
}
`)
}

func testAccTemplatesDataSourceConfig_region() string {
	return acctest.ConfigCompose(testAccTemplateConfig_basic(lambdaStorageQuotaCode, lambdaServiceCode, lambdaStorageValue), `
data "aws_servicequotas_templates" "test" {
  region = aws_servicequotas_template.test.region
}
`)
}
