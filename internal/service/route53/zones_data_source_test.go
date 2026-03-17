// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ZonesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	zoneName := acctest.RandomDomainName()
	dataSourceName := "data.aws_route53_zones.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccZonesDataSourceConfig_basic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "ids.#", 1),
				),
			},
		},
	})
}

func testAccZonesDataSourceConfig_basic(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "%[1]s."
}

data "aws_route53_zones" "test" {
  depends_on = [aws_route53_zone.test]
}
`, zoneName)
}
