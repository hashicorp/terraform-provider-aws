// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTemplatesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_servicequotas_templates.test"
	regionDataSourceName := "data.aws_region.current"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplatesDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrRegion, regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "templates.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccTemplatesDataSourceConfig_basic() string {
	return acctest.ConfigCompose(
		testAccTemplateConfig_basic(lambdaStorageQuotaCode, lambdaServiceCode, lambdaStorageValue),
		`
data "aws_servicequotas_templates" "test" {
  region = aws_servicequotas_template.test.region
}
`)
}
