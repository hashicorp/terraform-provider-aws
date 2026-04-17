// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53RecordsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_route53_records.test"
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsDataSourceConfig_basic(zoneName.String(), recordName.String()),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("resource_record_sets"), knownvalue.ListExact([]knownvalue.Check{knownvalue.MapPartial(map[string]knownvalue.Check{
						names.AttrName: knownvalue.StringExact(recordName.String() + "."),
						"resource_records": knownvalue.ListExact([]knownvalue.Check{knownvalue.MapExact(map[string]knownvalue.Check{
							names.AttrValue: knownvalue.StringExact("127.0.0.1"),
						})}),
						"ttl":          knownvalue.Int64Exact(30),
						names.AttrType: tfknownvalue.StringExact(awstypes.RRTypeA),
					})})),
				},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/46426.
func TestAccRoute53RecordsDataSource_wildcardRegexInvalid(t *testing.T) {
	ctx := acctest.Context(t)
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRecordsDataSourceConfig_wildcardRegexInvalid(zoneName.String(), recordName.String()),
				ExpectError: regexache.MustCompile(`Invalid Regexp Value`),
			},
		},
	})
}

func TestAccRoute53RecordsDataSource_wildcardRegex(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_route53_records.test"
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsDataSourceConfig_wildcardRegex(zoneName.String(), recordName.String()),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("resource_record_sets"), knownvalue.Null()),
				},
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

func testAccRecordsDataSourceConfig_wildcardRegexInvalid(zName, rName string) string {
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
  name_regex = "*.%[1]s"
}
`, zName, rName)
}

func testAccRecordsDataSourceConfig_wildcardRegex(zName, rName string) string {
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
  name_regex = "\\*\\.%[1]s"
}
`, zName, rName)
}
