// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda_test

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
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaFunctionScalingConfig_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lambda_function_scaling_config.test[0]"
	resourceName2 := "aws_lambda_function_scaling_config.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		CheckDestroy:             testAccCheckFunctionScalingConfigDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionScalingConfig/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("function_scaling_config").AtSliceIndex(0).AtMapKey("min_execution_environments"), knownvalue.Int32Exact(3)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("function_scaling_config").AtSliceIndex(0).AtMapKey("max_execution_environments"), knownvalue.Int32Exact(100)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/FunctionScalingConfig/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_lambda_function_scaling_config.test", map[string]knownvalue.Check{
						"function_name":     knownvalue.StringExact(rName + "-0"),
						"qualifier":         knownvalue.StringExact("$LATEST.PUBLISHED"),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
					querycheck.ExpectIdentity("aws_lambda_function_scaling_config.test", map[string]knownvalue.Check{
						"function_name":     knownvalue.StringExact(rName + "-1"),
						"qualifier":         knownvalue.StringExact("$LATEST.PUBLISHED"),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
				},
			},
		},
	})
}

func TestAccLambdaFunctionScalingConfig_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lambda_function_scaling_config.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		CheckDestroy:             testAccCheckFunctionScalingConfigDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionScalingConfig/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("function_scaling_config").AtSliceIndex(0).AtMapKey("min_execution_environments"), knownvalue.Int32Exact(3)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/FunctionScalingConfig/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lambda_function_scaling_config.test", identity1.Checks()),
					querycheck.ExpectResourceKnownValues("aws_lambda_function_scaling_config.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("function_arn"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("function_name"), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("function_scaling_config").AtSliceIndex(0).AtMapKey("min_execution_environments"), knownvalue.Int32Exact(3)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("function_scaling_config").AtSliceIndex(0).AtMapKey("max_execution_environments"), knownvalue.Int32Exact(100)),
					}),
				},
			},
		},
	})
}

func TestAccLambdaFunctionScalingConfig_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lambda_function_scaling_config.test[0]"
	resourceName2 := "aws_lambda_function_scaling_config.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionScalingConfig/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/FunctionScalingConfig/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_lambda_function_scaling_config.test", map[string]knownvalue.Check{
						"function_name":     knownvalue.StringExact(rName + "-0"),
						"qualifier":         knownvalue.StringExact("$LATEST.PUBLISHED"),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
					}),
					querycheck.ExpectIdentity("aws_lambda_function_scaling_config.test", map[string]knownvalue.Check{
						"function_name":     knownvalue.StringExact(rName + "-1"),
						"qualifier":         knownvalue.StringExact("$LATEST.PUBLISHED"),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
					}),
				},
			},
		},
	})
}
