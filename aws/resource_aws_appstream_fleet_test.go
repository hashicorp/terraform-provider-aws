package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsAppStreamFleet_basic(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.fleet"
	fleetName := acctest.RandomWithPrefix("tf-acc-test")
	instanceType := "stream.standard.small"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigBasic(fleetName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
				),
			},
			{
				Config:            testAccAwsAppStreamFleetConfigBasic(fleetName, instanceType),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppStreamFleet_disappears(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.fleet"
	fleetName := acctest.RandomWithPrefix("tf-acc-test")
	instanceType := "stream.standard.small"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigBasic(fleetName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppstreamFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsAppStreamFleet_withTags(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.fleet"
	fleetName := acctest.RandomWithPrefix("tf-acc-test")
	description := "Description of a fleet"
	fleetType := "ON_DEMAND"
	instanceType := "stream.standard.small"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigWithTags(fleetName, description, fleetType, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
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

func testAccCheckAwsAppStreamFleetExists(resourceName string, appStreamFleet *appstream.Fleet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appstreamconn
		resp, err := conn.DescribeFleets(&appstream.DescribeFleetsInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return err
		}

		if resp == nil && len(resp.Fleets) == 0 {
			return fmt.Errorf("appstream fleet %q does not exist", rs.Primary.ID)
		}

		*appStreamFleet = *resp.Fleets[0]

		return nil
	}
}

func testAccCheckAwsAppStreamFleetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appstreamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_fleet" {
			continue
		}

		resp, err := conn.DescribeFleets(&appstream.DescribeFleetsInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return err
		}

		if resp != nil && len(resp.Fleets) > 0 {
			return fmt.Errorf("appstream fleet %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsAppStreamFleetConfigBasic(fleetName, instaceType string) string {
	return fmt.Sprintf(`
resource "aws_appstream_fleet" "test_fleet" {
  name       = %[1]q
  compute_capacity {
    desired_instances = 1
  }
  instance_type                       = %[2]q
}
`, fleetName, instaceType)
}

func testAccAwsAppStreamFleetConfigWithTags(fleetName, description, fleetType, instaceType string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state            = "available"
}

resource "aws_vpc" "example" {
  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"
  vpc_id            = aws_vpc.example.id
}

resource "aws_appstream_fleet" "test_fleet" {
  name       = %[1]q
  compute_capacity {
    desired_instances = 1
  }
  description                         = %[2]q
  idle_disconnect_timeout_in_seconds  = 70
  display_name                        = %[1]q
  enable_default_internet_access      = false
  fleet_type                          = %[3]q
  instance_type                       = %[4]q
  max_user_duration_in_seconds        = 600
  vpc_config {
    subnet_ids                          = [aws_subnet.example.id]
  }
  tags = {
    Key = "value"
  }
}
`, fleetName, description, fleetType, instaceType)
}
