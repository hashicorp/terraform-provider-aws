// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"fmt"
	"regexp"
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

func TestAccRoute53RecordSetProtectionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_route53_record_set_protections.test"
	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t).String()
	// Route 53 returns record names fully qualified.
	recordNameRegex := regexache.MustCompile(`^` + regexp.QuoteMeta(recordName) + `\.?$`)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordSetProtectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordSetProtectionsDataSourceConfig_basic(zoneName.String(), recordName),
				ConfigStateChecks: []statecheck.StateCheck{
					// Route 53 sorts protections by record name, then by record type.
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("protections"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrComment: knownvalue.Null(),
							"record_name":     knownvalue.StringRegexp(recordNameRegex),
							"record_type":     tfknownvalue.StringExact(awstypes.RRTypeA),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrComment: knownvalue.StringExact("protected"),
							"record_name":     knownvalue.StringRegexp(recordNameRegex),
							"record_type":     tfknownvalue.StringExact(awstypes.RRTypeTxt),
						}),
					})),
				},
			},
		},
	})
}

func TestAccRoute53RecordSetProtectionsDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_route53_record_set_protections.test"
	zoneName := acctest.RandomDomain(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordSetProtectionsDataSourceConfig_empty(zoneName.String()),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("protections"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func testAccRecordSetProtectionsDataSourceConfig_basic(zoneName, recordName string) string {
	return acctest.ConfigCompose(testAccRecordSetProtectionConfig_base(zoneName), fmt.Sprintf(`
resource "aws_route53_record_set_protection" "a" {
  zone_id     = aws_route53_zone.test.zone_id
  record_name = %[1]q
  record_type = "A"
}

resource "aws_route53_record_set_protection" "txt" {
  zone_id     = aws_route53_zone.test.zone_id
  record_name = %[1]q
  record_type = "TXT"
  comment     = "protected"
}

data "aws_route53_record_set_protections" "test" {
  zone_id = aws_route53_zone.test.zone_id

  depends_on = [
    aws_route53_record_set_protection.a,
    aws_route53_record_set_protection.txt,
  ]
}
`, recordName))
}

func testAccRecordSetProtectionsDataSourceConfig_empty(zoneName string) string {
	return acctest.ConfigCompose(testAccRecordSetProtectionConfig_base(zoneName), `
data "aws_route53_record_set_protections" "test" {
  zone_id = aws_route53_zone.test.zone_id
}
`)
}
