// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53RecordsExclusive_name_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t)
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveNameConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveNameExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", zoneResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, recordName.String()),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_record_set.*", map[string]string{
						names.AttrType:       string(types.RRTypeA),
						"ttl":                "300",
						"resource_records.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_record_set.*", map[string]string{
						names.AttrType:       string(types.RRTypeMx),
						"ttl":                "300",
						"resource_records.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_record_set.*", map[string]string{
						names.AttrType:       string(types.RRTypeTxt),
						"ttl":                "300",
						"resource_records.#": "1",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccRecordsExclusiveNameImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "zone_id",
				// The custom type will handle suppressing differences like trailing periods
				// and case sensitivity, but the initial import will still flag the difference.
				ImportStateVerifyIgnore: []string{
					"resource_record_set.0.name",
					"resource_record_set.1.name",
					"resource_record_set.2.name",
				},
			},
		},
	})
}

func TestAccRoute53RecordsExclusive_name_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t)
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveNameConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveNameExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", zoneResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "3"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccRecordsExclusiveNameConfig_update(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveNameExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_record_set.*", map[string]string{
						names.AttrType: string(types.RRTypeA),
						"ttl":          "60",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_record_set.*", map[string]string{
						names.AttrType: string(types.RRTypeAaaa),
						"ttl":          "300",
					}),
					testAccCheckRecordsExclusiveNameRecordAbsent(ctx, t, zoneResourceName, recordName.String(), string(types.RRTypeMx)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccRoute53RecordsExclusive_name_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var zone route53.GetHostedZoneOutput
	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t)
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"

	addRecord := types.ResourceRecordSet{
		Type: types.RRTypeAaaa,
		Name: aws.String(recordName.String()),
		TTL:  aws.Int64(300),
		ResourceRecords: []types.ResourceRecord{
			{
				Value: aws.String("2001:db8::1"),
			},
		},
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveNameConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveNameExists(ctx, t, resourceName),
					testAccCheckZoneExists(ctx, t, zoneResourceName, &zone),
					testAccCheckRecordsExclusiveChangeRecord(ctx, t, &zone, types.ChangeActionCreate, &addRecord),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRecordsExclusiveNameConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveNameExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "3"),
					testAccCheckRecordsExclusiveNameRecordAbsent(ctx, t, zoneResourceName, recordName.String(), string(types.RRTypeAaaa)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccRoute53RecordsExclusive_name_preservesOtherNames(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var otherRecord types.ResourceRecordSet
	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t)
	otherRecordName := zoneName.RandomSubdomain(t)
	resourceName := "aws_route53_records_exclusive.test"
	otherResourceName := "aws_route53_record.other"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveNameConfig_preservesOtherNames(zoneName.String(), recordName.String(), otherRecordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveNameExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "3"),
					testAccCheckRecordExists(ctx, t, otherResourceName, &otherRecord),
				),
			},
		},
	})
}

func TestAccRoute53RecordsExclusive_name_empty(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var otherRecord types.ResourceRecordSet
	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t)
	otherRecordName := zoneName.RandomSubdomain(t)
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"
	otherResourceName := "aws_route53_record.other"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveNameConfig_empty(zoneName.String(), recordName.String(), otherRecordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveNameExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "0"),
					testAccCheckRecordsExclusiveNameRecordAbsent(ctx, t, zoneResourceName, recordName.String(), string(types.RRTypeA)),
					testAccCheckRecordExists(ctx, t, otherResourceName, &otherRecord),
				),
				// The scoped exclusive resource removes aws_route53_record.same, resulting in
				// a non-empty plan for that record on the post-apply check.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRecordsExclusiveNameExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Route53, create.ErrActionCheckingExistence, tfroute53.ResNameRecordsExclusive, name, errors.New("not found"))
		}

		zoneID := rs.Primary.Attributes["zone_id"]
		if zoneID == "" {
			return create.Error(names.Route53, create.ErrActionCheckingExistence, tfroute53.ResNameRecordsExclusive, name, errors.New("zone_id not set"))
		}

		recordName := rs.Primary.Attributes[names.AttrName]
		if recordName == "" {
			return create.Error(names.Route53, create.ErrActionCheckingExistence, tfroute53.ResNameRecordsExclusive, name, errors.New("name not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)
		out, err := tfroute53.FindResourceRecordSetsForHostedZoneByName(ctx, conn, zoneID, recordName)
		if err != nil {
			return create.Error(names.Route53, create.ErrActionCheckingExistence, tfroute53.ResNameRecordsExclusive, zoneID, err)
		}

		recordCount := rs.Primary.Attributes["resource_record_set.#"]
		if recordCount != strconv.Itoa(len(out)) {
			return create.Error(names.Route53, create.ErrActionCheckingExistence, tfroute53.ResNameRecordsExclusive, zoneID, errors.New("unexpected resource_record_set count"))
		}

		return nil
	}
}

func testAccCheckRecordsExclusiveNameRecordAbsent(ctx context.Context, t *testing.T, zoneResourceName, recordName, recordType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[zoneResourceName]
		if !ok {
			return create.Error(names.Route53, create.ErrActionCheckingExistence, tfroute53.ResNameRecordsExclusive, zoneResourceName, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)
		_, _, err := tfroute53.FindResourceRecordSetByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrID], recordName, recordType, "")
		if retry.NotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		return create.Error(names.Route53, create.ErrActionCheckingDestroyed, tfroute53.ResNameRecordsExclusive, recordName, errors.New("not destroyed"))
	}
}

func testAccRecordsExclusiveNameImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s%s%s", rs.Primary.Attributes["zone_id"], intflex.ResourceIdSeparator, rs.Primary.Attributes[names.AttrName]), nil
	}
}

func testAccRecordsExclusiveNameConfig_basic(zoneName, recordName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name          = %[1]q
  force_destroy = true
}

resource "aws_route53_records_exclusive" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[2]q

  resource_record_set {
    name = %[2]q
    type = "A"
    ttl  = 300

    resource_records {
      value = "192.0.2.10"
    }
  }

  resource_record_set {
    name = %[2]q
    type = "MX"
    ttl  = 300

    resource_records {
      value = "10 mail.%[1]s."
    }
  }

  resource_record_set {
    name = %[2]q
    type = "TXT"
    ttl  = 300

    resource_records {
      value = "\"some-app-config\""
    }
  }
}
`, zoneName, recordName)
}

func testAccRecordsExclusiveNameConfig_update(zoneName, recordName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name          = %[1]q
  force_destroy = true
}

resource "aws_route53_records_exclusive" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[2]q

  resource_record_set {
    name = %[2]q
    type = "A"
    ttl  = 60

    resource_records {
      value = "192.0.2.20"
    }
  }

  resource_record_set {
    name = %[2]q
    type = "AAAA"
    ttl  = 300

    resource_records {
      value = "2001:db8::10"
    }
  }

  resource_record_set {
    name = %[2]q
    type = "TXT"
    ttl  = 300

    resource_records {
      value = "\"some-app-config\""
    }
  }
}
`, zoneName, recordName)
}

func testAccRecordsExclusiveNameConfig_preservesOtherNames(zoneName, recordName, otherRecordName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name          = %[1]q
  force_destroy = true
}

resource "aws_route53_record" "other" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[3]q
  type    = "A"
  ttl     = 300
  records = ["192.0.2.30"]
}

resource "aws_route53_records_exclusive" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[2]q

  depends_on = [aws_route53_record.other]

  resource_record_set {
    name = %[2]q
    type = "A"
    ttl  = 300

    resource_records {
      value = "192.0.2.10"
    }
  }

  resource_record_set {
    name = %[2]q
    type = "MX"
    ttl  = 300

    resource_records {
      value = "10 mail.%[1]s."
    }
  }

  resource_record_set {
    name = %[2]q
    type = "TXT"
    ttl  = 300

    resource_records {
      value = "\"some-app-config\""
    }
  }
}
`, zoneName, recordName, otherRecordName)
}

func testAccRecordsExclusiveNameConfig_empty(zoneName, recordName, otherRecordName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name          = %[1]q
  force_destroy = true
}

resource "aws_route53_record" "same" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[2]q
  type    = "A"
  ttl     = 300
  records = ["192.0.2.10"]
}

resource "aws_route53_record" "other" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[3]q
  type    = "A"
  ttl     = 300
  records = ["192.0.2.30"]
}

resource "aws_route53_records_exclusive" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[2]q

  depends_on = [
    aws_route53_record.same,
    aws_route53_record.other,
  ]
}
`, zoneName, recordName, otherRecordName)
}
