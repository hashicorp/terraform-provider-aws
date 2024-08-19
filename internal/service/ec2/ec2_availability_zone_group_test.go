// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2AvailabilityZoneGroup_optInStatus(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_availability_zone_group.test"

	// Filter to one Availability Zone Group per Region as Local Zones become available
	// e.g. ensure there are not two us-west-2-XXX when adding to this list
	// (Not including in config to avoid lintignoring entire config.)
	localZone := "us-west-2-lax-1" // lintignore:AWSAT003

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USWest2RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneGroupConfig_optInStatus(localZone, string(awstypes.AvailabilityZoneOptInStatusOptedIn)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "opt_in_status", string(awstypes.AvailabilityZoneOptInStatusOptedIn)),
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
					Config: testAccAvailabilityZoneGroupConfig_optInStatus(ec2.AvailabilityZoneOptInStatusNotOptedIn),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusNotOptedIn),
					),
				},
				{
					Config: testAccAvailabilityZoneGroupConfig_optInStatus(ec2.AvailabilityZoneOptInStatusOptedIn),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
					),
				},
			*/
		},
	})
}

func testAccAvailabilityZoneGroupConfig_optInStatus(name, optInStatus string) string {
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
