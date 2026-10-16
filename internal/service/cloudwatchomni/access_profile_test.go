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
	checkAccessProfileARN                = tfknownvalue.RegionalARNRegexp("cloudwatch", regexache.MustCompile(`access-profile/cw-ap-[0-9]{12}-[0-9a-z-]+`))
	checkAccessProfileARNAlternateRegion = tfknownvalue.RegionalARNAlternateRegionRegexp("cloudwatch", regexache.MustCompile(`access-profile/cw-ap-[0-9]{12}-[0-9a-z-]+`))
	checkAccessProfileID                 = knownvalue.StringRegexp(regexache.MustCompile(`^cw-ap-[0-9]{12}-[0-9a-z-]+$`))
)

func testAccAccessProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessProfile
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_access_profile.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessProfile/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessProfileExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkAccessProfileARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("assume_status"), tfknownvalue.StringExact(awstypes.AssumeStatusAllowed)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("profile_id"), checkAccessProfileID),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("profile_type"), tfknownvalue.StringExact(awstypes.AccessProfileTypeCustomerManaged)),
					statecheck.CompareValuePairs(
						resourceName, tfjsonpath.New("space_id"),
						"aws_cloudwatchomni_space.test", tfjsonpath.New("space_id"),
						compare.ValuesSame(),
					),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessProfile/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccAccessProfileImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "profile_id",
			},
		},
	})
}

func testAccAccessProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessProfile
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_access_profile.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessProfile/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessProfileExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatchomni.ResourceAccessProfile, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccAccessProfile_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessProfile
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_access_profile.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessProfile/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("description1"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessProfileExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description1")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessProfile/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("description1"),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccAccessProfileImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "profile_id",
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessProfile/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("description2"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessProfileExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description2")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessProfile/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessProfileExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccAccessProfile_name(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessProfile
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_access_profile.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessProfile/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtName:  config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessProfileExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessProfile/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtName:  config.StringVariable(rName2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessProfileExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName2)),
				},
			},
		},
	})
}

func testAccCheckAccessProfileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatchomni_access_profile" {
				continue
			}

			_, err := tfcloudwatchomni.FindAccessProfileByTwoPartKey(ctx, conn, rs.Primary.Attributes["space_id"], rs.Primary.Attributes["profile_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Omni Access Profile %s still exists", rs.Primary.Attributes["profile_id"])
		}

		return nil
	}
}

func testAccCheckAccessProfileExists(ctx context.Context, t *testing.T, n string, v *awstypes.AccessProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		output, err := tfcloudwatchomni.FindAccessProfileByTwoPartKey(ctx, conn, rs.Primary.Attributes["space_id"], rs.Primary.Attributes["profile_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAccessProfileImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "space_id", "profile_id")
}
