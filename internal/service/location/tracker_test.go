package location_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
)

func TestAccLocationTracker_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "create_time"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "position_filtering", locationservice.PositionFilteringTimeBased),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tflocation.ResourceTracker(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLocationTracker_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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
					testAccCheckTrackerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccLocationTracker_kmsKeyID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.test", "arn"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_positionFiltering(rName, locationservice.PositionFilteringAccuracyBased),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(resourceName),
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
					testAccCheckTrackerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "position_filtering", locationservice.PositionFilteringDistanceBased),
				),
			},
		},
	})
}

func TestAccLocationTracker_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTrackerConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTrackerConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}
func testAccCheckTrackerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_location_tracker" {
			continue
		}

		input := &locationservice.DescribeTrackerInput{
			TrackerName: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeTracker(input)

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

func testAccCheckTrackerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn

		input := &locationservice.DescribeTrackerInput{
			TrackerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeTracker(input)

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
