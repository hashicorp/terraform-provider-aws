// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLocationTracker_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "position_filtering", "TimeBased"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "tracker_arn", "geo", fmt.Sprintf("tracker/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tracker_name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "update_time"),
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

func TestAccLocationTracker_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflocation.ResourceTracker(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLocationTracker_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTrackerConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccLocationTracker_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrARN),
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

func TestAccLocationTracker_positionFiltering(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_positionFiltering(rName, "AccuracyBased"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "position_filtering", "AccuracyBased"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTrackerConfig_positionFiltering(rName, "DistanceBased"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "position_filtering", "DistanceBased"),
				),
			},
		},
	})
}

func TestAccLocationTracker_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTrackerConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTrackerConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckTrackerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LocationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_location_tracker" {
				continue
			}

			input := &location.DescribeTrackerInput{
				TrackerName: aws.String(rs.Primary.ID),
			}

			output, err := conn.DescribeTracker(ctx, input)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Location Service Tracker (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Location Service Tracker (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckTrackerExists(ctx context.Context, t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).LocationClient(ctx)

		input := &location.DescribeTrackerInput{
			TrackerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeTracker(ctx, input)

		if err != nil {
			return fmt.Errorf("error getting Location Service Tracker (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccTrackerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  tracker_name = %[1]q
}
`, rName)
}

func testAccTrackerConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  tracker_name = %[1]q
  description  = %[2]q
}
`, rName, description)
}

func testAccTrackerConfig_kmsKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_location_tracker" "test" {
  tracker_name = %[1]q
  kms_key_id   = aws_kms_key.test.arn
}
`, rName)
}

func testAccTrackerConfig_positionFiltering(rName, positionFiltering string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  tracker_name       = %[1]q
  position_filtering = %[2]q
}
`, rName, positionFiltering)
}

func testAccTrackerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  tracker_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccTrackerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  tracker_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
