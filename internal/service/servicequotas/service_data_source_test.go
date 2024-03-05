// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceQuotasServiceDataSource_serviceName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_servicequotas_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_name("Amazon Virtual Private Cloud (Amazon VPC)"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "service_code", "vpc"),
				),
			},
		},
	})
}

func testAccServiceDataSourceConfig_name(serviceName string) string {
	return fmt.Sprintf(`
data "aws_servicequotas_service" "test" {
  service_name = %[1]q
}
`, serviceName)
}
