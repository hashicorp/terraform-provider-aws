// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fis_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tffis "github.com/hashicorp/terraform-provider-aws/internal/service/fis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// The FIS safety lever is an account/region singleton: it always exists, so no two tests in this
// package can be allowed to mutate it concurrently - acctest.ParallelTest's default t.Parallel()
// behavior would let TestAccFISSafetyLeverState_basic and _update race on the same "default"
// lever, each seeing the other's in-flight changes. RunSerialTests1Level forces them to run one
// at a time, exactly like TestAccEC2SerialConsoleAccess_serial does for the analogous singleton
// EC2 setting.
func TestAccFISSafetyLeverState_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"basic":                   testAccFISSafetyLeverState_basic,
		"update":                  testAccFISSafetyLeverState_update,
		"Identity_basic":          testAccFISSafetyLeverState_Identity_basic,
		"Identity_regionOverride": testAccFISSafetyLeverState_Identity_regionOverride,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

// AWS's UpdateSafetyLeverState rejects any call whose requested status already matches the
// account's live status ("Cannot update Safety Lever reason without a status change"), so
// testAccSafetyLeverStateOppositeStatus reads the live status first and the initial apply is
// always a genuine transition.
func testAccFISSafetyLeverState_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_fis_safety_lever_state.test"
	startStatus := testAccSafetyLeverStateOppositeStatus(ctx, t, acctest.Region())

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSafetyLeverStateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSafetyLeverStateConfig_basic(startStatus, "Managed by Terraform acceptance test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "state.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", startStatus),
					resource.TestCheckResourceAttr(resourceName, "state.0.reason", "Managed by Terraform acceptance test"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "fis", "safety-lever/default"),
				),
			},
			{
				// Always leave the account's safety lever disengaged when the test finishes. A
				// no-op (no diff, no API call) if startStatus was already "disengaged".
				Config: testAccSafetyLeverStateConfig_basic("disengaged", "Managed by Terraform acceptance test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", "disengaged"),
				),
			},
			{
				// The safety lever has no "id" attribute; import resolves it from resource
				// identity ({account_id, region}), and verify matches the single instance on arn.
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccFISSafetyLeverState_update(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_fis_safety_lever_state.test"
	startStatus := testAccSafetyLeverStateOppositeStatus(ctx, t, acctest.Region())
	otherStatus := "engaged"
	if startStatus == "engaged" {
		otherStatus = "disengaged"
	}
	// If otherStatus is already "disengaged", the final cleanup step below is a same-status no-op,
	// so its reason must match what step 2 actually applied - AWS rejects a reason-only change.
	cleanupReason := "Managed by Terraform acceptance test"
	if otherStatus == "disengaged" {
		cleanupReason = "Blocked for scheduled maintenance"
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSafetyLeverStateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSafetyLeverStateConfig_basic(startStatus, "Managed by Terraform acceptance test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", startStatus),
				),
			},
			{
				Config: testAccSafetyLeverStateConfig_basic(otherStatus, "Blocked for scheduled maintenance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", otherStatus),
					resource.TestCheckResourceAttr(resourceName, "state.0.reason", "Blocked for scheduled maintenance"),
				),
			},
			{
				// Always leave the account's safety lever disengaged when the test finishes. If
				// otherStatus was already "disengaged" this must be a true no-op (same status AND
				// same reason as the prior step) - reusing a different reason string here would
				// itself be an invalid reason-only change, since status wouldn't be transitioning.
				Config: testAccSafetyLeverStateConfig_basic("disengaged", cleanupReason),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", "disengaged"),
				),
			},
		},
	})
}

// There is no DeleteSafetyLever API and the resource's Delete is a no-op (framework.WithNoOpDelete),
// so there is nothing to verify here beyond the state being removed from Terraform.
func testAccCheckSafetyLeverStateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return nil
	}
}

func testAccCheckSafetyLeverStateExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[name]; !ok {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameSafetyLeverState, name, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).FISClient(ctx)

		// The safety lever is a singleton addressed by the fixed literal "default"; the resource
		// carries no "id" attribute (identity is {account_id, region}).
		if _, err := tffis.FindSafetyLever(ctx, conn, "default"); err != nil {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameSafetyLeverState, name, err)
		}

		return nil
	}
}

// testAccSafetyLeverStateOppositeStatus reads the live safety lever status in the given Region
// using a standalone SDK client (rather than acctest.ProviderMeta, which is not guaranteed
// configured this early - before resource.ParallelTest has run) and returns the opposite value,
// so the caller's first apply is guaranteed to be a genuine status transition.
func testAccSafetyLeverStateOppositeStatus(ctx context.Context, t *testing.T, region string) string {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		t.Fatalf("loading AWS config: %s", err)
	}

	conn := fis.NewFromConfig(cfg)
	current, err := tffis.FindSafetyLever(ctx, conn, "default")
	if err != nil {
		t.Fatalf("reading FIS safety lever: %s", err)
	}

	if string(current.State.Status) == "engaged" {
		return "disengaged"
	}
	return "engaged"
}

func testAccSafetyLeverStateConfig_basic(status, reason string) string {
	return `
resource "aws_fis_safety_lever_state" "test" {
  state {
    status = "` + status + `"
    reason = "` + reason + `"
  }
}
`
}

func testAccSafetyLeverStateConfig_region(region, status, reason string) string {
	return `
resource "aws_fis_safety_lever_state" "test" {
  region = "` + region + `"

  state {
    status = "` + status + `"
    reason = "` + reason + `"
  }
}
`
}
