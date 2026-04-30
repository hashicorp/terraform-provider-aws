// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arczonalshift_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/arczonalshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arczonalshift/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfarczonalshift "github.com/hashicorp/terraform-provider-aws/internal/service/arczonalshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccARCZonalShiftAutoshiftObserverNotificationStatus_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccARCZonalShiftAutoshiftObserverNotificationStatus_basic,
		acctest.CtDisappears: testAccARCZonalShiftAutoshiftObserverNotificationStatus_disappears,
		"update":             testAccARCZonalShiftAutoshiftObserverNotificationStatus_update,
		"Identity":           testAccARCZonalShiftAutoshiftObserverNotificationStatus_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccARCZonalShiftAutoshiftObserverNotificationStatus_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_arczonalshift_autoshift_observer_notification_status.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoshiftObserverNotificationStatusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoshiftObserverNotificationStatusConfig_basic(string(awstypes.AutoshiftObserverNotificationStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoshiftObserverNotificationStatusExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.AutoshiftObserverNotificationStatusEnabled)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
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

func testAccARCZonalShiftAutoshiftObserverNotificationStatus_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_arczonalshift_autoshift_observer_notification_status.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoshiftObserverNotificationStatusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoshiftObserverNotificationStatusConfig_basic(string(awstypes.AutoshiftObserverNotificationStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoshiftObserverNotificationStatusExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.AutoshiftObserverNotificationStatusEnabled)),
				),
			},
			{
				Config: testAccAutoshiftObserverNotificationStatusConfig_basic(string(awstypes.AutoshiftObserverNotificationStatusDisabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoshiftObserverNotificationStatusExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.AutoshiftObserverNotificationStatusDisabled)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccARCZonalShiftAutoshiftObserverNotificationStatus_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_arczonalshift_autoshift_observer_notification_status.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoshiftObserverNotificationStatusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoshiftObserverNotificationStatusConfig_basic(string(awstypes.AutoshiftObserverNotificationStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoshiftObserverNotificationStatusExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfarczonalshift.NewAutoshiftObserverNotificationStatusResource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAutoshiftObserverNotificationStatusDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ARCZonalShiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_arczonalshift_autoshift_observer_notification_status" {
				continue
			}

			out, err := findAutoshiftObserverNotificationStatus(ctx, conn)
			if err != nil {
				return create.Error(names.ARCZonalShift, create.ErrActionCheckingDestroyed, tfarczonalshift.ResNameAutoshiftObserverNotificationStatus, rs.Primary.ID, err)
			}

			if out.Status != awstypes.AutoshiftObserverNotificationStatusDisabled {
				return create.Error(names.ARCZonalShift, create.ErrActionCheckingDestroyed, tfarczonalshift.ResNameAutoshiftObserverNotificationStatus, rs.Primary.ID, errors.New("not disabled"))
			}
		}

		return nil
	}
}

func testAccCheckAutoshiftObserverNotificationStatusExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ARCZonalShift, create.ErrActionCheckingExistence, tfarczonalshift.ResNameAutoshiftObserverNotificationStatus, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ARCZonalShift, create.ErrActionCheckingExistence, tfarczonalshift.ResNameAutoshiftObserverNotificationStatus, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ARCZonalShiftClient(ctx)

		_, err := findAutoshiftObserverNotificationStatus(ctx, conn)
		if err != nil {
			return create.Error(names.ARCZonalShift, create.ErrActionCheckingExistence, tfarczonalshift.ResNameAutoshiftObserverNotificationStatus, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ARCZonalShiftClient(ctx)

	input := arczonalshift.GetAutoshiftObserverNotificationStatusInput{}

	_, err := conn.GetAutoshiftObserverNotificationStatus(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func findAutoshiftObserverNotificationStatus(ctx context.Context, conn *arczonalshift.Client) (*arczonalshift.GetAutoshiftObserverNotificationStatusOutput, error) {
	input := arczonalshift.GetAutoshiftObserverNotificationStatusInput{}

	out, err := conn.GetAutoshiftObserverNotificationStatus(ctx, &input)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, &retry.NotFoundError{}
	}

	return out, nil
}

func testAccAutoshiftObserverNotificationStatusConfig_basic(status string) string {
	return fmt.Sprintf(`
resource "aws_arczonalshift_autoshift_observer_notification_status" "test" {
  status = %[1]q
}
`, status)
}
