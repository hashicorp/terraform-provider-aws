// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

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

func TestAccDynamoDBGlobalSecondaryIndex_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_dynamodb_global_secondary_index.test[0]"
	resourceName2 := "aws_dynamodb_global_secondary_index.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/GlobalSecondaryIndex/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/GlobalSecondaryIndex/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_dynamodb_global_secondary_index.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_dynamodb_global_secondary_index.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_dynamodb_global_secondary_index.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_dynamodb_global_secondary_index.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_dynamodb_global_secondary_index.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_dynamodb_global_secondary_index.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_dynamodb_global_secondary_index.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/GlobalSecondaryIndex/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/GlobalSecondaryIndex/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_dynamodb_global_secondary_index.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_dynamodb_global_secondary_index.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_dynamodb_global_secondary_index.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("dynamodb", "table/"+rName+"/index/"+rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("index_name"), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"attribute_name": knownvalue.StringExact(rName + "-0"),
								"attribute_type": knownvalue.StringExact("S"),
								"key_type":       knownvalue.StringExact("HASH"),
							}),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("projection"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"non_key_attributes": knownvalue.Null(),
								"projection_type":    knownvalue.StringExact("ALL"),
							}),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"read_capacity_units":  knownvalue.Int64Exact(1),
								"write_capacity_units": knownvalue.Int64Exact(1),
							}),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTableName), knownvalue.StringExact(rName)),
					}),
				},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_dynamodb_global_secondary_index.test[0]"
	resourceName2 := "aws_dynamodb_global_secondary_index.test[1]"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/GlobalSecondaryIndex/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/GlobalSecondaryIndex/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_dynamodb_global_secondary_index.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_dynamodb_global_secondary_index.test", identity2.Checks()),
				},
			},
		},
	})
}
