package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func TestAccAwsAppStreamFleet_basic(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.test"
	instanceType := "stream.standard.small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigNameGenerated(instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				Config:            testAccAwsAppStreamFleetConfigNameGenerated(instanceType),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsAppStreamFleet_nameGenerated(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.test"
	instanceType := "stream.standard.small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigNameGenerated(instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
			{
				Config:            testAccAwsAppStreamFleetConfigNameGenerated(instanceType),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsAppStreamFleet_namePrefix(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.test"
	instanceType := "stream.standard.small"
	namePrefix := "tf-acc-test-prefix-"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigNamePrefix(instanceType, namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", namePrefix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", namePrefix),
				),
			},
			{
				Config:            testAccAwsAppStreamFleetConfigNameGenerated(instanceType),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsAppStreamFleet_disappears(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.test"
	instanceType := "stream.standard.small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigNameGenerated(instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppStreamFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppStreamFleet_completeWithStop(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.test"
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"
	fleetType := "ON_DEMAND"
	instanceType := "stream.standard.small"
	instanceTypeUpdate := "stream.standard.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigComplete(description, fleetType, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				Config: testAccAwsAppStreamFleetConfigComplete(descriptionUpdated, fleetType, instanceTypeUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceTypeUpdate),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
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

func TestAccAwsAppStreamFleet_completeWithoutStop(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.test"
	description := "Description of a test"
	fleetType := "ON_DEMAND"
	instanceType := "stream.standard.small"
	displayName := "display name of a test"
	displayNameUpdated := "display name of a test updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigCompleteWithoutStopping(description, fleetType, instanceType, displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "display_name", displayName),
				),
			},
			{
				Config: testAccAwsAppStreamFleetConfigCompleteWithoutStopping(description, fleetType, instanceType, displayNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "display_name", displayNameUpdated),
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

func TestAccAwsAppStreamFleet_withTags(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.test"
	description := "Description of a test"
	fleetType := "ON_DEMAND"
	instanceType := "stream.standard.small"
	displayName := "display name of a test"
	displayNameUpdated := "display name of a test updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigWithTags(description, fleetType, instanceType, displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				Config: testAccAwsAppStreamFleetConfigWithTags(description, fleetType, instanceType, displayNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
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

		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && len(resp.Fleets) > 0 {
			return fmt.Errorf("appstream fleet %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsAppStreamFleetConfigNameGenerated(instanceType string) string {
	return fmt.Sprintf(`
resource "aws_appstream_fleet" "test" {
  image_name = "Amazon-AppStream2-Sample-Image-02-04-2019"
  compute_capacity {
    desired_instances = 1
  }
  instance_type = %[1]q
}
`, instanceType)
}

func testAccAwsAppStreamFleetConfigNamePrefix(instanceType, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_appstream_fleet" "test" {
  image_name = "Amazon-AppStream2-Sample-Image-02-04-2019"
  compute_capacity {
    desired_instances = 1
  }
  instance_type = %[1]q
  name_prefix   = %[2]q
}
`, instanceType, namePrefix)
}

func testAccAwsAppStreamFleetConfigComplete(description, fleetType, instanceType string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "example" {
  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"
  vpc_id            = aws_vpc.example.id
}

resource "aws_appstream_fleet" "test" {
  image_name = "Amazon-AppStream2-Sample-Image-02-04-2019"
  compute_capacity {
    desired_instances = 1
  }
  description                        = %[1]q
  idle_disconnect_timeout_in_seconds = 70
  enable_default_internet_access     = false
  fleet_type                         = %[2]q
  instance_type                      = %[3]q
  max_user_duration_in_seconds       = 1000
  vpc_config {
    subnet_ids = [aws_subnet.example.id]
  }
}
`, description, fleetType, instanceType)
}

func testAccAwsAppStreamFleetConfigCompleteWithoutStopping(description, fleetType, instanceType, displayName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "example" {
  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"
  vpc_id            = aws_vpc.example.id
}

resource "aws_appstream_fleet" "test" {
  image_name = "Amazon-AppStream2-Sample-Image-02-04-2019"
  compute_capacity {
    desired_instances = 1
  }
  description                        = %[1]q
  display_name                       = %[4]q
  idle_disconnect_timeout_in_seconds = 70
  enable_default_internet_access     = false
  fleet_type                         = %[2]q
  instance_type                      = %[3]q
  max_user_duration_in_seconds       = 1000
  vpc_config {
    subnet_ids = [aws_subnet.example.id]
  }
}
`, description, fleetType, instanceType, displayName)
}

func testAccAwsAppStreamFleetConfigWithTags(description, fleetType, instanceType, displayName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "example" {
  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"
  vpc_id            = aws_vpc.example.id
}

resource "aws_appstream_fleet" "test" {
  image_name = "Amazon-AppStream2-Sample-Image-02-04-2019"
  compute_capacity {
    desired_instances = 1
  }
  description                        = %[1]q
  display_name                       = %[4]q
  idle_disconnect_timeout_in_seconds = 70
  enable_default_internet_access     = false
  fleet_type                         = %[2]q
  instance_type                      = %[3]q
  max_user_duration_in_seconds       = 1000
  vpc_config {
    subnet_ids = [aws_subnet.example.id]
  }
  tags = {
    Key = "value"
  }
}
`, description, fleetType, instanceType, displayName)
}
