// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.Route53ServiceID, testAccErrorCheckSkip)
}

func TestAccRoute53Record_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_basic(zoneName.String(), strings.ToUpper(recordName.String())),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "alias.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "allow_overwrite"),
					resource.TestCheckResourceAttr(resourceName, "cidr_routing_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "failover_routing_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "fqdn", recordName.String()),
					resource.TestCheckResourceAttr(resourceName, "geolocation_routing_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "geoproximity_routing_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "health_check_id", ""),
					resource.TestCheckResourceAttr(resourceName, "latency_routing_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "multivalue_answer_routing_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, recordName.String()),
					resource.TestCheckResourceAttr(resourceName, "records.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "records.*", "127.0.0.1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "records.*", "127.0.0.27"),
					resource.TestCheckResourceAttr(resourceName, "set_identifier", ""),
					resource.TestCheckResourceAttr(resourceName, "ttl", "30"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "A"),
					resource.TestCheckResourceAttr(resourceName, "weighted_routing_policy.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectAttributeFormat(resourceName, tfjsonpath.New(names.AttrID), "{zone_id}_{name}_{type}"),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Identity_SetIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				Config: testAccRecordConfig_healthCheckIdTypeCNAME(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectAttributeFormat(resourceName, tfjsonpath.New(names.AttrID), "{zone_id}_{name}_{type}_{set_identifier}"),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						"zone_id":           knownvalue.NotNull(),
						names.AttrName:      knownvalue.NotNull(),
						names.AttrType:      knownvalue.NotNull(),
						"set_identifier":    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("zone_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrName)),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrType)),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("set_identifier")),
				},
			},

			// Step 2: Import command
			{
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				ImportStateKind: resource.ImportBlockWithID,
				ResourceName:    resourceName,
				ImportState:     true,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("test")),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.StringExact("CNAME")),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("set_identifier"), knownvalue.StringExact("set-id")),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^[[:alnum:]]+_test_CNAME_set-id$`))),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ResourceName:    resourceName,
				ImportState:     true,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("test")),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.StringExact("CNAME")),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("set_identifier"), knownvalue.StringExact("set-id")),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^[[:alnum:]]+_test_CNAME_set-id$`))),
					},
				},
			},
		},
	})
}

func TestAccRoute53Record_Identity_ChangeOnUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccRecordConfig_setIdentifierRenameWeighted("before"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectAttributeFormat(resourceName, tfjsonpath.New(names.AttrID), "{zone_id}_{name}_{type}_{set_identifier}"),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						"zone_id":           knownvalue.NotNull(),
						names.AttrName:      knownvalue.NotNull(),
						names.AttrType:      knownvalue.NotNull(),
						"set_identifier":    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("zone_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrName)),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrType)),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("set_identifier")),
				},
			},

			// Step 2: Update
			{
				Config: testAccRecordConfig_setIdentifierRenameWeighted("after"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectAttributeFormat(resourceName, tfjsonpath.New(names.AttrID), "{zone_id}_{name}_{type}_{set_identifier}"),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						"zone_id":           knownvalue.NotNull(),
						names.AttrName:      knownvalue.NotNull(),
						names.AttrType:      knownvalue.NotNull(),
						"set_identifier":    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("zone_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrName)),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrType)),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("set_identifier")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccRoute53Record_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfroute53.ResourceRecord(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53Record_Disappears_multipleRecords(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3, v4, v5 awstypes.ResourceRecordSet
	zoneName := acctest.RandomDomain()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_multiple(zoneName.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, "aws_route53_record.test.0", &v1),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.test.1", &v2),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.test.2", &v3),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.test.3", &v4),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.test.4", &v5),
					acctest.CheckSDKResourceDisappears(ctx, t, tfroute53.ResourceRecord(), "aws_route53_record.test.0"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53Record_underscored(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.underscore"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_underscoreInName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_fqdn(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_fqdn,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},

			// Ensure that changing the name to include a trailing "dot" results in
			// nothing happening, because the name is stripped of trailing dots on
			// save. Otherwise, an update would occur and due to the
			// create_before_destroy, the record would actually be destroyed, and a
			// non-empty plan would appear, and the record will fail to exist in
			// testAccCheckRecordExists
			{
				Config: testAccRecordConfig_fqdnNoOp,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

// TestAccRoute53Record_trailingPeriodAndZoneID ensures an aws_route53_record
// created with a name configured with a trailing period and explicit zone_id gets imported correctly
func TestAccRoute53Record_trailingPeriodAndZoneID(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_nameTrailingPeriod,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Support_txt(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_txt,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "zone_id"},
			},
		},
	})
}

func TestAccRoute53Record_Support_spf(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_spf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
					resource.TestCheckTypeSetElemAttr(resourceName, "records.*", "include:domain.test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Support_caa(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_caa,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
					resource.TestCheckTypeSetElemAttr(resourceName, "records.*", "0 issue \"domainca.test;\""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Support_ds(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_ds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_generatesSuffix(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_suffix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_wildcard(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.wildcard"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_wildcard,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("fqdn"), knownvalue.StringExact("*.domain.test")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("*.domain.test")),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRecordConfig_wildcardUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("fqdn"), knownvalue.StringExact("*.domain.test")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("*.domain.test")),
				},
			},
		},
	})
}

func TestAccRoute53Record_failover(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.www-primary"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_failoverCNAME,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.www-secondary", &record2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Weighted_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2, record3 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.www-live"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_weightedCNAME,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, "aws_route53_record.www-dev", &record1),
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.www-off", &record3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_WeightedToSimple_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.www-server1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_weightedRoutingPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_simpleRoutingPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
		},
	})
}

func TestAccRoute53Record_Alias_elb(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.alias"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_aliasELB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Alias_s3(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_record.alias"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_aliasS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Alias_vpcEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_record.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccRecordConfig_aliasCustomVPCEndpointSwappedAliasAttributes(rName),
				ExpectError: regexache.MustCompile(`expected length of`),
			},
			{
				Config: testAccRecordConfig_customVPCEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Alias_uppercase(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.alias"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_aliasELBUppercase(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Weighted_alias(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2, record3, record4, record5, record6 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.elb_weighted_alias_live"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_weightedELBAlias(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.elb_weighted_alias_dev", &record2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},

			{
				Config: testAccRecordConfig_weightedAlias,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, "aws_route53_record.green_origin", &record3),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.r53_weighted_alias_live", &record4),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.blue_origin", &record5),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.r53_weighted_alias_dev", &record6),
				),
			},
		},
	})
}

func TestAccRoute53Record_cidr(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	locationName := sdkacctest.RandString(16)
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_cidr(rName, locationName, zoneName.String(), recordName.String(), "cidr-location-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "alias.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "allow_overwrite"),
					resource.TestCheckResourceAttr(resourceName, "cidr_routing_policy.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_routing_policy.0.collection_id", "aws_route53_cidr_collection.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_routing_policy.0.location_name", "aws_route53_cidr_location.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "failover_routing_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "fqdn", recordName.String()),
					resource.TestCheckResourceAttr(resourceName, "geolocation_routing_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "geoproximity_routing_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "health_check_id", ""),
					resource.TestCheckResourceAttr(resourceName, "latency_routing_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "multivalue_answer_routing_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, recordName.String()),
					resource.TestCheckResourceAttr(resourceName, "records.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "records.*", "2001:0db8::0123:4567:89ab:cdef"),
					resource.TestCheckResourceAttr(resourceName, "set_identifier", "cidr-location-1"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "60"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AAAA"),
					resource.TestCheckResourceAttr(resourceName, "weighted_routing_policy.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_cidr(rName, locationName, zoneName.String(), recordName.String(), "cidr-location-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cidr_routing_policy.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_routing_policy.0.collection_id", "aws_route53_cidr_collection.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_routing_policy.0.location_name", "aws_route53_cidr_location.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "set_identifier", "cidr-location-2"),
				),
			},
			{
				Config: testAccRecordConfig_cidrDefaultLocation(rName, locationName, zoneName.String(), recordName.String(), "cidr-location-3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cidr_routing_policy.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_routing_policy.0.collection_id", "aws_route53_cidr_collection.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cidr_routing_policy.0.location_name", "*"),
					resource.TestCheckResourceAttr(resourceName, "set_identifier", "cidr-location-3"),
				),
			},
		},
	})
}

func TestAccRoute53Record_Geolocation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2, record3, record4 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_geolocationCNAME,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, "aws_route53_record.default", &record1),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.california", &record2),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.oceania", &record3),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.denmark", &record4),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Geoproximity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2, record3 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.awsregion"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_geoproximityCNAME(endpoints.UsEast1RegionID, fmt.Sprintf("%s-atl-1", endpoints.UsEast1RegionID)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, "aws_route53_record.awsregion", &record1),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.localzonegroup", &record2),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.coordinates", &record3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_HealthCheckID_setIdentifierChange(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_healthCheckIdSetIdentifier("test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_healthCheckIdSetIdentifier("test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_HealthCheckID_typeChange(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_healthCheckIdTypeCNAME(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectAttributeFormat(resourceName, tfjsonpath.New(names.AttrID), "{zone_id}_{name}_{type}_{set_identifier}"),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_healthCheckIdTypeA(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectAttributeFormat(resourceName, tfjsonpath.New(names.AttrID), "{zone_id}_{name}_{type}_{set_identifier}"),
				},
			},
		},
	})
}

func TestAccRoute53Record_Latency_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2, record3 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.first_region"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_latencyCNAME(endpoints.UsEast1RegionID, endpoints.EuWest1RegionID, endpoints.ApNortheast1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.second_region", &record2),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.third_region", &record3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_typeChange(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.sample"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_typeChangePre,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.StringExact("CNAME")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("records"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("www.terraform.io"),
					})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_typeChangePost,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("records"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("127.0.0.1"),
						knownvalue.StringExact("8.8.8.8"),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccRoute53Record_nameChange(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.sample"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_nameChangePre,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("sample")),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_nameChangePost,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
					testAccCheckRecordDoesNotExist(ctx, t, "aws_route53_zone.main", "sample", "CNAME"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("sample-new")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccRoute53Record_setIdentifierChangeBasicToWeighted(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.basic_to_weighted"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierChangeBasicToWeightedPre,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRecordConfig_setIdentifierChangeBasicToWeightedPost,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geolocationContinent(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geolocation"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationContinent("AN", "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationContinent("AN", "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geolocationCountryDefault(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geolocation"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationCountry("*", "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationCountry("*", "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geolocationCountrySpecified(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geolocation"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationCountry("US", "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationCountry("US", "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geolocationCountrySubdivision(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geolocation"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationCountrySubdivision("US", "CA", "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationCountrySubdivision("US", "CA", "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geoproximityRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geoproximity"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityRegion(endpoints.UsEast1RegionID, "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityRegion(endpoints.UsEast1RegionID, "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geoproximityLocalZoneGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geoproximity"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityLocalZoneGroup(fmt.Sprintf("%s-atl-1", endpoints.UsEast1RegionID), "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityLocalZoneGroup(fmt.Sprintf("%s-atl-1", endpoints.UsEast1RegionID), "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geoproximityCoordinates(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geoproximity"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityCoordinates("49.22", "-74.01", "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityCoordinates("49.22", "-74.01", "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_failover(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_failover"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameFailover("before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameFailover("after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_latency(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_latency"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameLatency(endpoints.UsEast1RegionID, "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameLatency(endpoints.UsEast1RegionID, "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_multiValueAnswer(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_multivalue_answer"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameMultiValueAnswer("before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameMultiValueAnswer("after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_weighted(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameWeighted("before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameWeighted("after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_Alias_change(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_aliasChangePre(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRecordConfig_aliasChangePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_Alias_changeDualstack(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_aliasChangeDualstackPre(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
					testAccCheckRecordAliasNameDualstack(&record1, true),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			// Cause a change, which will trigger a refresh
			{
				Config: testAccRecordConfig_aliasChangeDualstackPost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record2),
					testAccCheckRecordAliasNameDualstack(&record2, false),
				),
			},
		},
	})
}

func TestAccRoute53Record_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.empty"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_emptyName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/8423
func TestAccRoute53Record_longTXTrecord(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.long_txt"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_longTxt,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_MultiValueAnswer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.www-server1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_multiValueAnswerA,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, "aws_route53_record.www-server1", &record1),
					testAccCheckRecordExists(ctx, t, "aws_route53_record.www-server2", &record2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_Allow_doNotOverwrite(t *testing.T) {
	ctx := acctest.Context(t)
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               testAccRecordOverwriteExpectErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_allowOverwrite(false),
			},
		},
	})
}

func TestAccRoute53Record_Allow_overwrite(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.overwriting"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_allowOverwrite(true),
				Check:  resource.ComposeTestCheckFunc(testAccCheckRecordExists(ctx, t, "aws_route53_record.overwriting", &record1)),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_ttl0(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_ttl(zoneName.String(), strings.ToUpper(recordName.String()), 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ttl", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_ttl(zoneName.String(), strings.ToUpper(recordName.String()), 45),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ttl", "45"),
				),
			},
			{
				Config: testAccRecordConfig_ttl(zoneName.String(), strings.ToUpper(recordName.String()), 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ttl", "0"),
				),
			},
		},
	})
}

func TestAccRoute53Record_aliasWildcardName(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	zoneName := acctest.RandomDomain()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_aliasWildcardName(rName, zoneName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
		},
	})
}

func TestAccRoute53Record_escapedSlash(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	zoneName := "0/24." + acctest.RandomDomain()
	recordName := "0"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_basic(zoneName.String(), recordName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoute53Record_escapedSpace(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	zoneName := "a\\040b." + acctest.RandomDomain()
	recordName := "0\\040to\\0401"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_basic(zoneName.String(), recordName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoute53Record_escapedJustSpace(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	// zone name always needs an escape code if any
	zoneName := "a\\040b." + acctest.RandomDomain()
	// as for record name, r53 API can accept a space as is but will send the escaped version of it back
	recordName := "0 to 1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_basic(zoneName.String(), recordName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, t, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccErrorCheckSkip skips Route53 tests that have error messages indicating unsupported features
func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Operations related to PublicDNS",
		"Regional control plane current does not support",
		"NoSuchHostedZone: The specified hosted zone",
	)
}

func testAccRecordOverwriteExpectErrorCheck(t *testing.T) resource.ErrorCheckFunc {
	return func(err error) error {
		f := acctest.ErrorCheck(t, names.Route53ServiceID)
		err = f(err)

		if err == nil {
			t.Fatalf("Expected an error but got none")
		}

		re := regexache.MustCompile(`Tried to create resource record set \[name='www.domain.test.', type='A'] but it already exists`)
		if !re.MatchString(err.Error()) {
			t.Fatalf("Expected an error with pattern, no match on: %s", err)
		}

		return nil
	}
}

func testAccCheckRecordDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_record" {
				continue
			}

			_, _, err := tfroute53.FindResourceRecordSetByFourPartKey(ctx, conn,
				tfroute53.CleanZoneID(rs.Primary.Attributes["zone_id"]),
				rs.Primary.Attributes[names.AttrName],
				rs.Primary.Attributes[names.AttrType],
				rs.Primary.Attributes["set_identifier"],
			)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route 53 Record %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRecordExists(ctx context.Context, t *testing.T, n string, v *awstypes.ResourceRecordSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		output, _, err := tfroute53.FindResourceRecordSetByFourPartKey(ctx, conn,
			tfroute53.CleanZoneID(rs.Primary.Attributes["zone_id"]),
			rs.Primary.Attributes[names.AttrName],
			rs.Primary.Attributes[names.AttrType],
			rs.Primary.Attributes["set_identifier"],
		)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRecordDoesNotExist(ctx context.Context, t *testing.T, zoneResourceName, recordName, recordType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		rs, ok := s.RootModule().Resources[zoneResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", zoneResourceName)
		}

		zone := rs.Primary.ID
		recordName := tfroute53.ExpandRecordName(recordName, zone)

		_, _, err := tfroute53.FindResourceRecordSetByFourPartKey(ctx, conn, zone, recordName, recordType, "")

		if retry.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Route 53 Record %s still exists", recordName)
	}
}

func testAccCheckRecordAliasNameDualstack(record *awstypes.ResourceRecordSet, expectPrefix bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		alias := record.AliasTarget
		if alias == nil {
			return fmt.Errorf("record has no alias target: %#v", record)
		}
		hasPrefix := strings.HasPrefix(*alias.DNSName, "dualstack.")
		if expectPrefix && !hasPrefix {
			return fmt.Errorf("alias name did not have expected prefix: %#v", alias)
		} else if !expectPrefix && hasPrefix {
			return fmt.Errorf("alias name had unexpected prefix: %#v", alias)
		}
		return nil
	}
}

func testAccRecordConfig_allowOverwrite(allowOverwrite bool) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test."
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www.domain.test"
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1"]
}

resource "aws_route53_record" "overwriting" {
  depends_on = [aws_route53_record.default]

  allow_overwrite = %[1]t
  zone_id         = aws_route53_zone.main.zone_id
  name            = "www.domain.test"
  type            = "A"
  ttl             = "30"
  records         = ["127.0.0.1"]
}
`, allowOverwrite)
}

func testAccRecordConfig_basic(zoneName, recordName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[1]q
}

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[2]q
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]
}
`, zoneName, recordName)
}

func testAccRecordConfig_multiple(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[1]q
}

resource "aws_route53_record" "test" {
  count = 5

  name    = "record${count.index}.${aws_route53_zone.test.name}"
  records = ["127.0.0.${count.index}"]
  ttl     = "30"
  type    = "A"
  zone_id = aws_route53_zone.test.zone_id
}
`, zoneName)
}

const testAccRecordConfig_nameTrailingPeriod = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www.DOmaiN.test."
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]
}
`

const testAccRecordConfig_fqdn = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www.DOmaiN.test"
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]

  lifecycle {
    create_before_destroy = true
  }
}
`

const testAccRecordConfig_fqdnNoOp = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www.DOmaiN.test."
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]

  lifecycle {
    create_before_destroy = true
  }
}
`

const testAccRecordConfig_suffix = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "subdomain"
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]
}
`

const testAccRecordConfig_wildcard = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "subdomain"
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]
}

resource "aws_route53_record" "wildcard" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "*.domain.test"
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1"]
}
`

const testAccRecordConfig_wildcardUpdate = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "subdomain"
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]
}

resource "aws_route53_record" "wildcard" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "*.domain.test"
  type    = "A"
  ttl     = "60"
  records = ["127.0.0.1"]
}
`

const testAccRecordConfig_txt = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = "/hostedzone/${aws_route53_zone.main.zone_id}"
  name    = "subdomain"
  type    = "TXT"
  ttl     = "30"
  records = ["lalalala"]
}
`

const testAccRecordConfig_spf = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "test"
  type    = "SPF"
  ttl     = "30"
  records = ["include:domain.test"]
}
`

const testAccRecordConfig_caa = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "test"
  type    = "CAA"
  ttl     = "30"

  records = ["0 issue \"domainca.test;\""]
}
`

const testAccRecordConfig_ds = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "test"
  type    = "DS"
  ttl     = "30"
  records = ["123 4 5 1234567890ABCDEF1234567890ABCDEF"]
}
`

func testAccRecordConfig_baseCIDR(rName, locationName, zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_cidr_collection" "test" {
  name = %[1]q
}

resource "aws_route53_cidr_location" "test" {
  cidr_collection_id = aws_route53_cidr_collection.test.id
  name               = %[2]q
  cidr_blocks        = ["2001:db8:1234::/48", "203.0.113.0/24"]
}

resource "aws_route53_zone" "test" {
  name = %[3]q
}
`, rName, locationName, zoneName)
}

func testAccRecordConfig_cidr(rName, locationName, zoneName, recordName, setIdentifier string) string {
	return acctest.ConfigCompose(testAccRecordConfig_baseCIDR(rName, locationName, zoneName), fmt.Sprintf(`
resource "aws_route53_record" "test" {
  zone_id        = aws_route53_zone.test.zone_id
  name           = %[1]q
  type           = "AAAA"
  ttl            = "60"
  set_identifier = %[2]q

  cidr_routing_policy {
    collection_id = aws_route53_cidr_collection.test.id
    location_name = aws_route53_cidr_location.test.name
  }

  records = ["2001:0db8::0123:4567:89ab:cdef"]
}
`, recordName, setIdentifier))
}

func testAccRecordConfig_cidrDefaultLocation(rName, locationName, zoneName, recordName, setIdentifier string) string {
	return acctest.ConfigCompose(testAccRecordConfig_baseCIDR(rName, locationName, zoneName), fmt.Sprintf(`
resource "aws_route53_record" "test" {
  zone_id        = aws_route53_zone.test.zone_id
  name           = %[1]q
  type           = "AAAA"
  ttl            = "60"
  set_identifier = %[2]q

  cidr_routing_policy {
    collection_id = aws_route53_cidr_collection.test.id
    location_name = "*"
  }

  records = ["2001:0db8::0123:4567:89ab:cdef"]
}
`, recordName, setIdentifier))
}

const testAccRecordConfig_failoverCNAME = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_health_check" "foo" {
  fqdn              = "dev.domain.test"
  port              = 80
  type              = "HTTP"
  resource_path     = "/"
  failure_threshold = "2"
  request_interval  = "30"

  tags = {
    Name = "tf-test-health-check"
  }
}

resource "aws_route53_record" "www-primary" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  failover_routing_policy {
    type = "PRIMARY"
  }

  health_check_id = aws_route53_health_check.foo.id
  set_identifier  = "www-primary"
  records         = ["primary.domain.test"]
}

resource "aws_route53_record" "www-secondary" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  failover_routing_policy {
    type = "SECONDARY"
  }

  set_identifier = "www-secondary"
  records        = ["secondary.domain.test"]
}
`

const testAccRecordConfig_weightedCNAME = `
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
`

const testAccRecordConfig_geolocationCNAME = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "default" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  geolocation_routing_policy {
    country = "*"
  }

  set_identifier = "Default"
  records        = ["dev.domain.test"]
}

resource "aws_route53_record" "california" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  geolocation_routing_policy {
    country     = "US"
    subdivision = "CA"
  }

  set_identifier = "California"
  records        = ["dev.domain.test"]
}

resource "aws_route53_record" "oceania" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  geolocation_routing_policy {
    continent = "OC"
  }

  set_identifier = "Oceania"
  records        = ["dev.domain.test"]
}

resource "aws_route53_record" "denmark" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  geolocation_routing_policy {
    country = "DK"
  }

  set_identifier = "Denmark"
  records        = ["dev.domain.test"]
}
`

func testAccRecordConfig_geoproximityCNAME(region string, localzonegroup string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "awsregion" {
  name    = "www"
  zone_id = aws_route53_zone.main.zone_id
  type    = "CNAME"
  ttl     = "5"

  geoproximity_routing_policy {
    aws_region = %[1]q
    bias       = 40
  }
  records        = ["dev.domain.test"]
  set_identifier = "awsregion"
}

resource "aws_route53_record" "localzonegroup" {
  name    = "www"
  zone_id = aws_route53_zone.main.zone_id
  type    = "CNAME"
  ttl     = "5"

  geoproximity_routing_policy {
    local_zone_group = %[2]q
  }
  records        = ["dev.domain.test"]
  set_identifier = "localzonegroup"
}

resource "aws_route53_record" "coordinates" {
  name    = "www"
  zone_id = aws_route53_zone.main.zone_id
  type    = "CNAME"
  ttl     = "5"

  geoproximity_routing_policy {
    coordinates {
      latitude  = "49.22"
      longitude = "-74.01"
    }
  }
  records        = ["dev.domain.test"]
  set_identifier = "coordinates"
}
`, region, localzonegroup)
}

func testAccRecordConfig_latencyCNAME(firstRegion, secondRegion, thirdRegion string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "first_region" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  latency_routing_policy {
    region = %[1]q
  }

  set_identifier = %[1]q
  records        = ["dev.domain.test"]
}

resource "aws_route53_record" "second_region" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  latency_routing_policy {
    region = %[2]q
  }

  set_identifier = %[2]q
  records        = ["dev.domain.test"]
}

resource "aws_route53_record" "third_region" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  latency_routing_policy {
    region = %[3]q
  }

  set_identifier = %[3]q
  records        = ["dev.domain.test"]
}
`, firstRegion, secondRegion, thirdRegion)
}

func testAccRecordConfig_aliasELB(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "alias" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "A"

  alias {
    zone_id                = aws_elb.main.zone_id
    name                   = aws_elb.main.dns_name
    evaluate_target_health = true
  }
}

resource "aws_elb" "main" {
  name = %[1]q

  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}
`, rName))
}

func testAccRecordConfig_aliasELBUppercase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "alias" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "A"

  alias {
    zone_id                = aws_elb.main.zone_id
    name                   = aws_elb.main.dns_name
    evaluate_target_health = true
  }
}

resource "aws_elb" "main" {
  name = %[1]q

  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}
`, rName))
}

func testAccRecordConfig_aliasS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_s3_bucket" "website" {
  bucket = %[1]q
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.website.id
  index_document {
    suffix = "index.html"
  }
}

resource "aws_route53_record" "alias" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "A"

  alias {
    zone_id                = aws_s3_bucket.website.hosted_zone_id
    name                   = aws_s3_bucket_website_configuration.test.website_domain
    evaluate_target_health = true
  }
}
`, rName)
}

func testAccRecordConfig_healthCheckIdSetIdentifier(setIdentifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  force_destroy = true
  name          = "domain.test"
}

resource "aws_route53_health_check" "test" {
  failure_threshold = "2"
  fqdn              = "test.domain.test"
  port              = 80
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTP"
}

resource "aws_route53_record" "test" {
  zone_id         = aws_route53_zone.test.zone_id
  health_check_id = aws_route53_health_check.test.id
  name            = "test"
  records         = ["127.0.0.1"]
  set_identifier  = %[1]q
  ttl             = "5"
  type            = "A"

  weighted_routing_policy {
    weight = 1
  }
}
`, setIdentifier)
}

func testAccRecordConfig_healthCheckIdTypeA() string {
	return `
resource "aws_route53_zone" "test" {
  force_destroy = true
  name          = "domain.test"
}

resource "aws_route53_health_check" "test" {
  failure_threshold = "2"
  fqdn              = "test.domain.test"
  port              = 80
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTP"
}

resource "aws_route53_record" "test" {
  zone_id         = aws_route53_zone.test.zone_id
  health_check_id = aws_route53_health_check.test.id
  name            = "test"
  records         = ["127.0.0.1"]
  set_identifier  = "set-id"
  ttl             = "5"
  type            = "A"

  weighted_routing_policy {
    weight = 1
  }
}
`
}

func testAccRecordConfig_healthCheckIdTypeCNAME() string {
	return `
resource "aws_route53_zone" "test" {
  force_destroy = true
  name          = "domain.test"
}

resource "aws_route53_health_check" "test" {
  failure_threshold = "2"
  fqdn              = "test.domain.test"
  port              = 80
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTP"
}

resource "aws_route53_record" "test" {
  zone_id         = aws_route53_zone.test.zone_id
  health_check_id = aws_route53_health_check.test.id
  name            = "test"
  records         = ["test1.domain.test"]
  set_identifier  = "set-id"
  ttl             = "5"
  type            = "CNAME"

  weighted_routing_policy {
    weight = 1
  }
}
`
}

func testAccRecordConfig_baseVPCEndpoint(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = [aws_lb.test.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = [aws_security_group.test.id]
  service_name        = aws_vpc_endpoint_service.test.service_name
  subnet_ids          = aws_subnet.test[*].id
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_zone" "test" {
  name = "domain.test"

  vpc {
    vpc_id = aws_vpc.test.id
  }
}
`, rName))
}

func testAccRecordConfig_aliasCustomVPCEndpointSwappedAliasAttributes(rName string) string {
	return acctest.ConfigCompose(testAccRecordConfig_baseVPCEndpoint(rName), `
resource "aws_route53_record" "test" {
  name    = "test"
  type    = "A"
  zone_id = aws_route53_zone.test.zone_id

  alias {
    evaluate_target_health = false
    name                   = lookup(aws_vpc_endpoint.test.dns_entry[0], "hosted_zone_id", "")
    zone_id                = lookup(aws_vpc_endpoint.test.dns_entry[0], "dns_name", "")
  }
}
`)
}

func testAccRecordConfig_customVPCEndpoint(rName string) string {
	return acctest.ConfigCompose(testAccRecordConfig_baseVPCEndpoint(rName), `
resource "aws_route53_record" "test" {
  name    = "test"
  type    = "A"
  zone_id = aws_route53_zone.test.zone_id

  alias {
    evaluate_target_health = false
    name                   = lookup(aws_vpc_endpoint.test.dns_entry[0], "dns_name", "")
    zone_id                = lookup(aws_vpc_endpoint.test.dns_entry[0], "hosted_zone_id", "")
  }
}
`)
}

func testAccRecordConfig_weightedELBAlias(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2), `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_elb" "live" {
  name = "foobar-terraform-elb-live"

  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_route53_record" "elb_weighted_alias_live" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "A"

  weighted_routing_policy {
    weight = 90
  }

  set_identifier = "live"

  alias {
    zone_id                = aws_elb.live.zone_id
    name                   = aws_elb.live.dns_name
    evaluate_target_health = true
  }
}

resource "aws_elb" "dev" {
  name = "foobar-terraform-elb-dev"

  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_route53_record" "elb_weighted_alias_dev" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "A"

  weighted_routing_policy {
    weight = 10
  }

  set_identifier = "dev"

  alias {
    zone_id                = aws_elb.dev.zone_id
    name                   = aws_elb.dev.dns_name
    evaluate_target_health = true
  }
}
`)
}

const testAccRecordConfig_weightedAlias = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "blue_origin" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "blue-origin"
  type    = "CNAME"
  ttl     = 5
  records = ["v1.terraform.io"]
}

resource "aws_route53_record" "r53_weighted_alias_live" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"

  weighted_routing_policy {
    weight = 90
  }

  set_identifier = "blue"

  alias {
    zone_id                = aws_route53_zone.main.zone_id
    name                   = "${aws_route53_record.blue_origin.name}.${aws_route53_zone.main.name}"
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "green_origin" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "green-origin"
  type    = "CNAME"
  ttl     = 5
  records = ["v2.terraform.io"]
}

resource "aws_route53_record" "r53_weighted_alias_dev" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"

  weighted_routing_policy {
    weight = 10
  }

  set_identifier = "green"

  alias {
    zone_id                = aws_route53_zone.main.zone_id
    name                   = "${aws_route53_record.green_origin.name}.${aws_route53_zone.main.name}"
    evaluate_target_health = false
  }
}
`

const testAccRecordConfig_typeChangePre = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "sample" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "sample"
  type    = "CNAME"
  ttl     = "30"
  records = ["www.terraform.io"]
}
`

const testAccRecordConfig_typeChangePost = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "sample" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "sample"
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "8.8.8.8"]
}
`

const testAccRecordConfig_nameChangePre = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "sample" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "sample"
  type    = "CNAME"
  ttl     = "30"
  records = ["www.terraform.io"]
}
`

const testAccRecordConfig_nameChangePost = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "sample" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "sample-new"
  type    = "CNAME"
  ttl     = "30"
  records = ["www.terraform.io"]
}
`

const testAccRecordConfig_setIdentifierChangeBasicToWeightedPre = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "basic_to_weighted" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "sample"
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "8.8.8.8"]
}
`

const testAccRecordConfig_setIdentifierChangeBasicToWeightedPost = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "basic_to_weighted" {
  zone_id        = aws_route53_zone.main.zone_id
  name           = "sample"
  type           = "A"
  ttl            = "30"
  records        = ["127.0.0.1", "8.8.8.8"]
  set_identifier = "cluster-a"

  weighted_routing_policy {
    weight = 100
  }
}
`

func testAccRecordConfig_setIdentifierRenameFailover(set_identifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_health_check" "foo" {
  fqdn              = "dev.domain.test"
  port              = 80
  type              = "HTTP"
  resource_path     = "/"
  failure_threshold = "2"
  request_interval  = "30"

  tags = {
    Name = "tf-test-health-check"
  }
}

resource "aws_route53_record" "set_identifier_rename_failover" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  failover_routing_policy {
    type = "PRIMARY"
  }

  health_check_id = aws_route53_health_check.foo.id
  set_identifier  = %[1]q
  records         = ["primary.domain.test"]
}
`, set_identifier)
}

func testAccRecordConfig_setIdentifierRenameGeolocationContinent(continent, set_identifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "set_identifier_rename_geolocation" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  geolocation_routing_policy {
    continent = %[1]q
  }

  set_identifier = %[2]q
  records        = ["primary.domain.test"]
}
`, continent, set_identifier)
}

func testAccRecordConfig_setIdentifierRenameGeolocationCountry(country, set_identifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "set_identifier_rename_geolocation" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  geolocation_routing_policy {
    country = %[1]q
  }

  set_identifier = %[2]q
  records        = ["primary.domain.test"]
}
`, country, set_identifier)
}

func testAccRecordConfig_setIdentifierRenameGeolocationCountrySubdivision(country, subdivision, set_identifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "set_identifier_rename_geolocation" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  geolocation_routing_policy {
    country     = %[1]q
    subdivision = %[2]q
  }

  set_identifier = %[3]q
  records        = ["primary.domain.test"]
}
`, country, subdivision, set_identifier)
}

func testAccRecordConfig_setIdentifierRenameGeoproximityRegion(region, set_identifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "set_identifier_rename_geoproximity" {
  name    = "www"
  zone_id = aws_route53_zone.main.zone_id
  type    = "CNAME"
  ttl     = "5"

  geoproximity_routing_policy {
    aws_region = %[1]q
  }

  records        = ["dev.example.com"]
  set_identifier = %[2]q
}
`, region, set_identifier)
}

func testAccRecordConfig_setIdentifierRenameGeoproximityLocalZoneGroup(zonegroup, set_identifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "set_identifier_rename_geoproximity" {
  name    = "www"
  zone_id = aws_route53_zone.main.zone_id
  type    = "CNAME"
  ttl     = "5"

  geoproximity_routing_policy {
    local_zone_group = %[1]q
  }

  records        = ["dev.example.com"]
  set_identifier = %[2]q
}
`, zonegroup, set_identifier)
}

func testAccRecordConfig_setIdentifierRenameGeoproximityCoordinates(latitude, longitude, set_identifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "set_identifier_rename_geoproximity" {
  name    = "www"
  zone_id = aws_route53_zone.main.zone_id
  type    = "CNAME"
  ttl     = "5"

  geoproximity_routing_policy {
    coordinates {
      latitude  = %[1]q
      longitude = %[2]q
    }
  }

  records        = ["dev.example.com"]
  set_identifier = %[3]q
}
`, latitude, longitude, set_identifier)
}

func testAccRecordConfig_setIdentifierRenameLatency(region, set_identifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "set_identifier_rename_latency" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = "5"

  latency_routing_policy {
    region = %[1]q
  }

  set_identifier = %[2]q
  records        = ["dev.domain.test"]
}

`, region, set_identifier)
}

func testAccRecordConfig_setIdentifierRenameMultiValueAnswer(set_identifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "set_identifier_rename_multivalue_answer" {
  zone_id                          = aws_route53_zone.main.zone_id
  name                             = "www"
  type                             = "A"
  ttl                              = "5"
  multivalue_answer_routing_policy = true
  set_identifier                   = %[1]q
  records                          = ["127.0.0.1"]
}
`, set_identifier)
}

func testAccRecordConfig_setIdentifierRenameWeighted(set_identifier string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "test" {
  zone_id        = aws_route53_zone.main.zone_id
  name           = "sample"
  type           = "A"
  ttl            = "30"
  records        = ["127.0.0.1", "8.8.8.8"]
  set_identifier = %[1]q

  weighted_routing_policy {
    weight = 100
  }
}
`, set_identifier)
}

func testAccRecordConfig_aliasChangePre(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_elb" "test" {
  name = %[1]q

  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "alias-change"
  type    = "A"

  alias {
    zone_id                = aws_elb.test.zone_id
    name                   = aws_elb.test.dns_name
    evaluate_target_health = true
  }
}
`, rName))
}

func testAccRecordConfig_aliasChangePost() string {
	return `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "alias-change"
  type    = "CNAME"
  ttl     = "30"
  records = ["www.terraform.io"]
}
`
}

const testAccRecordConfig_emptyName = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "empty" {
  zone_id = aws_route53_zone.main.zone_id
  name    = ""
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1"]
}
`

func testAccRecordConfig_aliasChangeDualstackPre(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "domain.test"
}

resource "aws_elb" "test" {
  name = %[1]q

  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = "alias-change-ds"
  type    = "A"

  alias {
    zone_id                = aws_elb.test.zone_id
    name                   = "dualstack.${aws_elb.test.dns_name}"
    evaluate_target_health = true
  }
}
 `, rName))
}

func testAccRecordConfig_aliasChangeDualstackPost(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "domain.test"
}

resource "aws_elb" "test" {
  name = %[1]q

  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = "alias-change-ds"
  type    = "A"

  alias {
    zone_id                = aws_elb.test.zone_id
    name                   = aws_elb.test.dns_name
    evaluate_target_health = true
  }
}
 `, rName))
}

const testAccRecordConfig_longTxt = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "long_txt" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "google.domain.test"
  type    = "TXT"
  ttl     = "30"
  records = [
    "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAiajKNMp\" \"/A12roF4p3MBm9QxQu6GDsBlWUWFx8EaS8TCo3Qe8Cj0kTag1JMjzCC1s6oM0a43JhO6mp6z/"
  ]
}
`

const testAccRecordConfig_underscoreInName = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "underscore" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_underscore.domain.test"
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1"]
}
`

const testAccRecordConfig_multiValueAnswerA = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "www-server1" {
  zone_id                          = aws_route53_zone.main.zone_id
  name                             = "www"
  type                             = "A"
  ttl                              = "5"
  multivalue_answer_routing_policy = true
  set_identifier                   = "server1"
  records                          = ["127.0.0.1"]
}

resource "aws_route53_record" "www-server2" {
  zone_id                          = aws_route53_zone.main.zone_id
  name                             = "www"
  type                             = "A"
  ttl                              = "5"
  multivalue_answer_routing_policy = true
  set_identifier                   = "server2"
  records                          = ["127.0.0.2"]
}
`

const testAccRecordConfig_weightedRoutingPolicy = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "www-server1" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "A"

  weighted_routing_policy {
    weight = 5
  }

  ttl            = "300"
  set_identifier = "server1"
  records        = ["127.0.0.1"]
}
`

const testAccRecordConfig_simpleRoutingPolicy = `
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_route53_record" "www-server1" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "A"
  ttl     = "300"
  records = ["127.0.0.1"]
}
`

func testAccRecordConfig_ttl(zoneName, recordName string, ttl int) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[1]q
}

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[2]q
  type    = "A"
  ttl     = %[3]d
  records = ["127.0.0.1", "127.0.0.27"]
}
`, zoneName, recordName, ttl)
}

func testAccRecordConfig_aliasWildcardName(rName, zoneName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 3

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.region}.s3"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false
  subnet_ids          = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_zone" "test1" {
  name = %[2]q
}

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test1.zone_id
  name    = "*.%[2]s"
  type    = "A"

  alias {
    name                   = aws_vpc_endpoint.test.dns_entry[0].dns_name
    zone_id                = aws_vpc_endpoint.test.dns_entry[0].hosted_zone_id
    evaluate_target_health = false
  }
}
`, rName, zoneName))
}
