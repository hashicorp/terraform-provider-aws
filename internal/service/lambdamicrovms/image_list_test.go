// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
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

func TestAccLambdaMicroVMsImage_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_lambdamicrovms_image.test[0]"
	resourceName2 := "aws_lambdamicrovms_image.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaMicroVMsServiceID),
		CheckDestroy:             testAccCheckImageDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Image/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), checkImageARN(rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), checkImageARN(rName+"-1")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Image/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lambdamicrovms_image.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lambdamicrovms_image.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_lambdamicrovms_image.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_lambdamicrovms_image.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lambdamicrovms_image.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_lambdamicrovms_image.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccLambdaMicroVMsImage_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambdamicrovms_image.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaMicroVMsServiceID),
		CheckDestroy:             testAccCheckImageDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Image/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Image/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lambdamicrovms_image.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lambdamicrovms_image.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_lambdamicrovms_image.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), checkImageARN(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrCreatedAt), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("latest_active_image_version"), knownvalue.StringExact("1.0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("latest_failed_image_version"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.MicrovmImageStateCreated)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1)})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1)})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("updated_at"), knownvalue.NotNull()),
					}),
				},
			},
		},
	})
}

func TestAccLambdaMicroVMsImage_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_lambdamicrovms_image.test[0]"
	resourceName2 := "aws_lambdamicrovms_image.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaMicroVMsServiceID),
		CheckDestroy:             testAccCheckImageDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Image/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), checkImageARNAlternateRegion(rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), checkImageARNAlternateRegion(rName+"-1")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Image/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lambdamicrovms_image.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_lambdamicrovms_image.test", identity2.Checks()),
				},
			},
		},
	})
}

func TestAccLambdaMicroVMsImage_List_nameFilter(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_lambdamicrovms_image.test[0]"
	resourceName2 := "aws_lambdamicrovms_image.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaMicroVMsServiceID),
		CheckDestroy:             testAccCheckImageDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Image/list_name_filter/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), checkImageARN(rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), checkImageARN(rName+"-1")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Image/list_name_filter/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lambdamicrovms_image.test", identity1.Checks()),
					tfquerycheck.ExpectNoIdentityFunc("aws_lambdamicrovms_image.test", identity2.Checks()),
				},
			},
		},
	})
}
