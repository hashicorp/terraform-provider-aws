// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

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

func TestAccRoute53ZoneAssociation_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route53_zone_association.test[0]"
	resourceName2 := "aws_route53_zone_association.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		CheckDestroy:             testAccCheckZoneAssociationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ZoneAssociation/list_basic/"),
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
				ConfigDirectory: config.StaticDirectory("testdata/ZoneAssociation/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route53_zone_association.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route53_zone_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(fmt.Sprintf("%s-0.example.com", rName))),
					tfquerycheck.ExpectNoResourceObject("aws_route53_zone_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_route53_zone_association.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route53_zone_association.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(fmt.Sprintf("%s-1.example.com", rName))),
					tfquerycheck.ExpectNoResourceObject("aws_route53_zone_association.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccRoute53ZoneAssociation_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route53_zone_association.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	vpcID := tfstatecheck.StateValue()
	zoneID := tfstatecheck.StateValue()
	resourceID := tfstatecheck.StateValue()
	owningAccount := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		CheckDestroy:             testAccCheckZoneAssociationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ZoneAssociation/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					vpcID.GetStateValue(resourceName1, tfjsonpath.New(names.AttrVPCID)),
					zoneID.GetStateValue(resourceName1, tfjsonpath.New("zone_id")),
					resourceID.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					owningAccount.GetStateValue(resourceName1, tfjsonpath.New("owning_account")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ZoneAssociation/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route53_zone_association.test", identity1.Checks()),
					querycheck.ExpectResourceKnownValues("aws_route53_zone_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), resourceID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("zone_id"), zoneID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVPCID), vpcID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("vpc_region"), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("owning_account"), owningAccount.ValueCheck()),
					}),
				},
			},
		},
	})
}
