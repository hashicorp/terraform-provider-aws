// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2AvailabilityZoneDataSource_allAvailabilityZones(t *testing.T) {
	ctx := acctest.Context(t)
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneDataSourceConfig_allAZs(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrGroupName, acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexache.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", string(awstypes.AvailabilityZoneOptInStatusOptInNotRequired)),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZoneDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrGroupName, acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexache.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", string(awstypes.AvailabilityZoneOptInStatusOptInNotRequired)),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZoneDataSource_localZone(t *testing.T) {
	ctx := acctest.Context(t)
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckLocalZoneAvailable(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneDataSourceConfig_type("local-zone"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrGroupName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexache.MustCompile(`^[0-9a-z][0-9a-z-]+$`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "network_border_group"),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", string(awstypes.AvailabilityZoneOptInStatusOptedIn)),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_name"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "local-zone"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZoneDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneDataSourceConfig_name(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrGroupName, acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexache.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", string(awstypes.AvailabilityZoneOptInStatusOptInNotRequired)),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZoneDataSource_wavelengthZone(t *testing.T) {
	ctx := acctest.Context(t)
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWavelengthZoneAvailable(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneDataSourceConfig_type("wavelength-zone"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrGroupName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexache.MustCompile(`^[0-9a-z][0-9a-z-]+$`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "network_border_group"),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", string(awstypes.AvailabilityZoneOptInStatusOptedIn)),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_name"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "wavelength-zone"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZoneDataSource_zoneID(t *testing.T) {
	ctx := acctest.Context(t)
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrGroupName, acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexache.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", string(awstypes.AvailabilityZoneOptInStatusOptInNotRequired)),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func testAccPreCheckLocalZoneAvailable(ctx context.Context, t *testing.T, groupNames ...string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeAvailabilityZonesInput{
		Filters: tfec2.NewAttributeFilterList(map[string]string{
			"zone-type":     "local-zone",
			"opt-in-status": "opted-in",
		}),
	}

	if len(groupNames) > 0 {
		input.Filters = append(input.Filters, awstypes.Filter{
			Name:   aws.String("group-name"),
			Values: groupNames,
		})
	}

	output, err := tfec2.FindAvailabilityZones(ctx, conn, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if len(output) == 0 {
		t.Skip("skipping since no Local Zones are available")
	}
}

func testAccAvailabilityZoneDataSourceConfig_allAZs() string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		`
data "aws_availability_zone" "test" {
  all_availability_zones = true
  name                   = data.aws_availability_zones.available.names[0]
}
`)
}

func testAccAvailabilityZoneDataSourceConfig_filter() string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		`
data "aws_availability_zone" "test" {
  filter {
    name   = "zone-name"
    values = [data.aws_availability_zones.available.names[0]]
  }
}
`)
}

func testAccAvailabilityZoneDataSourceConfig_name() string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		`
data "aws_availability_zone" "test" {
  name = data.aws_availability_zones.available.names[0]
}
`)
}

func testAccAvailabilityZoneDataSourceConfig_id() string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		`
data "aws_availability_zone" "test" {
  zone_id = data.aws_availability_zones.available.zone_ids[0]
}
`)
}

func testAccAvailabilityZoneDataSourceConfig_type(zoneType string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "zone-type"
    values = [%[1]q]
  }

  filter {
    name   = "opt-in-status"
    values = ["opted-in"]
  }
}

data "aws_availability_zone" "test" {
  zone_id = data.aws_availability_zones.available.zone_ids[0]
}
`, zoneType)
}
