// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2LoadBalancer_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lb.test[0]"
	resourceName2 := "aws_lb.test[1]"
	rName := acctest.ResourcePrefix + acctest.RandString(t, 16)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy: testAccCheckLoadBalancerDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/LoadBalancer/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/app/%s/[a-z0-9]{16}", rName+"-0")))),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/app/%s/[a-z0-9]{16}", rName+"-1")))),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/LoadBalancer/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lb.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lb.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_lb.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_lb.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lb.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_lb.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccELBV2LoadBalancer_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lb.test[0]"
	rName := acctest.ResourcePrefix + acctest.RandString(t, 16)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy: testAccCheckLoadBalancerDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/LoadBalancer/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/app/%s/[a-z0-9]{16}", rName+"-0")))),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/LoadBalancer/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lb.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lb.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_lb.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("access_logs"), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/app/%s/[a-z0-9]{16}", rName+"-0")))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("arn_suffix"), knownvalue.StringRegexp(regexache.MustCompile(fmt.Sprintf("^app/%s/[a-z0-9]{16}$", rName+"-0")))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("client_keep_alive"), knownvalue.Int32Exact(3600)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("connection_logs"), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("customer_owned_ipv4_pool"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("desync_mitigation_mode"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDNSName), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("dns_record_client_routing_policy"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("drop_invalid_header_fields"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_cross_zone_load_balancing"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_deletion_protection"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_http2"), knownvalue.Bool(true)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_tls_version_and_cipher_suite_headers"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_waf_fail_open"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_xff_client_port"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_zonal_shift"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enforce_security_group_inbound_rules_on_private_link_traffic"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("health_check_logs"), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("idle_timeout"), knownvalue.Int32Exact(30)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("internal"), knownvalue.Bool(true)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrIPAddressType), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipam_pools"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("load_balancer_type"), tfknownvalue.StringExact(awstypes.LoadBalancerTypeEnumApplication)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("minimum_load_balancer_capacity"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrNamePrefix), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("preserve_host_header"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("secondary_ips_auto_assigned_per_subnet"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSecurityGroups), knownvalue.SetSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("subnet_mapping"), knownvalue.SetSizeExact(2)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSubnets), knownvalue.SetSizeExact(2)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVPCID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("xff_header_processing_mode"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("zone_id"), knownvalue.NotNull()),
					}),
				},
			},
		},
	})
}

func TestAccELBV2LoadBalancer_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lb.test[0]"
	resourceName2 := "aws_lb.test[1]"
	rName := acctest.ResourcePrefix + acctest.RandString(t, 16)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy: testAccCheckLoadBalancerDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/LoadBalancer/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionRegexp("elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/app/%s/[a-z0-9]{16}", rName+"-0")))),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionRegexp("elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/app/%s/[a-z0-9]{16}", rName+"-1")))),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/LoadBalancer/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lb.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_lb.test", identity2.Checks()),
				},
			},
		},
	})
}
