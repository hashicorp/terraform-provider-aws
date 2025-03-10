// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53RecordsExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", zoneResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_record_set.*", map[string]string{
						names.AttrType:       string(types.RRTypeA),
						"ttl":                "30",
						"resource_records.#": "2",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccRecordsExclusiveImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "zone_id",
				// The custom type will handle suppressing differences like trailing periods
				// and case sensitivity, but the initial import will still flag the difference.
				ImportStateVerifyIgnore: []string{"resource_record_set.0.name"},
			},
		},
	})
}

func TestAccRoute53RecordsExclusive_disappears_Zone(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceZone(), zoneResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53RecordsExclusive_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	zoneName := acctest.RandomDomain()
	recordName1 := zoneName.RandomSubdomain()
	recordName2 := zoneName.RandomSubdomain()
	recordName3 := zoneName.RandomSubdomain()
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveConfig_multiple(zoneName.String(), recordName1.String(), recordName2.String(), recordName3.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", zoneResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "3"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccRecordsExclusiveImportStateIdFunc(resourceName),
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

// An empty resource_record_set argument should remove any existing, non-default record
// sets associated with the zone
func TestAccRoute53RecordsExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()
	resourceName := "aws_route53_records_exclusive.test"
	recordResourceName := "aws_route53_record.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveConfig_empty(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("resource_record_set"),
						knownvalue.SetExact([]knownvalue.Check{}),
					),
				},
				// The _exclusive resource will remove the record set created by the _record resource,
				// resulting in a non-empty plan.
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRecordsExclusiveConfig_empty(zoneName.String(), recordName.String()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(recordResourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// A record set added out of band should be removed
func TestAccRoute53RecordsExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var zone route53.GetHostedZoneOutput
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()
	recordName2 := zoneName.RandomSubdomain()
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"

	addRecord := types.ResourceRecordSet{
		Type: types.RRTypeA,
		Name: aws.String(recordName2.String()),
		TTL:  aws.Int64(30),
		ResourceRecords: []types.ResourceRecord{
			{
				Value: aws.String("127.0.0.1"),
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					testAccCheckZoneExists(ctx, zoneResourceName, &zone),
					testAccCheckRecordsExclusiveChangeRecord(ctx, &zone, types.ChangeActionCreate, &addRecord),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRecordsExclusiveConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "1"),
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

// A record set removed out of band should be re-created
func TestAccRoute53RecordsExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var zone route53.GetHostedZoneOutput
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"

	// Must match the _basic test record entry exactly
	removeRecord := types.ResourceRecordSet{
		Type: types.RRTypeA,
		Name: aws.String(recordName.String()),
		TTL:  aws.Int64(30),
		ResourceRecords: []types.ResourceRecord{
			{
				Value: aws.String("127.0.0.1"),
			},
			{
				Value: aws.String("127.0.0.27"),
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					testAccCheckZoneExists(ctx, zoneResourceName, &zone),
					testAccCheckRecordsExclusiveChangeRecord(ctx, &zone, types.ChangeActionDelete, &removeRecord),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRecordsExclusiveConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "1"),
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

// A record set updated out of band should be upserted with the configured definition
func TestAccRoute53RecordsExclusive_outOfBandUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var zone route53.GetHostedZoneOutput
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"

	// Name and Type must match the _basic test record entry exactly
	updateRecord := types.ResourceRecordSet{
		Type: types.RRTypeA,
		Name: aws.String(recordName.String()),
		TTL:  aws.Int64(60), // changed from 30
		ResourceRecords: []types.ResourceRecord{
			{
				Value: aws.String("127.0.0.1"),
			},
			{
				Value: aws.String("127.0.0.27"),
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					testAccCheckZoneExists(ctx, zoneResourceName, &zone),
					testAccCheckRecordsExclusiveChangeRecord(ctx, &zone, types.ChangeActionUpsert, &updateRecord),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRecordsExclusiveConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "1"),
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

func testAccCheckRecordsExclusiveDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_records_exclusive" {
				continue
			}

			zoneID := rs.Primary.Attributes["zone_id"]
			_, err := tfroute53.FindResourceRecordSetsForHostedZone(ctx, conn, zoneID)
			if errs.IsA[*types.NoSuchHostedZone](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Route53, create.ErrActionCheckingDestroyed, tfroute53.ResNameRecordsExclusive, zoneID, err)
			}

			return create.Error(names.Route53, create.ErrActionCheckingDestroyed, tfroute53.ResNameRecordsExclusive, zoneID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRecordsExclusiveExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Route53, create.ErrActionCheckingExistence, tfroute53.ResNameRecordsExclusive, name, errors.New("not found"))
		}

		zoneID := rs.Primary.Attributes["zone_id"]
		if zoneID == "" {
			return create.Error(names.Route53, create.ErrActionCheckingExistence, tfroute53.ResNameRecordsExclusive, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)
		out, err := tfroute53.FindResourceRecordSetsForHostedZone(ctx, conn, zoneID)
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

// testAccCheckRecordsExclusiveChangeRecord is a test helper used to modify recordm sets out of band
func testAccCheckRecordsExclusiveChangeRecord(ctx context.Context, zone *route53.GetHostedZoneOutput, action types.ChangeAction, record *types.ResourceRecordSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		out, err := conn.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: zone.HostedZone.Id,
			ChangeBatch: &types.ChangeBatch{
				Changes: []types.Change{
					{
						Action:            action,
						ResourceRecordSet: record,
					},
				},
			},
		})
		if err != nil {
			return err
		}
		if out.ChangeInfo == nil || out.ChangeInfo.Id == nil {
			return errors.New("empty change ID")
		}

		// wait for the change to complete
		if _, err := tfroute53.WaitChangeInsync(ctx, conn, aws.ToString(out.ChangeInfo.Id)); err != nil {
			return err
		}

		return nil
	}
}

func testAccRecordsExclusiveImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["zone_id"], nil
	}
}

func testAccRecordsExclusiveConfig_basic(zoneName, recordName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name          = %[1]q
  force_destroy = true
}

resource "aws_route53_records_exclusive" "test" {
  zone_id = aws_route53_zone.test.zone_id

  resource_record_set {
    name = %[2]q
    type = "A"
    ttl  = "30"

    resource_records {
      value = "127.0.0.1"
    }
    resource_records {
      value = "127.0.0.27"
    }
  }
}
`, zoneName, recordName)
}

func testAccRecordsExclusiveConfig_multiple(zoneName, recordName1, recordName2, recordName3 string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name          = %[1]q
  force_destroy = true
}

resource "aws_route53_records_exclusive" "test" {
  zone_id = aws_route53_zone.test.zone_id

  resource_record_set {
    name = %[2]q
    type = "A"
    ttl  = "30"

    resource_records {
      value = "127.0.0.1"
    }
  }

  resource_record_set {
    name = %[3]q
    type = "A"
    ttl  = "30"

    resource_records {
      value = "127.0.0.2"
    }
  }

  resource_record_set {
    name = %[4]q
    type = "A"
    ttl  = "30"

    resource_records {
      value = "127.0.0.3"
    }
  }
}
`, zoneName, recordName1, recordName2, recordName3)
}

func testAccRecordsExclusiveConfig_empty(zoneName, recordName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name          = %[1]q
  force_destroy = true
}

# Create a record set which the _exclusive resource will remove
resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[2]q
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1"]
}

resource "aws_route53_records_exclusive" "test" {
  zone_id = aws_route53_record.test.zone_id
}
`, zoneName, recordName)
}
