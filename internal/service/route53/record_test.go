// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_basic(zoneName.String(), strings.ToUpper(recordName.String())),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "alias.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "allow_overwrite"),
					resource.TestCheckResourceAttr(resourceName, "cidr_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "failover_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fqdn", recordName.String()),
					resource.TestCheckResourceAttr(resourceName, "geolocation_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "geoproximity_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "health_check_id", ""),
					resource.TestCheckResourceAttr(resourceName, "latency_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "multivalue_answer_routing_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, recordName.String()),
					resource.TestCheckResourceAttr(resourceName, "records.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "records.*", "127.0.0.1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "records.*", "127.0.0.27"),
					resource.TestCheckResourceAttr(resourceName, "set_identifier", ""),
					resource.TestCheckResourceAttr(resourceName, "ttl", "30"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "A"),
					resource.TestCheckResourceAttr(resourceName, "weighted_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceRecord(), resourceName),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_multiple(zoneName.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, "aws_route53_record.test.0", &v1),
					testAccCheckRecordExists(ctx, "aws_route53_record.test.1", &v2),
					testAccCheckRecordExists(ctx, "aws_route53_record.test.2", &v3),
					testAccCheckRecordExists(ctx, "aws_route53_record.test.3", &v4),
					testAccCheckRecordExists(ctx, "aws_route53_record.test.4", &v5),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceRecord(), "aws_route53_record.test.0"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_underscoreInName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_fqdn(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_fqdn,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
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
					testAccCheckRecordExists(ctx, resourceName, &record2),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_nameTrailingPeriod,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_Support_txt(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_txt,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight, "zone_id"},
			},
		},
	})
}

func TestAccRoute53Record_Support_spf(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_spf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
					resource.TestCheckTypeSetElemAttr(resourceName, "records.*", "include:domain.test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_Support_caa(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_caa,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
					resource.TestCheckTypeSetElemAttr(resourceName, "records.*", "0 issue \"domainca.test;\""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_Support_ds(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_ds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_generatesSuffix(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_suffix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_wildcard(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.wildcard"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_wildcard,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRecordConfig_wildcardUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_failover(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.www-primary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_failoverCNAME,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
					testAccCheckRecordExists(ctx, "aws_route53_record.www-secondary", &record2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_Weighted_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2, record3 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.www-live"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_weightedCNAME,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, "aws_route53_record.www-dev", &record1),
					testAccCheckRecordExists(ctx, resourceName, &record2),
					testAccCheckRecordExists(ctx, "aws_route53_record.www-off", &record3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_WeightedToSimple_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.www-server1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_weightedRoutingPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
			{
				Config: testAccRecordConfig_simpleRoutingPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
		},
	})
}

func TestAccRoute53Record_Alias_elb(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.alias"

	rs := sdkacctest.RandString(10)
	testAccRecordConfig_config := fmt.Sprintf(testAccRecordConfig_aliasELB, rs)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_Alias_s3(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_record.alias"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_aliasS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_Alias_vpcEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_record.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccRecordConfig_aliasCustomVPCEndpointSwappedAliasAttributes(rName),
				ExpectError: regexache.MustCompile(`expected length of`),
			},
			{
				Config: testAccRecordConfig_customVPCEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_Alias_uppercase(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.alias"

	rs := sdkacctest.RandString(10)
	testAccRecordConfig_config := fmt.Sprintf(testAccRecordConfig_aliasELBUppercase, rs)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_Weighted_alias(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2, record3, record4, record5, record6 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.elb_weighted_alias_live"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_weightedELBAlias,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
					testAccCheckRecordExists(ctx, "aws_route53_record.elb_weighted_alias_dev", &record2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},

			{
				Config: testAccRecordConfig_weightedAlias,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, "aws_route53_record.green_origin", &record3),
					testAccCheckRecordExists(ctx, "aws_route53_record.r53_weighted_alias_live", &record4),
					testAccCheckRecordExists(ctx, "aws_route53_record.blue_origin", &record5),
					testAccCheckRecordExists(ctx, "aws_route53_record.r53_weighted_alias_dev", &record6),
				),
			},
		},
	})
}

func TestAccRoute53Record_cidr(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	locationName := sdkacctest.RandString(16)
	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_cidr(rName, locationName, zoneName.String(), recordName.String(), "cidr-location-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "alias.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "allow_overwrite"),
					resource.TestCheckResourceAttr(resourceName, "cidr_routing_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_routing_policy.0.collection_id", "aws_route53_cidr_collection.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_routing_policy.0.location_name", "aws_route53_cidr_location.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "failover_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fqdn", recordName.String()),
					resource.TestCheckResourceAttr(resourceName, "geolocation_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "geoproximity_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "health_check_id", ""),
					resource.TestCheckResourceAttr(resourceName, "latency_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "multivalue_answer_routing_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, recordName.String()),
					resource.TestCheckResourceAttr(resourceName, "records.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "records.*", "2001:0db8::0123:4567:89ab:cdef"),
					resource.TestCheckResourceAttr(resourceName, "set_identifier", "cidr-location-1"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "60"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AAAA"),
					resource.TestCheckResourceAttr(resourceName, "weighted_routing_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
			{
				Config: testAccRecordConfig_cidr(rName, locationName, zoneName.String(), recordName.String(), "cidr-location-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cidr_routing_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_routing_policy.0.collection_id", "aws_route53_cidr_collection.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_routing_policy.0.location_name", "aws_route53_cidr_location.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "set_identifier", "cidr-location-2"),
				),
			},
			{
				Config: testAccRecordConfig_cidrDefaultLocation(rName, locationName, zoneName.String(), recordName.String(), "cidr-location-3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cidr_routing_policy.#", acctest.Ct1),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_geolocationCNAME,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, "aws_route53_record.default", &record1),
					testAccCheckRecordExists(ctx, "aws_route53_record.california", &record2),
					testAccCheckRecordExists(ctx, "aws_route53_record.oceania", &record3),
					testAccCheckRecordExists(ctx, "aws_route53_record.denmark", &record4),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_Geoproximity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2, record3 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.awsregion"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_geoproximityCNAME(names.USEast1RegionID, fmt.Sprintf("%s-atl-1", names.USEast1RegionID)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, "aws_route53_record.awsregion", &record1),
					testAccCheckRecordExists(ctx, "aws_route53_record.localzonegroup", &record2),
					testAccCheckRecordExists(ctx, "aws_route53_record.coordinates", &record3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_HealthCheckID_setIdentifierChange(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_healthCheckIdSetIdentifier("test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
			{
				Config: testAccRecordConfig_healthCheckIdSetIdentifier("test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_HealthCheckID_typeChange(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_healthCheckIdTypeCNAME(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
			{
				Config: testAccRecordConfig_healthCheckIdTypeA(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_Latency_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2, record3 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.first_region"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_latencyCNAME(names.USEast1RegionID, names.EUWest1RegionID, names.APNortheast1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
					testAccCheckRecordExists(ctx, "aws_route53_record.second_region", &record2),
					testAccCheckRecordExists(ctx, "aws_route53_record.third_region", &record3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_typeChange(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.sample"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_typeChangePre,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRecordConfig_typeChangePost,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_nameChange(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.sample"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_nameChangePre,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
			// Cause a change, which will trigger a refresh
			{
				Config: testAccRecordConfig_nameChangePost,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
					testAccCheckRecordDoesNotExist(ctx, "aws_route53_zone.main", "sample", "CNAME"),
				),
			},
		},
	})
}

func TestAccRoute53Record_setIdentifierChangeBasicToWeighted(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.basic_to_weighted"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierChangeBasicToWeightedPre,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRecordConfig_setIdentifierChangeBasicToWeightedPost,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geolocationContinent(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geolocation"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationContinent("AN", "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
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
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geolocationCountryDefault(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geolocation"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationCountry("*", "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
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
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geolocationCountrySpecified(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geolocation"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationCountry("US", "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
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
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geolocationCountrySubdivision(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geolocation"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeolocationCountrySubdivision("US", "CA", "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
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
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geoproximityRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geoproximity"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityRegion(names.USEast1RegionID, "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityRegion(names.USEast1RegionID, "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geoproximityLocalZoneGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geoproximity"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityLocalZoneGroup(fmt.Sprintf("%s-atl-1", names.USEast1RegionID), "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityLocalZoneGroup(fmt.Sprintf("%s-atl-1", names.USEast1RegionID), "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_geoproximityCoordinates(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_geoproximity"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameGeoproximityCoordinates("49.22", "-74.01", "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
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
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_failover(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_failover"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameFailover("before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
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
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_latency(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_latency"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameLatency(names.USEast1RegionID, "before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite"},
			},
			{
				Config: testAccRecordConfig_setIdentifierRenameLatency(names.USEast1RegionID, "after"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_multiValueAnswer(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_multivalue_answer"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameMultiValueAnswer("before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
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
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_SetIdentifierRename_weighted(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.set_identifier_rename_weighted"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_setIdentifierRenameWeighted("before"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
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
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_Alias_change(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_aliasChangePre(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRecordConfig_aliasChangePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
				),
			},
		},
	})
}

func TestAccRoute53Record_Alias_changeDualstack(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_aliasChangeDualstackPre(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
					testAccCheckRecordAliasNameDualstack(&record1, true),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
			// Cause a change, which will trigger a refresh
			{
				Config: testAccRecordConfig_aliasChangeDualstackPost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record2),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_emptyName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/8423
func TestAccRoute53Record_longTXTrecord(t *testing.T) {
	ctx := acctest.Context(t)
	var record1 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.long_txt"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_longTxt,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_MultiValueAnswer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var record1, record2 awstypes.ResourceRecordSet
	resourceName := "aws_route53_record.www-server1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_multiValueAnswerA,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordExists(ctx, "aws_route53_record.www-server1", &record1),
					testAccCheckRecordExists(ctx, "aws_route53_record.www-server2", &record2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
			},
		},
	})
}

func TestAccRoute53Record_Allow_doNotOverwrite(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               testAccRecordOverwriteExpectErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordConfig_allowOverwrite(true),
				Check:  resource.ComposeTestCheckFunc(testAccCheckRecordExists(ctx, "aws_route53_record.overwriting", &record1)),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", names.AttrWeight},
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

func testAccCheckRecordDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_record" {
				continue
			}

			parts := tfroute53.RecordParseResourceID(rs.Primary.ID)
			zone := parts[0]
			recordName := parts[1]
			recordType := parts[2]

			_, _, err := tfroute53.FindResourceRecordSetByFourPartKey(ctx, conn, tfroute53.CleanZoneID(zone), recordName, recordType, rs.Primary.Attributes["set_identifier"])

			if tfresource.NotFound(err) {
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

func testAccCheckRecordExists(ctx context.Context, n string, v *awstypes.ResourceRecordSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		parts := tfroute53.RecordParseResourceID(rs.Primary.ID)
		zone := parts[0]
		recordName := parts[1]
		recordType := parts[2]

		output, _, err := tfroute53.FindResourceRecordSetByFourPartKey(ctx, conn, tfroute53.CleanZoneID(zone), recordName, recordType, rs.Primary.Attributes["set_identifier"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRecordDoesNotExist(ctx context.Context, zoneResourceName string, recordName string, recordType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		zoneResource, ok := s.RootModule().Resources[zoneResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", zoneResourceName)
		}

		zoneId := zoneResource.Primary.ID
		en := tfroute53.ExpandRecordName(recordName, zoneResource.Primary.Attributes["zone_id"])

		lopts := &route53.ListResourceRecordSetsInput{
			HostedZoneId: aws.String(tfroute53.CleanZoneID(zoneId)),
		}

		resp, err := conn.ListResourceRecordSets(ctx, lopts)
		if err != nil {
			return err
		}

		found := false
		for _, rec := range resp.ResourceRecordSets {
			recName := tfroute53.CleanRecordName(*rec.Name)
			if tfroute53.FQDN(strings.ToLower(recName)) == tfroute53.FQDN(strings.ToLower(en)) && rec.Type == awstypes.RRType(recordType) {
				found = true
				break
			}
		}

		if found {
			return fmt.Errorf("Record exists but should not: %s", en)
		}

		return nil
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

const testAccRecordConfig_aliasELB = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

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
  name               = "foobar-terraform-elb-%s"
  availability_zones = slice(data.aws_availability_zones.available.names, 0, 1)

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}
`

const testAccRecordConfig_aliasELBUppercase = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

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
  name               = "FOOBAR-TERRAFORM-ELB-%s"
  availability_zones = slice(data.aws_availability_zones.available.names, 0, 1)

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}
`

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
  set_identifier  = "test"
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
  set_identifier  = "test"
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
    name                   = lookup(aws_vpc_endpoint.test.dns_entry[0], "hosted_zone_id")
    zone_id                = lookup(aws_vpc_endpoint.test.dns_entry[0], "dns_name")
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
    name                   = lookup(aws_vpc_endpoint.test.dns_entry[0], "dns_name")
    zone_id                = lookup(aws_vpc_endpoint.test.dns_entry[0], "hosted_zone_id")
  }
}
`)
}

const testAccRecordConfig_weightedELBAlias = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_elb" "live" {
  name               = "foobar-terraform-elb-live"
  availability_zones = slice(data.aws_availability_zones.available.names, 0, 1)

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
  name               = "foobar-terraform-elb-dev"
  availability_zones = slice(data.aws_availability_zones.available.names, 0, 1)

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
`

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

resource "aws_route53_record" "set_identifier_rename_weighted" {
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
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = slice(data.aws_availability_zones.available.names, 0, 1)

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
`, rName)
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
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_route53_zone" "test" {
  name = "domain.test"
}

resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = slice(data.aws_availability_zones.available.names, 0, 1)

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
 `, rName)
}

func testAccRecordConfig_aliasChangeDualstackPost(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_route53_zone" "test" {
  name = "domain.test"
}

resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = slice(data.aws_availability_zones.available.names, 0, 1)

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
 `, rName)
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
