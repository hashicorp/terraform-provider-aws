// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatchomni "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchomni"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	checkOrganizationAccessGrantARN                = tfknownvalue.RegionalARNRegexp("cloudwatch", regexache.MustCompile(`organization-access-grant/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`))
	checkOrganizationAccessGrantARNAlternateRegion = tfknownvalue.RegionalARNAlternateRegionRegexp("cloudwatch", regexache.MustCompile(`organization-access-grant/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`))
	checkOrganizationAccessGrantID                 = knownvalue.StringRegexp(regexache.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`))
)

func testAccOrganizationAccessGrant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.OrganizationAccessGrant
	resourceName := "aws_cloudwatchomni_organization_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOrganizationAccessGrant(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationAccessGrant/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkOrganizationAccessGrantARN),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("domain_id"), "aws_cloudwatchomni_organization_domain.test", tfjsonpath.New("domain_id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("grant_id"), checkOrganizationAccessGrantID),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("grant_type"), tfknownvalue.StringExact(awstypes.AccessGrantTypeCustomerManaged)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("permission"), tfknownvalue.StringExact(awstypes.OrganizationGrantPermissionAdmin)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPrincipal), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"principal_attribute": knownvalue.ListSizeExact(0),
						"principal_id":        knownvalue.NotNull(),
						"principal_type":      tfknownvalue.StringExact(awstypes.OrganizationGrantPrincipalTypeIamRole),
					})})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrPrincipal).AtSliceIndex(0).AtMapKey("principal_id"), "aws_iam_role.grantee", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationAccessGrant/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "grant_id"),
				ImportStateVerifyIdentifierAttribute: "grant_id",
			},
		},
	})
}

func testAccOrganizationAccessGrant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.OrganizationAccessGrant
	resourceName := "aws_cloudwatchomni_organization_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOrganizationAccessGrant(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationAccessGrant/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAccessGrantExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatchomni.ResourceOrganizationAccessGrant, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationAccessGrant_name(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.OrganizationAccessGrant
	resourceName := "aws_cloudwatchomni_organization_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	name1, name2 := rName, acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOrganizationAccessGrant(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationAccessGrant/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtName:  config.StringVariable(name1),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(name1)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationAccessGrant/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtName:  config.StringVariable(name2),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(name2)),
				},
			},
		},
	})
}

func testAccOrganizationAccessGrant_principal(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.OrganizationAccessGrant
	resourceName := "aws_cloudwatchomni_organization_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOrganizationAccessGrant(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationAccessGrant/principal/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"principal_index": config.IntegerVariable(0),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPrincipal).AtSliceIndex(0).AtMapKey("principal_type"), tfknownvalue.StringExact(awstypes.OrganizationGrantPrincipalTypeIamRole)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrPrincipal).AtSliceIndex(0).AtMapKey("principal_id"), "aws_iam_role.principal[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationAccessGrant/principal/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"principal_index": config.IntegerVariable(1),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrPrincipal).AtSliceIndex(0).AtMapKey("principal_id"), "aws_iam_role.principal[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

// testAccPreCheckOrganizationAccessGrant skips unless the caller is the management account of an organization
// that has trusted access enabled for CloudWatch Omni and no CloudWatch Omni domain.
func testAccPreCheckOrganizationAccessGrant(ctx context.Context, t *testing.T) {
	acctest.PreCheckOrganizationManagementAccount(ctx, t)
	acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "observabilityadmin.amazonaws.com")
	testAccPreCheck(ctx, t)
}

func testAccCheckOrganizationAccessGrantDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatchomni_organization_access_grant" {
				continue
			}

			_, err := tfcloudwatchomni.FindOrganizationAccessGrantByID(ctx, conn, rs.Primary.Attributes["grant_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Omni Organization Access Grant %s still exists", rs.Primary.Attributes["grant_id"])
		}

		return nil
	}
}

func testAccCheckOrganizationAccessGrantExists(ctx context.Context, t *testing.T, n string, v *awstypes.OrganizationAccessGrant) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		output, err := tfcloudwatchomni.FindOrganizationAccessGrantByID(ctx, conn, rs.Primary.Attributes["grant_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
