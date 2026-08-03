// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/compare"
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

func TestAccELBV2TargetGroupAttachment_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lb_target_group_attachment.test[0]"
	resourceName2 := "aws_lb_target_group_attachment.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("target_group_arn"), "aws_lb_target_group.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("target_id"), "aws_instance.test[0]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("target_group_arn")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("target_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New(names.AttrPort)),

					identity2.GetIdentity(resourceName2),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("target_group_arn"), "aws_lb_target_group.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("target_id"), "aws_instance.test[1]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("target_group_arn")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("target_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New(names.AttrPort)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lb_target_group_attachment.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lb_target_group_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^i-[a-f0-9]+ \(arn:aws[a-z-]*:elasticloadbalancing:[a-z0-9-]+:[0-9]{12}:targetgroup/.+/.+\)$`))),
					tfquerycheck.ExpectNoResourceObject("aws_lb_target_group_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_lb_target_group_attachment.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lb_target_group_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^i-[a-f0-9]+ \(arn:aws[a-z-]*:elasticloadbalancing:[a-z0-9-]+:[0-9]{12}:targetgroup/.+/.+\)$`))),
					tfquerycheck.ExpectNoResourceObject("aws_lb_target_group_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_lb_target_group_attachment.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("target_group_arn"), "aws_lb_target_group.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("target_id"), "aws_instance.test[0]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("target_group_arn")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("target_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrPort)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lb_target_group_attachment.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lb_target_group_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^i-[a-f0-9]+ \(arn:aws[a-z-]*:elasticloadbalancing:[a-z0-9-]+:[0-9]{12}:targetgroup/.+/.+\)$`))),
					querycheck.ExpectResourceKnownValues("aws_lb_target_group_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("target_group_arn"), tfknownvalue.RegionalARNRegexp("elasticloadbalancing", regexache.MustCompile(`targetgroup/.+/.+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("target_id"), knownvalue.StringRegexp(regexache.MustCompile(`^i-[a-f0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPort), knownvalue.Int64Exact(80)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAvailabilityZone), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("quic_server_id"), knownvalue.Null()),
					}),
				},
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lb_target_group_attachment.test[0]"
	resourceName2 := "aws_lb_target_group_attachment.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

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
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("target_group_arn"), "aws_lb_target_group.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("target_id"), "aws_instance.test[0]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("target_group_arn")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("target_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New(names.AttrPort)),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("target_group_arn"), "aws_lb_target_group.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("target_id"), "aws_instance.test[1]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("target_group_arn")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("target_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New(names.AttrPort)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lb_target_group_attachment.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_lb_target_group_attachment.test", identity2.Checks()),
				},
			},
		},
	})
}
