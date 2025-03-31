// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
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
	resourceName := "aws_route53_records_exclusive.test"
	zoneResourceName := "aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsExclusiveConfig_multiple(zoneName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", zoneResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "18"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// This check is not exhaustive of all resource_record_set blocks, but
					// attempts to cover a variety of record types and optional arguments
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("resource_record_set"),
						knownvalue.SetPartial([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringRegexp(regexache.MustCompile("a." + zoneName.String())),
								names.AttrType: knownvalue.StringExact(string(types.RRTypeA)),
								"ttl":          knownvalue.Int64Exact(30),
								"resource_records": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										names.AttrValue: knownvalue.StringExact("127.0.0.1"),
									}),
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										names.AttrValue: knownvalue.StringExact("127.0.0.27"),
									}),
								}),
							}),
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName:   knownvalue.StringRegexp(regexache.MustCompile("alias." + zoneName.String())),
								names.AttrType:   knownvalue.StringExact(string(types.RRTypeA)),
								"set_identifier": knownvalue.StringExact("live"),
								"ttl":            knownvalue.Null(), // Omitted for alias target records
								names.AttrWeight: knownvalue.Int64Exact(90),
								"alias_target": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"evaluate_target_health": knownvalue.Bool(true),
									}),
								}),
							}),
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName:   knownvalue.StringRegexp(regexache.MustCompile("alias." + zoneName.String())),
								names.AttrType:   knownvalue.StringExact(string(types.RRTypeA)),
								"set_identifier": knownvalue.StringExact("dev"),
								"ttl":            knownvalue.Null(), // Omitted for alias target records
								names.AttrWeight: knownvalue.Int64Exact(10),
								"alias_target": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"evaluate_target_health": knownvalue.Bool(true),
									}),
								}),
							}),
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName:   knownvalue.StringRegexp(regexache.MustCompile("failover." + zoneName.String())),
								names.AttrType:   knownvalue.StringExact(string(types.RRTypeCname)),
								"ttl":            knownvalue.Int64Exact(5),
								"set_identifier": knownvalue.StringExact("failover-primary"),
								"failover":       knownvalue.StringExact(string(types.ResourceRecordSetFailoverPrimary)),
							}),
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName:   knownvalue.StringRegexp(regexache.MustCompile("failover." + zoneName.String())),
								names.AttrType:   knownvalue.StringExact(string(types.RRTypeCname)),
								"ttl":            knownvalue.Int64Exact(5),
								"set_identifier": knownvalue.StringExact("failover-secondary"),
								"failover":       knownvalue.StringExact(string(types.ResourceRecordSetFailoverSecondary)),
							}),
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName:   knownvalue.StringRegexp(regexache.MustCompile("regional." + zoneName.String())),
								names.AttrType:   knownvalue.StringExact(string(types.RRTypeCname)),
								"ttl":            knownvalue.Int64Exact(5),
								"set_identifier": knownvalue.StringExact("ue1"),
								names.AttrRegion: knownvalue.StringExact(string(types.ResourceRecordSetRegionUsEast1)),
							}),
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName:   knownvalue.StringRegexp(regexache.MustCompile("regional." + zoneName.String())),
								names.AttrType:   knownvalue.StringExact(string(types.RRTypeCname)),
								"ttl":            knownvalue.Int64Exact(5),
								"set_identifier": knownvalue.StringExact("uw2"),
								names.AttrRegion: knownvalue.StringExact(string(types.ResourceRecordSetRegionUsWest2)),
							}),
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringRegexp(regexache.MustCompile("txt." + zoneName.String())),
								names.AttrType: knownvalue.StringExact(string(types.RRTypeTxt)),
								"ttl":          knownvalue.Int64Exact(30),
							}),
						}),
					),
				},
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
					"resource_record_set",
				},
			},
		},
	})
}

// Modification of the ttl attribute should trigger an upsert change request, rather
// than one delete and one create. The test check cannot verify this as the
// determination of whether to upsert the change is an internal detail of the resource
// implementation, but this test can be run with debug logging (`TF_LOG=debug`) to
// confirm the expected behavior by grepping for the `UPSERT` change action in the
// resulting output.
func TestAccRoute53RecordsExclusive_upsert(t *testing.T) {
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
				Config: testAccRecordsExclusiveConfig_upsert(zoneName.String(), recordName.String(), "30"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", zoneResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("resource_record_set"),
						knownvalue.SetPartial([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringRegexp(regexache.MustCompile(recordName.String())),
								"ttl":          knownvalue.Int64Exact(30),
							}),
						}),
					),
				},
			},
			{
				Config: testAccRecordsExclusiveConfig_upsert(zoneName.String(), recordName.String(), "20"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", zoneResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("resource_record_set"),
						knownvalue.SetPartial([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringRegexp(regexache.MustCompile(recordName.String())),
								"ttl":          knownvalue.Int64Exact(20),
							}),
						}),
					),
				},
			},
			{
				Config: testAccRecordsExclusiveConfig_upsert(zoneName.String(), recordName.String(), "30"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", zoneResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "resource_record_set.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("resource_record_set"),
						knownvalue.SetPartial([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringRegexp(regexache.MustCompile(recordName.String())),
								"ttl":          knownvalue.Int64Exact(30),
							}),
						}),
					),
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
		timeout := time.Minute * 5
		if _, err := tfroute53.WaitChangeInsync(ctx, conn, aws.ToString(out.ChangeInfo.Id), timeout); err != nil {
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

func testAccRecordsExclusiveConfig_multiple(zoneName string) string {
	// lintignore:AWSAT003
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_route53_zone" "test" {
  name          = %[1]q
  force_destroy = true
}

resource "aws_route53_health_check" "test" {
  fqdn              = "health.%[1]s"
  port              = 80
  type              = "HTTP"
  resource_path     = "/"
  failure_threshold = "2"
  request_interval  = "30"

  tags = {
    Name = "tf-test-health-check"
  }
}

resource "aws_elb" "test_live" {
  name               = "test-elb-live"
  availability_zones = slice(data.aws_availability_zones.available.names, 0, 1)

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_elb" "test_dev" {
  name               = "test-elb-dev"
  availability_zones = slice(data.aws_availability_zones.available.names, 0, 1)

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_route53_records_exclusive" "test" {
  zone_id = aws_route53_zone.test.zone_id

  ### A record
  resource_record_set {
    name = "a.%[1]s"
    type = "A"
    ttl  = "30"

    resource_records {
      value = "127.0.0.1"
    }
    resource_records {
      value = "127.0.0.27"
    }
  }

  ### Alias target
  resource_record_set {
    name           = "alias.%[1]s"
    type           = "A"
    set_identifier = "live"

    weight = 90
    alias_target {
      hosted_zone_id         = aws_elb.test_live.zone_id
      dns_name               = aws_elb.test_live.dns_name
      evaluate_target_health = true
    }
  }

  resource_record_set {
    name           = "alias.%[1]s"
    type           = "A"
    set_identifier = "dev"

    weight = 10
    alias_target {
      hosted_zone_id         = aws_elb.test_dev.zone_id
      dns_name               = aws_elb.test_dev.dns_name
      evaluate_target_health = true
    }
  }

  ### CAA record
  resource_record_set {
    name = "caa.%[1]s"
    type = "CAA"
    ttl  = "30"

    resource_records {
      value = "0 issue \"domainca.test;\""
    }
  }

  ### DS record
  resource_record_set {
    name = "ds.%[1]s"
    type = "DS"
    ttl  = "30"

    resource_records {
      value = "123 4 5 1234567890ABCDEF1234567890ABCDEF"
    }
  }

  ### Healthcheck
  resource_record_set {
    name           = "healthcheck.%[1]s"
    type           = "A"
    set_identifier = "test"
    ttl            = "5"

    health_check_id = aws_route53_health_check.test.id
    weight          = 1
    resource_records {
      value = "127.0.0.1"
    }
  }

  ### Failover
  resource_record_set {
    name           = "failover.%[1]s"
    type           = "CNAME"
    set_identifier = "failover-primary"
    ttl            = "5"
    failover       = "PRIMARY"

    health_check_id = aws_route53_health_check.test.id
    resource_records {
      value = "primary.%[1]s"
    }
  }

  resource_record_set {
    name           = "failover.%[1]s"
    type           = "CNAME"
    set_identifier = "failover-secondary"
    ttl            = "5"
    failover       = "SECONDARY"

    resource_records {
      value = "secondary.%[1]s"
    }
  }

  ### Geolocation
  resource_record_set {
    name           = "geolocation.%[1]s"
    type           = "A"
    set_identifier = "geolocation1"
    ttl            = "30"

    geolocation {
      continent_code = "NA"
    }
    resource_records {
      value = "127.0.0.1"
    }
  }

  resource_record_set {
    name           = "geolocation.%[1]s"
    type           = "A"
    set_identifier = "geolocation2"
    ttl            = "30"

    geolocation {
      continent_code = "SA"
    }
    resource_records {
      value = "127.0.0.2"
    }
  }

  ### Geoproximity location
  resource_record_set {
    name           = "geoproximity.%[1]s"
    type           = "A"
    set_identifier = "geoproximity1"
    ttl            = "30"

    geoproximity_location {
      bias = -10
      coordinates {
        latitude  = "38.79"
        longitude = "106.53"
      }
    }
    resource_records {
      value = "127.0.0.1"
    }
  }

  resource_record_set {
    name           = "geoproximity.%[1]s"
    type           = "A"
    set_identifier = "geoproximity2"
    ttl            = "30"

    geoproximity_location {
      bias = 10
      coordinates {
        latitude  = "56.13"
        longitude = "106.34"
      }
    }
    resource_records {
      value = "127.0.0.2"
    }
  }

  ### Multi-value answer
  resource_record_set {
    name           = "multivalueanswer.%[1]s"
    type           = "A"
    set_identifier = "multivalue1"
    ttl            = "30"

    multi_value_answer = true
    resource_records {
      value = "127.0.0.1"
    }
  }

  resource_record_set {
    name           = "multivalueanswer.%[1]s"
    type           = "A"
    set_identifier = "multivalue2"
    ttl            = "30"

    multi_value_answer = true
    resource_records {
      value = "127.0.0.2"
    }
  }

  ### Regional
  resource_record_set {
    name           = "regional.%[1]s"
    type           = "CNAME"
    set_identifier = "ue1"
    ttl            = "5"

    region = "us-east-1"
    resource_records {
      value = "regional-ue1.%[1]s"
    }
  }

  resource_record_set {
    name           = "regional.%[1]s"
    type           = "CNAME"
    set_identifier = "uw2"
    ttl            = "5"

    region = "us-west-2"
    resource_records {
      value = "regional-uw2.%[1]s"
    }
  }

  ### SPF record
  resource_record_set {
    name = "spf.%[1]s"
    type = "SPF"
    ttl  = "30"

    resource_records {
      value = "\"include:%[1]s\""
    }
  }

  ### TXT record
  resource_record_set {
    name = "txt.%[1]s"
    type = "TXT"
    ttl  = "30"

    resource_records {
      value = "\"foobar\""
    }
  }
}
`, zoneName)
}

func testAccRecordsExclusiveConfig_upsert(zoneName, recordName, ttl string) string {
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
    ttl  = %[3]q

    resource_records {
      value = "127.0.0.1"
    }
  }
}
`, zoneName, recordName, ttl)
}
