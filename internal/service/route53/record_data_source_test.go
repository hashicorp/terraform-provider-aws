// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53RecordDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// resourceName := "aws_route53_record.test"
	dataSourceName := "data.aws_route53_record.test"
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordDataSourceConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "filter_record", recordName.String()),
					resource.TestCheckResourceAttr(dataSourceName, "record_sets.0.name", fmt.Sprintf("%s.", recordName.String())),
					resource.TestCheckResourceAttr(dataSourceName, "record_sets.0.ttl", "30"),
					resource.TestCheckResourceAttr(dataSourceName, "record_sets.0.type", "A"),
					resource.TestCheckResourceAttr(dataSourceName, "record_sets.0.values.0", "127.0.0.1"),
				),
			},
		},
	})
}

func TestAccRoute53RecordDataSource_weightedPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	// resourceName := "aws_route53_record.test"
	dataSourceName := "data.aws_route53_record.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfigDataSource_weightedCNAME,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "record_sets.0.name", "www.domain.test."),
					resource.TestCheckResourceAttr(dataSourceName, "record_sets.0.ttl", "5"),
					resource.TestCheckResourceAttr(dataSourceName, "record_sets.0.type", "CNAME"),
					resource.TestCheckResourceAttr(dataSourceName, "record_sets.0.set_identifier", "dev"),
				),
			},
		},
	})
}

func testAccRecordDataSourceConfig_basic(zName, rName string) string {
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

data "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test.zone_id
  filter_record = aws_route53_record.test.name
}
`, zName, rName)
}

const testAccRecordConfigDataSource_weightedCNAME = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "www-dev" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  weighted_routing_policy {
    weight = 10
  }

  set_identifier = "dev"
  records        = ["dev.domain.test"]
}

resource "aws_route53_record" "www-live" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  weighted_routing_policy {
    weight = 90
  }

  set_identifier = "live"
  records        = ["dev.domain.test"]
}

resource "aws_route53_record" "www-off" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  weighted_routing_policy {
    weight = 0
  }

  set_identifier = "off"
  records        = ["dev.domain.test"]
}

data "aws_route53_record" "test" {
	zone_id = aws_route53_zone.main.zone_id
	filter_record = "^www"
	depends_on = [aws_route53_record.www-dev, aws_route53_record.www-live, aws_route53_record.www-off]
  }
`
