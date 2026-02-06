// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRRepository_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ecr_repository.test[0]"
	resourceName2 := "aws_ecr_repository.test[1]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	name1 := tfstatecheck.StateValue()
	name2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Repository/list_basic"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					name1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrName)),
					name2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrName)),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Repository/list_basic"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_ecr_repository.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrName:      name1.Value(),
					}),
					querycheck.ExpectIdentity("aws_ecr_repository.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrName:      name2.Value(),
					}),
				},
			},
		},
	})
}

func TestAccECRRepository_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	name := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ECRServiceID),
		CheckDestroy: testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Repository/list_region_override"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					name.GetStateValue("aws_ecr_repository.test", tfjsonpath.New(names.AttrName)),
				},
			},
			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Repository/list_region_override"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_ecr_repository.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrName:      name.Value(),
					}),
				},
			},
		},
	})
}
