// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccVPCNetworkInterface_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_network_interface.test[0]"
	resourceName2 := "aws_network_interface.test[1]"

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckNetworkInterfaceDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/NetworkInterface/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "network-interface/{id}"),

					identity2.GetIdentity(resourceName2),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "network-interface/{id}"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/NetworkInterface/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_network_interface.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_interface.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), id1.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_network_interface.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_network_interface.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_interface.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), id2.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_network_interface.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccVPCNetworkInterface_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_network_interface.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	id1 := tfstatecheck.StateValue()
	identity1 := tfstatecheck.Identity()
	subnetID := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckNetworkInterfaceDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/NetworkInterface/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					subnetID.GetStateValue("aws_subnet.test", tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "network-interface/{id}"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/NetworkInterface/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_network_interface.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_interface.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(fmt.Sprintf("^%s \\(eni-[a-z0-9]+\\)$", rName)))),
					querycheck.ExpectResourceKnownValues("aws_network_interface.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`network-interface/eni-[a-z0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("attachment"), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("")),
						// enable_primary_ipv6, ipv6_address_list_enabled, and private_ip_list_enabled are Terraform-side
						// configuration flags that are not derived from the AWS API response, so they are not populated
						// by the resource's flatten function and remain null for a freshly listed resource.
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_primary_ipv6"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ena_srd_specification"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("interface_type"), knownvalue.StringExact("interface")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv4_prefixes"), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv4_prefix_count"), knownvalue.Int32Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_address_count"), knownvalue.Int32Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_address_list"), knownvalue.ListExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_address_list_enabled"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_addresses"), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_prefixes"), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_prefix_count"), knownvalue.Int32Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("mac_address"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrOutpostARN), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrOwnerID), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("private_dns_name"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("private_ip"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("private_ips"), knownvalue.SetExact([]knownvalue.Check{knownvalue.NotNull()})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("private_ips_count"), knownvalue.Int32Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("private_ip_list"), knownvalue.ListExact([]knownvalue.Check{knownvalue.NotNull()})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("private_ip_list_enabled"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSecurityGroups), knownvalue.SetExact([]knownvalue.Check{knownvalue.NotNull()})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("source_dest_check"), knownvalue.Bool(true)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSubnetID), subnetID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), id1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name":         knownvalue.StringExact(rName),
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name":         knownvalue.StringExact(rName),
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
					}),
				},
			},
		},
	})
}

func TestAccVPCNetworkInterface_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_network_interface.test[0]"
	resourceName2 := "aws_network_interface.test[1]"

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()

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
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckNetworkInterfaceDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/NetworkInterface/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "network-interface/{id}"),

					identity2.GetIdentity(resourceName2),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "network-interface/{id}"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/NetworkInterface/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_network_interface.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_interface.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), id1.ValueCheck()),

					tfquerycheck.ExpectIdentityFunc("aws_network_interface.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_interface.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), id2.ValueCheck()),
				},
			},
		},
	})
}
