// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestTrackerAssociationParseID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected tflocation.TrackerAssociationID
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: tflocation.TrackerAssociationID{},
			Error:    true,
		},
		{
			TestName: "no pipe",
			Input:    "trackerNameConsumerARN",
			Expected: tflocation.TrackerAssociationID{},
			Error:    true,
		},
		{
			TestName: "valid",
			Input:    "trackerName|consumerARN",
			Expected: tflocation.TrackerAssociationID{
				TrackerName: "trackerName",
				ConsumerARN: "consumerARN",
			},
			Error: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := tflocation.TrackerAssociationParseID(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAccLocationTrackerAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_tracker_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "consumer_arn", "aws_location_geofence_collection.test", "collection_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "tracker_name", "aws_location_tracker.test", "tracker_name"),
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

func TestAccLocationTrackerAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_tracker_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflocation.ResourceTrackerAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTrackerAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LocationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_location_tracker_association" {
				continue
			}

			trackerAssociationId, err := tflocation.TrackerAssociationParseID(rs.Primary.ID)

			if err != nil {
				return create.Error(names.Location, create.ErrActionCheckingDestroyed, tflocation.ResNameTrackerAssociation, rs.Primary.ID, err)
			}

			err = tflocation.FindTrackerAssociationByTrackerNameAndConsumerARN(ctx, conn, trackerAssociationId.TrackerName, trackerAssociationId.ConsumerARN)

			if err != nil {
				if retry.NotFound(err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
					return nil
				}
				return err
			}

			return create.Error(names.Location, create.ErrActionCheckingDestroyed, tflocation.ResNameTrackerAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTrackerAssociationExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameTrackerAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameTrackerAssociation, name, errors.New("not set"))
		}

		trackerAssociationId, err := tflocation.TrackerAssociationParseID(rs.Primary.ID)

		if err != nil {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameTrackerAssociation, name, err)
		}

		conn := acctest.ProviderMeta(ctx, t).LocationClient(ctx)

		err = tflocation.FindTrackerAssociationByTrackerNameAndConsumerARN(ctx, conn, trackerAssociationId.TrackerName, trackerAssociationId.ConsumerARN)

		if err != nil {
			return create.Error(names.Location, create.ErrActionCheckingExistence, tflocation.ResNameTrackerAssociation, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccTrackerAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q
}

resource "aws_location_tracker" "test" {
  tracker_name = %[1]q
}

resource "aws_location_tracker_association" "test" {
  consumer_arn = aws_location_geofence_collection.test.collection_arn
  tracker_name = aws_location_tracker.test.tracker_name
}
`, rName)
}
