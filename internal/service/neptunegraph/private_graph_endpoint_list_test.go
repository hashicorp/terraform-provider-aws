// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptunegraph_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneGraphPrivateGraphEndpoint_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_neptunegraph_private_graph_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneGraphServiceID),
		CheckDestroy:             testAccCheckPrivateGraphEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/PrivateGraphEndpoint/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/PrivateGraphEndpoint/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_neptunegraph_private_graph_endpoint.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_neptunegraph_private_graph_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_neptunegraph_private_graph_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks())),
				},
			},
		},
	})
}

func TestAccNeptuneGraphPrivateGraphEndpoint_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_neptunegraph_private_graph_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneGraphServiceID),
		CheckDestroy:             testAccCheckPrivateGraphEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/PrivateGraphEndpoint/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/PrivateGraphEndpoint/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_neptunegraph_private_graph_endpoint.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_neptunegraph_private_graph_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccNeptuneGraphPrivateGraphEndpoint_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_neptunegraph_private_graph_endpoint.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneGraphServiceID),
		CheckDestroy:             testAccCheckPrivateGraphEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/PrivateGraphEndpoint/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/PrivateGraphEndpoint/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_neptunegraph_private_graph_endpoint.test", identity.Checks()),
				},
			},
		},
	})
}
