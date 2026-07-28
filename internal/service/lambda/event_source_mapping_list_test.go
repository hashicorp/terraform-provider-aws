// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
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

func TestAccLambdaEventSourceMapping_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lambda_event_source_mapping.test[0]"
	resourceName2 := "aws_lambda_event_source_mapping.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	uuid1 := tfstatecheck.StateValue()
	uuid2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/EventSourceMapping/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					uuid1.GetStateValue(resourceName1, tfjsonpath.New("uuid")),
					identity2.GetIdentity(resourceName2),
					uuid2.GetStateValue(resourceName2, tfjsonpath.New("uuid")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/EventSourceMapping/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lambda_event_source_mapping.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lambda_event_source_mapping.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), uuid1.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_lambda_event_source_mapping.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),
					tfquerycheck.ExpectIdentityFunc("aws_lambda_event_source_mapping.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lambda_event_source_mapping.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), uuid2.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_lambda_event_source_mapping.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lambda_event_source_mapping.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	arn1 := tfstatecheck.StateValue()
	id1 := tfstatecheck.StateValue()
	uuid1 := tfstatecheck.StateValue()
	eventSourceARN1 := tfstatecheck.StateValue()
	functionARN1 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		CheckDestroy:             testAccCheckEventSourceMappingDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/EventSourceMapping/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					arn1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrARN)),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					uuid1.GetStateValue(resourceName1, tfjsonpath.New("uuid")),
					eventSourceARN1.GetStateValue(resourceName1, tfjsonpath.New("event_source_arn")),
					functionARN1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrFunctionARN)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/EventSourceMapping/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lambda_event_source_mapping.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lambda_event_source_mapping.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), uuid1.ValueCheck()),
					querycheck.ExpectResourceKnownValues("aws_lambda_event_source_mapping.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), arn1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("batch_size"), knownvalue.Int64Exact(10)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrEnabled), knownvalue.Bool(true)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("event_source_arn"), eventSourceARN1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrFunctionARN), functionARN1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("function_name"), functionARN1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), id1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("last_modified"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrState), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("uuid"), uuid1.ValueCheck()),
					}),
				},
			},
		},
	})
}

func TestAccLambdaEventSourceMapping_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lambda_event_source_mapping.test[0]"
	resourceName2 := "aws_lambda_event_source_mapping.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	uuid1 := tfstatecheck.StateValue()
	uuid2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/EventSourceMapping/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					uuid1.GetStateValue(resourceName1, tfjsonpath.New("uuid")),
					identity2.GetIdentity(resourceName2),
					uuid2.GetStateValue(resourceName2, tfjsonpath.New("uuid")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/EventSourceMapping/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lambda_event_source_mapping.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lambda_event_source_mapping.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), uuid1.ValueCheck()),
					tfquerycheck.ExpectIdentityFunc("aws_lambda_event_source_mapping.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lambda_event_source_mapping.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), uuid2.ValueCheck()),
				},
			},
		},
	})
}
