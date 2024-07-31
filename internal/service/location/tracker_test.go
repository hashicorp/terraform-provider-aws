// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLocationTracker_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "position_filtering", locationservice.PositionFilteringTimeBased),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, "tracker_arn", "geo", fmt.Sprintf("tracker/%s", rName)),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflocation.ResourceTracker(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLocationTracker_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, resourceName),
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
					testAccCheckTrackerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccLocationTracker_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, resourceName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_positionFiltering(rName, locationservice.PositionFilteringAccuracyBased),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "position_filtering", locationservice.PositionFilteringAccuracyBased),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTrackerConfig_positionFiltering(rName, locationservice.PositionFilteringDistanceBased),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "position_filtering", locationservice.PositionFilteringDistanceBased),
				),
			},
		},
	})
}

func TestAccLocationTracker_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckTrackerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTrackerConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckTrackerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_location_tracker" {
				continue
			}

			input := &locationservice.DescribeTrackerInput{
				TrackerName: aws.String(rs.Primary.ID),
			}

			output, err := conn.DescribeTrackerWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
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

func testAccCheckTrackerExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn(ctx)

		input := &locationservice.DescribeTrackerInput{
			TrackerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeTrackerWithContext(ctx, input)

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
