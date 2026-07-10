// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53Record_List_fullName(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route53_record.test[0]"
	resourceName2 := "aws_route53_record.test[1]"

	zoneName := acctest.RandomDomain(t).String()
	subdomainName := acctest.RandString(t, 8)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Record/list_fullname/"),
				ConfigVariables: config.Variables{
					"subdomainName":  config.StringVariable(subdomainName),
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("fqdn"), knownvalue.StringExact(subdomainName+"-0."+zoneName)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(subdomainName+"-0."+zoneName)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrType), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("zone_id"), knownvalue.NotNull()),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("fqdn"), knownvalue.StringExact(subdomainName+"-1."+zoneName)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrName), knownvalue.StringExact(subdomainName+"-1."+zoneName)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrType), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Record/list_fullname/"),
				ConfigVariables: config.Variables{
					"subdomainName":  config.StringVariable(subdomainName),
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_route53_record.test", map[string]knownvalue.Check{
						"zone_id":           knownvalue.NotNull(),
						names.AttrName:      knownvalue.StringExact(subdomainName + "-0." + zoneName),
						names.AttrType:      knownvalue.StringExact("A"),
						"set_identifier":    knownvalue.Null(),
						names.AttrAccountID: knownvalue.NotNull(),
					}),
					querycheck.ExpectResourceDisplayName("aws_route53_record.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(subdomainName+"-0."+zoneName+" A")),

					querycheck.ExpectIdentity("aws_route53_record.test", map[string]knownvalue.Check{
						"zone_id":           knownvalue.NotNull(),
						names.AttrName:      knownvalue.StringExact(subdomainName + "-1." + zoneName),
						names.AttrType:      knownvalue.StringExact("A"),
						"set_identifier":    knownvalue.Null(),
						names.AttrAccountID: knownvalue.NotNull(),
					}),
					querycheck.ExpectResourceDisplayName("aws_route53_record.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(subdomainName+"-1."+zoneName+" A")),
				},
			},
		},
	})
}

func TestAccRoute53Record_List_fullName_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_route53_record.test[0]"

	zoneName := acctest.RandomDomain(t).String()
	subdomainName := acctest.RandString(t, 8)

	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Record/list_fullname_include_resource/"),
				ConfigVariables: config.Variables{
					"subdomainName":  config.StringVariable(subdomainName),
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("fqdn"), knownvalue.StringExact(subdomainName+"-0."+zoneName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(subdomainName+"-0."+zoneName)),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Record/list_fullname_include_resource/"),
				ConfigVariables: config.Variables{
					"subdomainName":  config.StringVariable(subdomainName),
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route53_record.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route53_record.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact(subdomainName+"-0."+zoneName+" A")),
					querycheck.ExpectResourceKnownValues("aws_route53_record.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAlias), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("cidr_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("failover_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("fqdn"), knownvalue.StringExact(subdomainName+"-0."+zoneName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("geolocation_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("geoproximity_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("health_check_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("latency_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("multivalue_answer_routing_policy"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(subdomainName+"-0."+zoneName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("records"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("10.0.0.0"),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("set_identifier"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("weighted_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
					}),
				},
			},
		},
	})
}

func TestAccRoute53Record_List_shortName(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route53_record.test[0]"
	resourceName2 := "aws_route53_record.test[1]"

	zoneName := acctest.RandomDomain(t).String()
	subdomainName := acctest.RandString(t, 8)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Record/list_shortname/"),
				ConfigVariables: config.Variables{
					"subdomainName":  config.StringVariable(subdomainName),
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("fqdn"), knownvalue.StringExact(subdomainName+"-0."+zoneName)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(subdomainName+"-0")),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrType), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("zone_id"), knownvalue.NotNull()),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("fqdn"), knownvalue.StringExact(subdomainName+"-1."+zoneName)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrName), knownvalue.StringExact(subdomainName+"-1")),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrType), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Record/list_shortname/"),
				ConfigVariables: config.Variables{
					"subdomainName":  config.StringVariable(subdomainName),
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_route53_record.test", map[string]knownvalue.Check{
						"zone_id":           knownvalue.NotNull(),
						names.AttrName:      knownvalue.StringExact(subdomainName + "-0." + zoneName),
						names.AttrType:      knownvalue.StringExact("A"),
						"set_identifier":    knownvalue.Null(),
						names.AttrAccountID: knownvalue.NotNull(),
					}),
					querycheck.ExpectResourceDisplayName("aws_route53_record.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(subdomainName+"-0."+zoneName+" A")),

					querycheck.ExpectIdentity("aws_route53_record.test", map[string]knownvalue.Check{
						"zone_id":           knownvalue.NotNull(),
						names.AttrName:      knownvalue.StringExact(subdomainName + "-1." + zoneName),
						names.AttrType:      knownvalue.StringExact("A"),
						"set_identifier":    knownvalue.Null(),
						names.AttrAccountID: knownvalue.NotNull(),
					}),
					querycheck.ExpectResourceDisplayName("aws_route53_record.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(subdomainName+"-1."+zoneName+" A")),
				},
			},
		},
	})
}

func TestAccRoute53Record_List_shortName_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_route53_record.test[0]"

	zoneName := acctest.RandomDomain(t).String()
	subdomainName := acctest.RandString(t, 8)

	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Record/list_shortname_include_resource/"),
				ConfigVariables: config.Variables{
					"subdomainName":  config.StringVariable(subdomainName),
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("fqdn"), knownvalue.StringExact(subdomainName+"-0."+zoneName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(subdomainName+"-0")),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Record/list_shortname_include_resource/"),
				ConfigVariables: config.Variables{
					"subdomainName":  config.StringVariable(subdomainName),
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route53_record.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route53_record.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact(subdomainName+"-0."+zoneName+" A")),
					querycheck.ExpectResourceKnownValues("aws_route53_record.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAlias), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("cidr_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("failover_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("fqdn"), knownvalue.StringExact(subdomainName+"-0."+zoneName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("geolocation_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("geoproximity_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("health_check_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("latency_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("multivalue_answer_routing_policy"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(subdomainName+"-0."+zoneName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("records"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("10.0.0.0"),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("set_identifier"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("weighted_routing_policy"), knownvalue.ListExact([]knownvalue.Check{})),
					}),
				},
			},
		},
	})
}
