package appstream_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(appstream.EndpointsID, testAccErrorCheckSkipAppStream)
	resource.AddTestSweepers("aws_appstream_fleet", &resource.Sweeper{
		Name: "aws_appstream_fleet",
		F:    testSweepAppStreamFleet,
		Dependencies: []string{
			"aws_vpc",
			"aws_subnet",
		},
	})
}

func testSweepAppStreamFleet(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).AppStreamConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &appstream.DescribeFleetsInput{}

	err = tfappstream.DescribeFleetsPagesWithContext(context.TODO(), conn, input, func(page *appstream.DescribeFleetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, fleet := range page.Fleets {
			if fleet == nil {
				continue
			}

			id := aws.StringValue(fleet.Name)

			r := ResourceImageBuilder()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppStream Fleets: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppStream Fleets for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Fleets sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

// testAccErrorCheckSkipAppStream skips AppStream tests that have error messages indicating unsupported features
func testAccErrorCheckSkipAppStream(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"ResourceNotFoundException: The image",
	)
}

func TestAccAwsAppStreamFleet_basic(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.test"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfig(rName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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

func TestAccAwsAppStreamFleet_disappears(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.test"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfig(rName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppStreamFleet_completeWithStop(t *testing.T) {
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"
	fleetType := "ON_DEMAND"
	instanceType := "stream.standard.small"
	instanceTypeUpdate := "stream.standard.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigComplete(rName, description, fleetType, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
				),
			},
			{
				Config: testAccAwsAppStreamFleetConfigComplete(rName, descriptionUpdated, fleetType, instanceTypeUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceTypeUpdate),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	fleetType := "ON_DEMAND"
	instanceType := "stream.standard.small"
	displayName := "display name of a test"
	displayNameUpdated := "display name of a test updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigCompleteWithoutStopping(rName, description, fleetType, instanceType, displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "display_name", displayName),
				),
			},
			{
				Config: testAccAwsAppStreamFleetConfigCompleteWithoutStopping(rName, description, fleetType, instanceType, displayNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	fleetType := "ON_DEMAND"
	instanceType := "stream.standard.small"
	displayName := "display name of a test"
	displayNameUpdated := "display name of a test updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigWithTags(rName, description, fleetType, instanceType, displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
				),
			},
			{
				Config: testAccAwsAppStreamFleetConfigWithTags(rName, description, fleetType, instanceType, displayNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.FleetStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

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

func testAccAwsAppStreamFleetConfig(name, instanceType string) string {
	// "Amazon-AppStream2-Sample-Image-02-04-2019" is not available in GovCloud
	return fmt.Sprintf(`
resource "aws_appstream_fleet" "test" {
  name          = %[1]q
  image_name    = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type = %[2]q

  compute_capacity {
    desired_instances = 1
  }
}
`, name, instanceType)
}

func testAccAwsAppStreamFleetConfigComplete(name, description, fleetType, instanceType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_fleet" "test" {
  name       = %[1]q
  image_name = "Amazon-AppStream2-Sample-Image-02-04-2019"

  compute_capacity {
    desired_instances = 1
  }

  description                        = %[2]q
  idle_disconnect_timeout_in_seconds = 70
  enable_default_internet_access     = false
  fleet_type                         = %[3]q
  instance_type                      = %[4]q
  max_user_duration_in_seconds       = 1000

  vpc_config {
    subnet_ids = aws_subnet.test.*.id
  }
}
`, name, description, fleetType, instanceType))
}

func testAccAwsAppStreamFleetConfigCompleteWithoutStopping(name, description, fleetType, instanceType, displayName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_fleet" "test" {
  name       = %[1]q
  image_name = "Amazon-AppStream2-Sample-Image-02-04-2019"

  compute_capacity {
    desired_instances = 1
  }

  description                        = %[2]q
  display_name                       = %[5]q
  idle_disconnect_timeout_in_seconds = 70
  enable_default_internet_access     = false
  fleet_type                         = %[3]q
  instance_type                      = %[4]q
  max_user_duration_in_seconds       = 1000

  vpc_config {
    subnet_ids = aws_subnet.test.*.id
  }
}
`, name, description, fleetType, instanceType, displayName))
}

func testAccAwsAppStreamFleetConfigWithTags(name, description, fleetType, instanceType, displayName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_fleet" "test" {
  name       = %[1]q
  image_name = "Amazon-AppStream2-Sample-Image-02-04-2019"

  compute_capacity {
    desired_instances = 1
  }

  description                        = %[2]q
  display_name                       = %[5]q
  idle_disconnect_timeout_in_seconds = 70
  enable_default_internet_access     = false
  fleet_type                         = %[3]q
  instance_type                      = %[4]q
  max_user_duration_in_seconds       = 1000

  tags = {
    Key = "value"
  }

  vpc_config {
    subnet_ids = aws_subnet.test.*.id
  }
}
`, name, description, fleetType, instanceType, displayName))
}
