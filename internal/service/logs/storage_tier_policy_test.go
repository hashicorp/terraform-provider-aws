// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsStorageTierPolicy_serial(t *testing.T) {
	t.Parallel()
	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccLogsStorageTierPolicy_basic,
		acctest.CtDisappears: testAccLogsStorageTierPolicy_disappears,
		"update":             testAccLogsStorageTierPolicy_update,
	}
	acctest.RunSerialTests1Level(t, testCases, 0)
}

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
		CheckDestroy:             testAccCheckStorageTierPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageTierPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageTierPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "INTELLIGENT_TIERING"),
				),
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
		CheckDestroy:             testAccCheckStorageTierPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageTierPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageTierPolicyExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflogs.ResourceStorageTierPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
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
		CheckDestroy:             testAccCheckStorageTierPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageTierPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageTierPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "INTELLIGENT_TIERING"),
				),
			},
			{
				Config: testAccStorageTierPolicyConfig_standard(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageTierPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "STANDARD"),
				),
			},
			{
				Config: testAccStorageTierPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageTierPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "INTELLIGENT_TIERING"),
				),
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

func testAccCheckStorageTierPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_storage_tier_policy" {
				continue
			}

			// Policy always exists, verify it's reset to STANDARD after destroy
			out, err := tflogs.FindStorageTierPolicy(ctx, conn)
			if err != nil {
				return err
			}

			if out.StorageTier != awstypes.StorageTierStandard {
				return fmt.Errorf("CloudWatch Logs Storage Tier Policy not reset to STANDARD after destroy, got: %s", out.StorageTier)
			}
		}

		return nil
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

func testAccStorageTierPolicyConfig_basic() string {
	return `
resource "aws_cloudwatch_log_storage_tier_policy" "test" {
  storage_tier = "INTELLIGENT_TIERING"
}
`
}

func testAccStorageTierPolicyConfig_standard() string {
	return `
resource "aws_cloudwatch_log_storage_tier_policy" "test" {
  storage_tier = "STANDARD"
}
`
}
