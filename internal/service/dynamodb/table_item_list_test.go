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
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBTableItem_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_dynamodb_table_item.test[0]"
	resourceName2 := "aws_dynamodb_table_item.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		CheckDestroy:             testAccCheckTableItemDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/TableItem/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("hash_key"), knownvalue.StringExact("TestTableHashKey")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("hash_key"), knownvalue.StringExact("TestTableHashKey")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TableItem/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_dynamodb_table_item.test", identity1.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_dynamodb_table_item.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_dynamodb_table_item.test", identity2.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_dynamodb_table_item.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccDynamoDBTableItem_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_dynamodb_table_item.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		CheckDestroy:             testAccCheckTableItemDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/TableItem/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("hash_key"), knownvalue.StringExact("TestTableHashKey")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TableItem/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_dynamodb_table_item.test", identity1.Checks()),
					querycheck.ExpectResourceKnownValues("aws_dynamodb_table_item.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTableName), knownvalue.StringExact(rName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("hash_key"), knownvalue.StringExact("TestTableHashKey")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("item"), knownvalue.NotNull()),
					}),
				},
			},
		},
	})
}

func TestAccDynamoDBTableItem_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_dynamodb_table_item.test[0]"
	resourceName2 := "aws_dynamodb_table_item.test[1]"
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
		CheckDestroy:             testAccCheckTableItemDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/TableItem/list_region_override/"),
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
				ConfigDirectory: config.StaticDirectory("testdata/TableItem/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_dynamodb_table_item.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_dynamodb_table_item.test", identity2.Checks()),
				},
			},
		},
	})
}
