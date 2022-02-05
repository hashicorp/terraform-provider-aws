package location_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccLocationTracker_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); PreCheckLocationService(t) },
		ErrorCheck:   acctest.ErrorCheck(t, locationservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationTrackerConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationTrackerExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "geo", regexp.MustCompile(`tracker/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "id", rName),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", "RequestBasedUsage"),
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

func TestAccLocationTracker_disappers(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_location_tracker.test"

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Instance
		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn
		_, err := conn.DeleteTracker(&locationservice.DeleteTrackerInput{
			TrackerName: aws.String(rName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Location tracker Instance in disappear test: %s", err)
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); PreCheckLocationService(t) },
		ErrorCheck:   acctest.ErrorCheck(t, locationservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationTrackerConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationTrackerExists(resourceName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLocationTracker_Description(t *testing.T) {
	var v1, v2 locationservice.DescribeTrackerOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); PreCheckLocationService(t) },
		ErrorCheck:   acctest.ErrorCheck(t, locationservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationTrackerConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationTrackerExistsWithDescribeTrackerOutput(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccLocationTrackerConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationTrackerExistsWithDescribeTrackerOutput(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					testAccCheckAwsLocationTrackerRecreated("description", &v1, &v2),
				),
			},
		},
	})
}

func TestAccLocationTracker_Tags(t *testing.T) {
	var v1, v2, v3 locationservice.DescribeTrackerOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); PreCheckLocationService(t) },
		ErrorCheck:   acctest.ErrorCheck(t, locationservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationTrackerConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationTrackerExistsWithDescribeTrackerOutput(resourceName, &v1),
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
				Config: testAccLocationTrackerConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationTrackerExistsWithDescribeTrackerOutput(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationTrackerConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationTrackerExistsWithDescribeTrackerOutput(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckLocationTrackerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		_, err := conn.DescribeTracker(&locationservice.DescribeTrackerInput{
			TrackerName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckLocationTrackerExistsWithDescribeTrackerOutput(name string, res *locationservice.DescribeTrackerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		output, err := conn.DescribeTracker(&locationservice.DescribeTrackerInput{
			TrackerName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *output

		return nil
	}
}

func testAccCheckAwsLocationTrackerRecreated(attributeName string, v1, v2 *locationservice.DescribeTrackerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v2.CreateTime.Equal(*v1.CreateTime) {
			return fmt.Errorf("Location tracker not recreated when changing %s", attributeName)
		}

		return nil
	}
}

func testAccCheckLocationTrackerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acm_certificate" {
			continue
		}
		_, err := conn.DescribeTracker(&locationservice.DescribeTrackerInput{
			TrackerName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Tracker still exists.")
		}

		// Verify the error is what we want
		if !tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException, "") {
			return err
		}
	}

	return nil
}

func PreCheckLocationService(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn
	input := &locationservice.ListTrackersInput{}
	_, err := conn.ListTrackers(input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccLocationTrackerConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  name         = %[1]q
  pricing_plan = "RequestBasedUsage"
}
`, rName)
}

func testAccLocationTrackerConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  name         = %[1]q
  description  = %[2]q
  pricing_plan = "RequestBasedUsage"
}
`, rName, description)
}

func testAccLocationTrackerConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  name         = %[1]q
  pricing_plan = "RequestBasedUsage"
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccLocationTrackerConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  name         = %[1]q
  pricing_plan = "RequestBasedUsage"
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
