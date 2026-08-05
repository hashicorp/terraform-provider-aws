// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccLogsStorageTierPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_log_storage_tier_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStorageTierPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/StorageTierPolicy/basic/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageTierPolicyExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_tier"), tfknownvalue.StringExact(awstypes.StorageTierIntelligentTiering)),
				},
			},
		},
	})
}

func testAccLogsStorageTierPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_log_storage_tier_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStorageTierPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/StorageTierPolicy/basic/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageTierPolicyExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflogs.ResourceStorageTierPolicy, resourceName), // nosemgrep:ci.semgrep.acctest.disappears-expect-resource-action
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccLogsStorageTierPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_log_storage_tier_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStorageTierPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/StorageTierPolicy/basic/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageTierPolicyExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_tier"), tfknownvalue.StringExact(awstypes.StorageTierIntelligentTiering)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/StorageTierPolicy/standard/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageTierPolicyExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_tier"), tfknownvalue.StringExact(awstypes.StorageTierStandard)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/StorageTierPolicy/basic/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageTierPolicyExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_tier"), tfknownvalue.StringExact(awstypes.StorageTierIntelligentTiering)),
				},
			},
		},
	})
}

func testAccPreCheckStorageTierPolicy(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

	_, err := conn.GetStorageTierPolicy(ctx, &cloudwatchlogs.GetStorageTierPolicyInput{})
	// Service-specific errors first
	if errs.IsA[*awstypes.AccessDeniedException](err) {
		t.Skipf("skipping acceptance testing: %s", err)
		return
	}
	// Generic skip conditions second
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
		return
	}
	// Unexpected errors last
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckStorageTierPolicyExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		_, err := tflogs.FindStorageTierPolicy(ctx, conn)

		return err
	}
}
