// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceQuotasServiceQuotasDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	const dataSourceName = "data.aws_servicequotas_service_quotas.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotasDataSourceConfig_basic("vpc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "service_code", "vpc"),
					resource.TestCheckResourceAttrSet(dataSourceName, "quotas.#"),
				),
			},
		},
	})
}

func TestAccServiceQuotasServiceQuotasDataSource_quotaAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	const dataSourceName = "data.aws_servicequotas_service_quotas.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotasDataSourceConfig_basic("vpc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "quotas.0.quota_code"),
					resource.TestCheckResourceAttrSet(dataSourceName, "quotas.0.quota_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "quotas.0.arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "quotas.0.value"),
					resource.TestCheckResourceAttrSet(dataSourceName, "quotas.0.default_value"),
					resource.TestCheckResourceAttrSet(dataSourceName, "quotas.0.service_code"),
					resource.TestCheckResourceAttrSet(dataSourceName, "quotas.0.service_name"),
				),
			},
		},
	})
}

func testAccServiceQuotasDataSourceConfig_basic(serviceCode string) string { // nosemgrep:ci.servicequotas-in-func-name
	return testAccServiceQuotasDataSourceConfig(serviceCode)
}

func testAccServiceQuotasDataSourceConfig(serviceCode string) string { // nosemgrep:ci.servicequotas-in-func-name
	return fmt.Sprintf(`
data "aws_servicequotas_service_quotas" "test" {
  service_code = %[1]q
}
`, serviceCode)
}
