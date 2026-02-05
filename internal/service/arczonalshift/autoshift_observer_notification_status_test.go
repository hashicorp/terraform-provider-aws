// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arczonalshift_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/arczonalshift"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfarczonalshift "github.com/hashicorp/terraform-provider-aws/internal/service/arczonalshift"
)

func TestAccARCZonalShiftAutoshiftObserverNotificationStatus_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var status arczonalshift.GetAutoshiftObserverNotificationStatusOutput
	resourceName := "aws_arczonalshift_autoshift_observer_notification_status.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoshiftObserverNotificationStatusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoshiftObserverNotificationStatusConfig_basic("ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoshiftObserverNotificationStatusExists(ctx, t, resourceName, &status),
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
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

func TestAccARCZonalShiftAutoshiftObserverNotificationStatus_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var status arczonalshift.GetAutoshiftObserverNotificationStatusOutput
	resourceName := "aws_arczonalshift_autoshift_observer_notification_status.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoshiftObserverNotificationStatusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoshiftObserverNotificationStatusConfig_basic("ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoshiftObserverNotificationStatusExists(ctx, t, resourceName, &status),
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
				),
			},
			{
				Config: testAccAutoshiftObserverNotificationStatusConfig_basic("DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoshiftObserverNotificationStatusExists(ctx, t, resourceName, &status),
					resource.TestCheckResourceAttr(resourceName, "status", "DISABLED"),
				),
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

			// This is a singleton resource, so it can't be truly destroyed.
			// When deleted, it should be set to DISABLED.
			out, err := findAutoshiftObserverNotificationStatus(ctx, conn, rs.Primary.ID)
			if err != nil {
				return create.Error(names.ARCZonalShift, create.ErrActionCheckingDestroyed, tfarczonalshift.ResNameAutoshiftObserverNotificationStatus, rs.Primary.ID, err)
			}

			if out.Status != "DISABLED" {
				return create.Error(names.ARCZonalShift, create.ErrActionCheckingDestroyed, tfarczonalshift.ResNameAutoshiftObserverNotificationStatus, rs.Primary.ID, errors.New("not disabled"))
			}
		}

		return nil
	}
}

func testAccCheckAutoshiftObserverNotificationStatusExists(ctx context.Context, t *testing.T, name string, status *arczonalshift.GetAutoshiftObserverNotificationStatusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ARCZonalShift, create.ErrActionCheckingExistence, tfarczonalshift.ResNameAutoshiftObserverNotificationStatus, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ARCZonalShift, create.ErrActionCheckingExistence, tfarczonalshift.ResNameAutoshiftObserverNotificationStatus, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ARCZonalShiftClient(ctx)

		resp, err := findAutoshiftObserverNotificationStatus(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ARCZonalShift, create.ErrActionCheckingExistence, tfarczonalshift.ResNameAutoshiftObserverNotificationStatus, rs.Primary.ID, err)
		}

		*status = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ARCZonalShiftClient(ctx)

	input := &arczonalshift.GetAutoshiftObserverNotificationStatusInput{}

	_, err := conn.GetAutoshiftObserverNotificationStatus(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func findAutoshiftObserverNotificationStatus(ctx context.Context, conn *arczonalshift.Client, id string) (*arczonalshift.GetAutoshiftObserverNotificationStatusOutput, error) {
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
	return `
resource "aws_arczonalshift_autoshift_observer_notification_status" "test" {
  status = "` + status + `"
}
`
}
