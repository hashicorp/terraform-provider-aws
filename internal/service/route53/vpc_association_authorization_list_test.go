// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func TestAccRoute53VPCAssociationAuthorization_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route53_vpc_association_authorization.test[0]"
	resourceName2 := "aws_route53_vpc_association_authorization.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	providers := make(map[string]*schema.Provider)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	resourceID1 := tfstatecheck.StateValue()
	resourceID2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.Route53ServiceID),
		CheckDestroy: testAccCheckVPCAssociationAuthorizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
				ConfigDirectory:          config.StaticDirectory("testdata/VPCAssociationAuthorization/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
					resourceID1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					resourceID2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
				},
			},

			// Step 2: Query
			{
				ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
				Query:                    true,
				ConfigDirectory:          config.StaticDirectory("testdata/VPCAssociationAuthorization/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route53_vpc_association_authorization.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route53_vpc_association_authorization.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), resourceID1.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_route53_vpc_association_authorization.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_route53_vpc_association_authorization.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route53_vpc_association_authorization.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), resourceID2.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_route53_vpc_association_authorization.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccRoute53VPCAssociationAuthorization_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route53_vpc_association_authorization.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	providers := make(map[string]*schema.Provider)

	identity1 := tfstatecheck.Identity()
	vpcID1 := tfstatecheck.StateValue()
	zoneID := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.Route53ServiceID),
		CheckDestroy: testAccCheckVPCAssociationAuthorizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
				ConfigDirectory:          config.StaticDirectory("testdata/VPCAssociationAuthorization/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					vpcID1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrVPCID)),
					zoneID.GetStateValue(resourceName1, tfjsonpath.New("zone_id")),
				},
			},

			// Step 2: Query
			{
				ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
				Query:                    true,
				ConfigDirectory:          config.StaticDirectory("testdata/VPCAssociationAuthorization/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route53_vpc_association_authorization.test", identity1.Checks()),
					querycheck.ExpectResourceKnownValues("aws_route53_vpc_association_authorization.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("zone_id"), zoneID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVPCID), vpcID1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("vpc_region"), knownvalue.StringExact(acctest.Region())),
					}),
				},
			},
		},
	})
}
