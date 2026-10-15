// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

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
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSVolumeAttachment_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_volume_attachment.test[0]"
	resourceName2 := "aws_volume_attachment.test[1]"

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
		CheckDestroy:             testAccCheckEBSVolumeAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/EBSVolumeAttachment/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New(names.AttrInstanceID), "aws_instance.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New(names.AttrDeviceName)),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("volume_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New(names.AttrInstanceID)),

					identity2.GetIdentity(resourceName2),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New(names.AttrInstanceID), "aws_instance.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New(names.AttrDeviceName)),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("volume_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New(names.AttrInstanceID)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/EBSVolumeAttachment/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_volume_attachment.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_volume_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^i-[a-f0-9]+ \(/dev/sd[h-l] - vol-[a-f0-9]+\)$`))),
					tfquerycheck.ExpectNoResourceObject("aws_volume_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_volume_attachment.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_volume_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^i-[a-f0-9]+ \(/dev/sd[h-l] - vol-[a-f0-9]+\)$`))),
					tfquerycheck.ExpectNoResourceObject("aws_volume_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeAttachment_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_volume_attachment.test[0]"

	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckEBSVolumeAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/EBSVolumeAttachment/list_include_resource/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrInstanceID), "aws_instance.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrDeviceName)),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("volume_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrInstanceID)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/EBSVolumeAttachment/list_include_resource/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_volume_attachment.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_volume_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^i-[a-f0-9]+ \(/dev/sd[h-l] - vol-[a-f0-9]+\)$`))),
					querycheck.ExpectResourceKnownValues("aws_volume_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDeviceName), knownvalue.StringExact("/dev/sdh")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrInstanceID), knownvalue.StringRegexp(regexache.MustCompile(`^i-[a-f0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("volume_id"), knownvalue.StringRegexp(regexache.MustCompile(`^vol-[a-f0-9]+$`))),
					}),
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeAttachment_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_volume_attachment.test[0]"
	resourceName2 := "aws_volume_attachment.test[1]"

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
		CheckDestroy:             testAccCheckEBSVolumeAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/EBSVolumeAttachment/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New(names.AttrInstanceID), "aws_instance.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New(names.AttrDeviceName)),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("volume_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New(names.AttrInstanceID)),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New(names.AttrInstanceID), "aws_instance.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New(names.AttrDeviceName)),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("volume_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New(names.AttrInstanceID)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/EBSVolumeAttachment/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_volume_attachment.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_volume_attachment.test", identity2.Checks()),
				},
			},
		},
	})
}
