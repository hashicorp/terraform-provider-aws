// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"fmt"
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

func TestAccOpenSearchServerlessVPCEndpoint_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_opensearchserverless_vpc_endpoint.test[0]"
	// VPC endpoint names must match [a-z][a-z0-9-]+ and are limited to 32
	// characters. Use a short random base (28 chars) so that appending
	// "-<index>" stays within the limit.
	// OpenSearch Serverless allows only one VPC endpoint per VPC, so resource_count is 1.
	rName := acctest.RandStringFromCharSet(t, 1, acctest.CharSetAlpha) + acctest.RandString(t, 27)
	displayName1 := fmt.Sprintf("%s-0", rName)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVPCEndpoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_basic/"),
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
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_opensearchserverless_vpc_endpoint.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_opensearchserverless_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(displayName1)),
					tfquerycheck.ExpectNoResourceObject("aws_opensearchserverless_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),
				},
			},
		},
	})
}

func TestAccOpenSearchServerlessVPCEndpoint_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_opensearchserverless_vpc_endpoint.test[0]"
	// VPC endpoint names must match [a-z][a-z0-9-]+ and are limited to 32
	// characters. Use a short random base (28 chars) so that appending
	// "-<index>" stays within the limit.
	rName := acctest.RandStringFromCharSet(t, 1, acctest.CharSetAlpha) + acctest.RandString(t, 27)
	displayName1 := fmt.Sprintf("%s-0", rName)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVPCEndpoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_include_resource/"),
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
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_opensearchserverless_vpc_endpoint.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_opensearchserverless_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(displayName1)),
					querycheck.ExpectResourceKnownValues("aws_opensearchserverless_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(displayName1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVPCID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSubnetIDs), knownvalue.SetSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSecurityGroupIDs), knownvalue.SetSizeExact(1)),
					}),
				},
			},
		},
	})
}

func TestAccOpenSearchServerlessVPCEndpoint_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_opensearchserverless_vpc_endpoint.test[0]"
	// VPC endpoint names must match [a-z][a-z0-9-]+ and are limited to 32
	// characters. Use a short random base (28 chars) so that appending
	// "-<index>" stays within the limit.
	// OpenSearch Serverless allows only one VPC endpoint per VPC, so resource_count is 1.
	rName := acctest.RandStringFromCharSet(t, 1, acctest.CharSetAlpha) + acctest.RandString(t, 27)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckVPCEndpoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_opensearchserverless_vpc_endpoint.test", identity1.Checks()),
				},
			},
		},
	})
}
