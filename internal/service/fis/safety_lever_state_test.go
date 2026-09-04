// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tffis "github.com/hashicorp/terraform-provider-aws/internal/service/fis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	safetyLeverStatusEngaged    = string(awstypes.SafetyLeverStatusInputEngaged)
	safetyLeverStatusDisengaged = string(awstypes.SafetyLeverStatusInputDisengaged)
)

func TestAccFISSafetyLeverState_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:           testAccFISSafetyLeverState_basic,
		"update":                  testAccFISSafetyLeverState_update,
		"Identity_basic":          testAccFISSafetyLeverState_Identity_basic,
		"Identity_regionOverride": testAccFISSafetyLeverState_Identity_regionOverride,
		"List_basic":              testAccFISSafetyLeverState_List_basic,
		"List_includeResource":    testAccFISSafetyLeverState_List_includeResource,
		"List_regionOverride":     testAccFISSafetyLeverState_List_regionOverride,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

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
				Config: testAccSafetyLeverStateConfig_basic(safetyLeverStatusDisengaged, "Managed by Terraform acceptance test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", safetyLeverStatusDisengaged),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrRegion),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccFISSafetyLeverState_update(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_fis_safety_lever_state.test"
	startStatus := testAccSafetyLeverStateOppositeStatus(ctx, t, acctest.Region())
	otherStatus := safetyLeverStatusEngaged
	if startStatus == safetyLeverStatusEngaged {
		otherStatus = safetyLeverStatusDisengaged
	}

	cleanupReason := "Managed by Terraform acceptance test"
	if otherStatus == safetyLeverStatusDisengaged {
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", otherStatus),
					resource.TestCheckResourceAttr(resourceName, "state.0.reason", "Blocked for scheduled maintenance"),
				),
			},
			{
				Config: testAccSafetyLeverStateConfig_basic(safetyLeverStatusDisengaged, cleanupReason),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", safetyLeverStatusDisengaged),
				),
			},
		},
	})
}

// Delete is no-op
func testAccCheckSafetyLeverStateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return nil
	}
}

func testAccCheckSafetyLeverStateExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[n]; !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).FISClient(ctx)

		_, err := tffis.FindSafetyLever(ctx, conn, "default")

		return err
	}
}

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

	if current.State.Status == awstypes.SafetyLeverStatusEngaged {
		return safetyLeverStatusDisengaged
	}
	return safetyLeverStatusEngaged
}

func testAccSafetyLeverStateConfig_basic(status, reason string) string {
	return fmt.Sprintf(`
resource "aws_fis_safety_lever_state" "test" {
  state {
    status = %[1]q
    reason = %[2]q
  }
}
`, status, reason)
}

func testAccSafetyLeverStateConfig_region(region, status, reason string) string {
	return fmt.Sprintf(`
resource "aws_fis_safety_lever_state" "test" {
  region = %[1]q

  state {
    status = %[2]q
    reason = %[3]q
  }
}
`, region, status, reason)
}
