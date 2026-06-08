// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAppBundle_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("customer_managed_key_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAppBundle_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfappfabric.ResourceAppBundle, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccAppBundle_cmk(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appfabric_app_bundle.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_cmk(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("customer_managed_key_arn"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAppBundle_upgradeFromV5(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.AppFabricServiceID),
		CheckDestroy: testAccCheckAppBundleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.92.0",
					},
				},
				Config: testAccAppBundleConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAppBundleConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
		},
	})
}

func testAccAppBundle_upgradeFromV5PlanRefreshFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.AppFabricServiceID),
		CheckDestroy: testAccCheckAppBundleDestroy(ctx, t),
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccAppBundleConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAppBundleConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
		},
	})
}

func testAccAppBundle_upgradeFromV5WithUpdatePlanRefreshFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.AppFabricServiceID),
		CheckDestroy: testAccCheckAppBundleDestroy(ctx, t),
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccAppBundleConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAppBundleConfig_tags1(acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
					})),
				},
			},
		},
	})
}

func testAccAppBundle_upgradeFromV5WithDefaultRegionRefreshFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.AppFabricServiceID),
		CheckDestroy: testAccCheckAppBundleDestroy(ctx, t),
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccAppBundleConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAppBundleConfig_region(acctest.Region()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
		},
	})
}

func testAccAppBundle_upgradeFromV5WithNewRegionRefreshFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.AppFabricServiceID),
		CheckDestroy: testAccCheckAppBundleDestroy(ctx, t),
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccAppBundleConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAppBundleConfig_region(endpoints.EuWest1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExistsInRegion(ctx, t, resourceName, &appbundle, endpoints.EuWest1RegionID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(endpoints.EuWest1RegionID)),
				},
			},
		},
	})
}

func testAccAppBundle_regionCreateNull(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_region("null"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppBundleConfig_region(acctest.Region()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppBundleConfig_region(endpoints.ApNortheast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExistsInRegion(ctx, t, resourceName, &appbundle, endpoints.ApNortheast1RegionID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(endpoints.ApNortheast1RegionID)),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(endpoints.ApNortheast1RegionID)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAppBundleRegionImportStateIDFunc(resourceName, endpoints.ApNortheast1RegionID),
			},
		},
	})
}

func testAccAppBundle_regionCreateNonNull(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_region(endpoints.EuWest1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExistsInRegion(ctx, t, resourceName, &appbundle, endpoints.EuWest1RegionID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(endpoints.EuWest1RegionID)),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(endpoints.EuWest1RegionID)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAppBundleRegionImportStateIDFunc(resourceName, endpoints.EuWest1RegionID),
			},
			{
				Config: testAccAppBundleConfig_region(acctest.Region()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, t, resourceName, &appbundle),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAppBundleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_app_bundle" {
				continue
			}

			_, err := tfappfabric.FindAppBundleByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Invalid Resource Identifier") {
				// This can happen when a per-resource Region override is in effect.
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppFabric App Bundle %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppBundleExists(ctx context.Context, t *testing.T, n string, v *awstypes.AppBundle) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppFabricClient(ctx)

		output, err := tfappfabric.FindAppBundleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAppBundleExistsInRegion(ctx context.Context, t *testing.T, n string, v *awstypes.AppBundle, region string) resource.TestCheckFunc {
	// Push region into Context.
	ctx = conns.NewResourceContext(ctx, "AppFabric", "App Bundle", "aws_appfabric_app_bundle", region)
	return testAccCheckAppBundleExists(ctx, t, n, v)
}

func testAccAppBundleRegionImportStateIDFunc(n, region string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s@%s", rs.Primary.Attributes[names.AttrID], region), nil
	}
}

const testAccAppBundleConfig_basic = `
resource "aws_appfabric_app_bundle" "test" {}
`

func testAccAppBundleConfig_cmk(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_appfabric_app_bundle" "test" {
  customer_managed_key_arn = aws_kms_key.test.arn
}
`, rName)
}

func testAccAppBundleConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccAppBundleConfig_region(region string) string {
	if region != "null" {
		region = strconv.Quote(region)
	}

	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  region = %[1]s
}
`, region)
}
