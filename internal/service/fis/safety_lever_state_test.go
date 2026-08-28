// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fis_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fis/types"
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
		"basic":  testAccFISSafetyLeverState_basic,
		"update": testAccFISSafetyLeverState_update,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

// It also cannot simply declare a fixed starting status: AWS's UpdateSafetyLeverState rejects a
// call whenever the requested status already matches the account's live status ("Cannot update
// Safety Lever reason without a status change"), so the resource's Create deliberately errors in
// that case instead of silently no-op'ing. testAccSafetyLeverStateOppositeStatus reads the live
// status first so the initial apply is always a genuine transition.
func testAccFISSafetyLeverState_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var safetyLever awstypes.SafetyLever
	resourceName := "aws_fis_safety_lever_state.test"
	startStatus := testAccSafetyLeverStateOppositeStatus(ctx, t)

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
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName, &safetyLever),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, "default"),
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
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName, &safetyLever),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", "disengaged"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFISSafetyLeverState_update(t *testing.T) {
	ctx := acctest.Context(t)

	var before, after awstypes.SafetyLever
	resourceName := "aws_fis_safety_lever_state.test"
	startStatus := testAccSafetyLeverStateOppositeStatus(ctx, t)
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
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", startStatus),
				),
			},
			{
				Config: testAccSafetyLeverStateConfig_basic(otherStatus, "Blocked for scheduled maintenance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName, &after),
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
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", "disengaged"),
				),
			},
		},
	})
}

// There is no DeleteSafetyLever API and the resource's Delete implementation intentionally makes
// no AWS API call (see the comment on safetyLeverStateResource.Delete), so there is nothing to
// verify here beyond the state being removed from Terraform.
func testAccCheckSafetyLeverStateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return nil
	}
}

func testAccCheckSafetyLeverStateExists(ctx context.Context, t *testing.T, name string, safetyLever *awstypes.SafetyLever) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameSafetyLeverState, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameSafetyLeverState, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).FISClient(ctx)

		resp, err := tffis.FindSafetyLever(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameSafetyLeverState, rs.Primary.ID, err)
		}

		*safetyLever = *resp

		return nil
	}
}

// testAccSafetyLeverStateOppositeStatus reads the account's live safety lever status using a
// standalone SDK client (rather than acctest.ProviderMeta, which is not guaranteed configured
// this early - before resource.ParallelTest has run) and returns the opposite value, so the
// caller's first apply is guaranteed to be a genuine status transition.
func testAccSafetyLeverStateOppositeStatus(ctx context.Context, t *testing.T) string {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(acctest.Region()))
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
