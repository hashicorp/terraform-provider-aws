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

func TestAccEC2EBSVolume_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ebs_volume.test[0]"
	resourceName2 := "aws_ebs_volume.test[1]"

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
		CheckDestroy:             testAccCheckVolumeDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Volume/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "volume/{id}"),

					identity2.GetIdentity(resourceName2),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "volume/{id}"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Volume/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ebs_volume.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ebs_volume.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), id1.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_ebs_volume.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_ebs_volume.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ebs_volume.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), id2.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_ebs_volume.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccEC2EBSVolume_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ebs_volume.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	id1 := tfstatecheck.StateValue()
	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVolumeDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Volume/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						"Name":         config.StringVariable(rName),
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "volume/{id}"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Volume/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						"Name":         config.StringVariable(rName),
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ebs_volume.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ebs_volume.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(fmt.Sprintf("^%s \\(vol-[a-z0-9]+\\)$", rName)))),
					querycheck.ExpectResourceKnownValues("aws_ebs_volume.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`volume/vol-[a-z0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAvailabilityZone), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrCreateTime), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrEncrypted), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), id1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrIOPS), knownvalue.Int32Exact(100)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrKMSKeyID), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("multi_attach_enabled"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("outpost_arn"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSize), knownvalue.Int32Exact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSnapshotID), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name":         knownvalue.StringExact(rName),
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name":         knownvalue.StringExact(rName),
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrThroughput), knownvalue.Int32Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrType), knownvalue.StringExact("gp2")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("volume_initialization_rate"), knownvalue.Int32Exact(0)),
					}),
				},
			},
		},
	})
}

func TestAccEC2EBSVolume_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ebs_volume.test[0]"
	resourceName2 := "aws_ebs_volume.test[1]"

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
		CheckDestroy:             testAccCheckVolumeDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Volume/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "volume/{id}"),

					identity2.GetIdentity(resourceName2),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "volume/{id}"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Volume/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ebs_volume.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ebs_volume.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), id1.ValueCheck()),

					tfquerycheck.ExpectIdentityFunc("aws_ebs_volume.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ebs_volume.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), id2.ValueCheck()),
				},
			},
		},
	})
}
