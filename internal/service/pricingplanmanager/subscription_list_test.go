// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pricingplanmanager_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/pricingplanmanager/types"
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

func TestAccPricingPlanManagerSubscription_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_pricingplanmanager_subscription.test[0]"
	resourceName2 := "aws_pricingplanmanager_subscription.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PricingPlanManagerServiceID),
		CheckDestroy:             testAccCheckSubscriptionDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), checkSubscriptionARN),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), checkSubscriptionARN),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_pricingplanmanager_subscription.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_pricingplanmanager_subscription.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), checkSubscriptionARN),
					tfquerycheck.ExpectNoResourceObject("aws_pricingplanmanager_subscription.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_pricingplanmanager_subscription.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_pricingplanmanager_subscription.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), checkSubscriptionARN),
					tfquerycheck.ExpectNoResourceObject("aws_pricingplanmanager_subscription.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccPricingPlanManagerSubscription_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_pricingplanmanager_subscription.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PricingPlanManagerServiceID),
		CheckDestroy:             testAccCheckSubscriptionDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), checkSubscriptionARN),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_pricingplanmanager_subscription.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_pricingplanmanager_subscription.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), checkSubscriptionARN),
					querycheck.ExpectResourceKnownValues("aws_pricingplanmanager_subscription.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), checkSubscriptionARN),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("etag"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("plan_family"), knownvalue.StringExact("CloudFront")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("plan_tier"), knownvalue.StringExact("FREE")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("resource_arns"), knownvalue.SetSizeExact(2)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.StatusActive)),
					}),
				},
			},
		},
	})
}
