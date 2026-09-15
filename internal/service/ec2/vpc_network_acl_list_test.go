// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
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

func TestAccVPCNetworkACL_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_network_acl.test[0]"
	resourceName2 := "aws_network_acl.test[1]"

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCNetworkACL/list_basic"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPCNetworkACL/list_basic"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_network_acl.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id1.ValueCheck(),
					}),
					querycheck.ExpectIdentity("aws_network_acl.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id2.ValueCheck(),
					}),
				},
			},
		},
	})
}

func TestAccVPCNetworkACL_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_network_acl.test[0]"
	resourceName2 := "aws_network_acl.test[1]"

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCNetworkACL/list_region_override"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPCNetworkACL/list_region_override"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_network_acl.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        id1.ValueCheck(),
					}),
					querycheck.ExpectIdentity("aws_network_acl.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        id2.ValueCheck(),
					}),
				},
			},
		},
	})
}

func TestAccVPCNetworkACL_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_network_acl.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()
	id := tfstatecheck.StateValue()
	vpcID := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCNetworkACL/list_include_resource"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					id.GetStateValue(resourceName, tfjsonpath.New(names.AttrID)),
					vpcID.GetStateValue("aws_vpc.test", tfjsonpath.New(names.AttrID)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPCNetworkACL/list_include_resource"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_network_acl.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_acl.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), id.ValueCheck()),
					querycheck.ExpectResourceKnownValues("aws_network_acl.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`network-acl/acl-.+`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("egress"), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), id.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ingress"), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrOwnerID), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSubnetIDs), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name": knownvalue.StringExact(rName + "-0"),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name": knownvalue.StringExact(rName + "-0"),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVPCID), vpcID.ValueCheck()),
					}),
				},
			},
		},
	})
}
