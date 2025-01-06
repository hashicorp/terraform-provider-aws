// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53RecordsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_route53_records.test"
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsDataSourceConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "resource_record_sets.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "resource_record_sets.0.name", recordName.String()+"."),
					resource.TestCheckResourceAttr(dataSourceName, "resource_record_sets.0.ttl", "30"),
					resource.TestCheckResourceAttr(dataSourceName, "resource_record_sets.0.type", "A"),
					resource.TestCheckResourceAttr(dataSourceName, "resource_record_sets.0.resource_records.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "resource_record_sets.0.resource_records.0.value", "127.0.0.1"),
				),
			},
		},
	})
}

func testAccRecordsDataSourceConfig_basic(zName, rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "%[1]s."
}

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[2]q
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1"]
}

data "aws_route53_records" "test" {
  zone_id    = aws_route53_record.test.zone_id
  name_regex = "^%[2]s"
}
`, zName, rName)
}
