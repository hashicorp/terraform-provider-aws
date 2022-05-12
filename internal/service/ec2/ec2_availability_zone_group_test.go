package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2AvailabilityZoneGroup_optInStatus(t *testing.T) {
	resourceName := "aws_ec2_availability_zone_group.test"

	// Filter to one Availability Zone Group per Region as Local Zones become available
	// e.g. ensure there are not two us-west-2-XXX when adding to this list
	// (Not including in config to avoid lintignoring entire config.)
	localZone := "us-west-2-lax-1" // lintignore:AWSAT003

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckRegion(t, endpoints.UsWest2RegionID) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2AvailabilityZoneGroupOptInStatusConfig(localZone, ec2.AvailabilityZoneOptInStatusOptedIn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// InvalidOptInStatus: Opting out of Local Zones is not currently supported. Contact AWS Support for additional assistance.
			/*
				{
					Config: testAccEc2AvailabilityZoneGroupConfigOptInStatus(ec2.AvailabilityZoneOptInStatusNotOptedIn),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusNotOptedIn),
					),
				},
				{
					Config: testAccEc2AvailabilityZoneGroupConfigOptInStatus(ec2.AvailabilityZoneOptInStatusOptedIn),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
					),
				},
			*/
		},
	})
}

func testAccEc2AvailabilityZoneGroupOptInStatusConfig(name, optInStatus string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "test" {
  all_availability_zones = true

  filter {
    name = "group-name"
    values = [
      %[1]q,
    ]
  }
}

resource "aws_ec2_availability_zone_group" "test" {
  # The above group-name filter should ensure one Availability Zone Group per Region
  group_name    = tolist(data.aws_availability_zones.test.group_names)[0]
  opt_in_status = %[2]q
}
`, name, optInStatus)
}
