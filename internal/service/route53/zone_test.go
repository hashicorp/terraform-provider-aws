// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53Zone_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_basic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					acctest.MatchResourceAttrGlobalARNNoAccount(resourceName, names.AttrARN, "route53", regexache.MustCompile("hostedzone/.+")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, zoneName),
					resource.TestCheckResourceAttr(resourceName, "name_servers.#", "4"),
					resource.TestCheckResourceAttrSet(resourceName, "primary_name_server"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "enable_accelerated_recovery", acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			// Test import using an ID with "/hosrtezone/" prefix.
			// https://github.com/hashicorp/terraform-provider-aws/issues/37817.
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(v *route53.GetHostedZoneOutput) resource.ImportStateIdFunc {
					return func(s *terraform.State) (string, error) {
						return aws.ToString(v.HostedZone.Id), nil
					}
				}(&zone),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "zone_id",
				ImportStateVerifyIgnore:              []string{names.AttrForceDestroy, names.AttrID},
			},
		},
	})
}

func TestAccRoute53Zone_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_basic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					acctest.CheckSDKResourceDisappears(ctx, t, tfroute53.ResourceZone(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53Zone_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var zone0, zone1, zone2, zone3, zone4 route53.GetHostedZoneOutput
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_multiple(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, "aws_route53_zone.test.0", &zone0),
					testAccCheckDomainName(&zone0, fmt.Sprintf("subdomain0.%s.", domainName)),
					testAccCheckZoneExists(ctx, t, "aws_route53_zone.test.1", &zone1),
					testAccCheckDomainName(&zone1, fmt.Sprintf("subdomain1.%s.", domainName)),
					testAccCheckZoneExists(ctx, t, "aws_route53_zone.test.2", &zone2),
					testAccCheckDomainName(&zone2, fmt.Sprintf("subdomain2.%s.", domainName)),
					testAccCheckZoneExists(ctx, t, "aws_route53_zone.test.3", &zone3),
					testAccCheckDomainName(&zone3, fmt.Sprintf("subdomain3.%s.", domainName)),
					testAccCheckZoneExists(ctx, t, "aws_route53_zone.test.4", &zone4),
					testAccCheckDomainName(&zone4, fmt.Sprintf("subdomain4.%s.", domainName)),
				),
			},
		},
	})
}

func TestAccRoute53Zone_comment(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_comment(zoneName, "comment1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "comment1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccZoneConfig_comment(zoneName, "comment2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "comment2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccRoute53Zone_delegationSetID(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	delegationSetResourceName := "aws_route53_delegation_set.test"
	resourceName := "aws_route53_zone.test"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_delegationSetID(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					resource.TestCheckResourceAttrPair(resourceName, "delegation_set_id", delegationSetResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccRoute53Zone_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_forceDestroy(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					// Add >100 records to verify pagination works ok
					testAccCreateRandomRecordsInZoneID(ctx, t, &zone, 100),
					testAccCreateRandomRecordsInZoneID(ctx, t, &zone, 5),
				),
			},
		},
	})
}

func TestAccRoute53Zone_ForceDestroy_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_forceDestroyTrailingPeriod(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					// Add >100 records to verify pagination works ok
					testAccCreateRandomRecordsInZoneID(ctx, t, &zone, 100),
					testAccCreateRandomRecordsInZoneID(ctx, t, &zone, 5),
				),
			},
		},
	})
}

func TestAccRoute53Zone_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_tags1(zoneName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccZoneConfig_tags2(zoneName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccZoneConfig_tags1(zoneName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccRoute53Zone_VPC_single(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_zone.test"
	vpcResourceName := "aws_vpc.test1"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_vpcSingle(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "1"),
					testAccCheckZoneAssociatesVPC(vpcResourceName, &zone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccRoute53Zone_VPC_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_zone.test"
	vpcResourceName1 := "aws_vpc.test1"
	vpcResourceName2 := "aws_vpc.test2"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_vpcMultiple(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "2"),
					testAccCheckZoneAssociatesVPC(vpcResourceName1, &zone),
					testAccCheckZoneAssociatesVPC(vpcResourceName2, &zone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccRoute53Zone_VPC_updates(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_zone.test"
	vpcResourceName1 := "aws_vpc.test1"
	vpcResourceName2 := "aws_vpc.test2"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_vpcSingle(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "1"),
					testAccCheckZoneAssociatesVPC(vpcResourceName1, &zone),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccZoneConfig_vpcMultiple(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "2"),
					testAccCheckZoneAssociatesVPC(vpcResourceName1, &zone),
					testAccCheckZoneAssociatesVPC(vpcResourceName2, &zone),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccZoneConfig_vpcSingle(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "1"),
					testAccCheckZoneAssociatesVPC(vpcResourceName1, &zone),
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

// Excercises exception handling during forced destruction in partitions
// which do no support DNSSEC (e.g. GovCloud).
//
// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/22334
func TestAccRoute53Zone_VPC_single_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_zone.test"
	vpcResourceName := "aws_vpc.test1"
	zoneName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_vpcSingle_forceDestroy(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "1"),
					testAccCheckZoneAssociatesVPC(vpcResourceName, &zone),
					// Add >100 records to verify pagination works ok
					testAccCreateRandomRecordsInZoneID(ctx, t, &zone, 100),
					testAccCreateRandomRecordsInZoneID(ctx, t, &zone, 5),
				),
			},
		},
	})
}

func TestAccRoute53Zone_escapedCharacter(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// "a%b.<RandomDomainName>".
				// Double escaped due to use of 'fmt.Sprintf'.
				Config: testAccZoneConfig_basic("a\\\\044b." + domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "a\\044b."+domainName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccRoute53Zone_classlessDelegation(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	// https://datatracker.ietf.org/doc/html/rfc2317.
	zoneName := "1.0/25.2.0.192.in-addr.arpa"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_basic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
				),
			},
		},
	})
}

func TestAccRoute53Zone_escapedSlash(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	zoneName := "0/24." + acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_basic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccRoute53Zone_escapedSpace(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	// double-escape is required for templating the tf resource
	zoneName := "a\\\\040b." + acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_basic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccRoute53Zone_enableAcceleratedRecovery(t *testing.T) {
	ctx := acctest.Context(t)
	var zone1, zone2 route53.GetHostedZoneOutput
	resourceName1 := "aws_route53_zone.test1"
	resourceName2 := "aws_route53_zone.test2"
	zoneName1 := acctest.RandomDomainName()
	zoneName2 := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_enableAcceleratedRecovery(zoneName1, zoneName2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName1, &zone1),
					testAccCheckZoneExists(ctx, t, resourceName2, &zone2),
					resource.TestCheckResourceAttr(resourceName1, "enable_accelerated_recovery", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName2, "enable_accelerated_recovery", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				ResourceName:            resourceName2,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				// Disable accelerated recovery
				Config: testAccZoneConfig_enableAcceleratedRecovery(zoneName1, zoneName2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName1, &zone1),
					testAccCheckZoneExists(ctx, t, resourceName2, &zone2),
					resource.TestCheckResourceAttr(resourceName1, "enable_accelerated_recovery", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName2, "enable_accelerated_recovery", acctest.CtFalse),
				),
			},
			{
				// Re-enable accelerated recovery
				// Check a resource can be destroyed with it enabled
				Config: testAccZoneConfig_enableAcceleratedRecovery(zoneName1, zoneName2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName1, &zone1),
					testAccCheckZoneExists(ctx, t, resourceName2, &zone2),
					resource.TestCheckResourceAttr(resourceName1, "enable_accelerated_recovery", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName2, "enable_accelerated_recovery", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRoute53Zone_nameUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var zone1, zone2 route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone.test"
	zoneName1 := acctest.RandomDomainName()
	zoneName2 := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneConfig_basic(zoneName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, zoneName1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccZoneConfig_basic(zoneName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneExists(ctx, t, resourceName, &zone2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, zoneName2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccCheckZoneDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_zone" {
				continue
			}

			_, err := tfroute53.FindHostedZoneByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Hosted Zone %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCreateRandomRecordsInZoneID(ctx context.Context, t *testing.T, zone *route53.GetHostedZoneOutput, recordsCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		var changes []awstypes.Change
		if recordsCount > 100 {
			return fmt.Errorf("Route53 API only allows 100 record sets in a single batch")
		}
		for range recordsCount {
			changes = append(changes, awstypes.Change{
				Action: awstypes.ChangeActionUpsert,
				ResourceRecordSet: &awstypes.ResourceRecordSet{
					Name: aws.String(fmt.Sprintf("%d-tf-acc-random.%s", acctest.RandInt(t), *zone.HostedZone.Name)),
					Type: awstypes.RRTypeCname,
					ResourceRecords: []awstypes.ResourceRecord{
						{Value: aws.String(fmt.Sprintf("random.%s", *zone.HostedZone.Name))},
					},
					TTL: aws.Int64(30),
				},
			})
		}

		input := &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: zone.HostedZone.Id,
			ChangeBatch: &awstypes.ChangeBatch{
				Comment: aws.String("Generated by Terraform"),
				Changes: changes,
			},
		}
		output, err := conn.ChangeResourceRecordSets(ctx, input)

		if err != nil {
			return err
		}

		timeout := 30 * time.Minute
		if output.ChangeInfo != nil {
			if _, err := tfroute53.WaitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id), timeout); err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckZoneExists(ctx context.Context, t *testing.T, n string, v *route53.GetHostedZoneOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		output, err := tfroute53.FindHostedZoneByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckZoneAssociatesVPC(n string, zone *route53.GetHostedZoneOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		for _, vpc := range zone.VPCs {
			if aws.ToString(vpc.VPCId) == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("VPC: %s is not associated to Zone: %v", n, tfroute53.CleanZoneID(aws.ToString(zone.HostedZone.Id)))
	}
}

func testAccCheckDomainName(zone *route53.GetHostedZoneOutput, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if zone.HostedZone.Name == nil {
			return fmt.Errorf("Empty name in HostedZone for domain %s", domain)
		}

		if aws.ToString(zone.HostedZone.Name) == domain {
			return nil
		}

		return fmt.Errorf("Invalid domain name. Expected %s is %s", domain, *zone.HostedZone.Name)
	}
}

func testAccZoneConfig_basic(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "%[1]s."
}
`, zoneName)
}

func testAccZoneConfig_multiple(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  count = 5

  name = "subdomain${count.index}.%[1]s"
}
`, domainName)
}

func testAccZoneConfig_comment(zoneName, comment string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  comment = %[1]q
  name    = "%[2]s."
}
`, comment, zoneName)
}

func testAccZoneConfig_delegationSetID(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_delegation_set" "test" {}

resource "aws_route53_zone" "test" {
  delegation_set_id = aws_route53_delegation_set.test.id
  name              = "%[1]s."
}
`, zoneName)
}

func testAccZoneConfig_forceDestroy(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  force_destroy = true
  name          = "%[1]s"
}
`, zoneName)
}

func testAccZoneConfig_forceDestroyTrailingPeriod(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  force_destroy = true
  name          = "%[1]s."
}
`, zoneName)
}

func testAccZoneConfig_tags1(zoneName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "%[1]s."

  tags = {
    %[2]q = %[3]q
  }
}
`, zoneName, tag1Key, tag1Value)
}

func testAccZoneConfig_tags2(zoneName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "%[1]s."

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, zoneName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccZoneConfig_vpcSingle(rName, zoneName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_zone" "test" {
  name = "%[2]s."

  vpc {
    vpc_id = aws_vpc.test1.id
  }
}
`, rName, zoneName)
}

func testAccZoneConfig_vpcMultiple(rName, zoneName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_zone" "test" {
  name = "%[2]s."

  vpc {
    vpc_id = aws_vpc.test1.id
  }

  vpc {
    vpc_id = aws_vpc.test2.id
  }
}
`, rName, zoneName)
}

func testAccZoneConfig_vpcSingle_forceDestroy(rName, zoneName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_zone" "test" {
  force_destroy = true
  name          = "%[2]s."

  vpc {
    vpc_id = aws_vpc.test1.id
  }
}
`, rName, zoneName)
}

func testAccZoneConfig_enableAcceleratedRecovery(zoneName1, zoneName2 string, enableAcceleratedRecovery bool) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test1" {
  name = "%[1]s."

  enable_accelerated_recovery = %[3]t
}
resource "aws_route53_zone" "test2" {
  name = "%[2]s."

  enable_accelerated_recovery = %[3]t
}
`, zoneName1, zoneName2, enableAcceleratedRecovery)
}
