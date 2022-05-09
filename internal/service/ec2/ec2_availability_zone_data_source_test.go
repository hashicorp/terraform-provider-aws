package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2AvailabilityZoneDataSource_allAvailabilityZones(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneAllAvailabilityZonesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_name", acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptInNotRequired),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "region", acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZoneDataSource_filter(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneFilterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_name", acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptInNotRequired),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "region", acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZoneDataSource_localZone(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckLocalZoneAvailable(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneZoneTypeDataSourceConfig("local-zone"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z0-9][a-z0-9-]+$`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "network_border_group"),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_name"),
					resource.TestCheckResourceAttr(dataSourceName, "region", acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "local-zone"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZoneDataSource_name(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneNameDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_name", acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptInNotRequired),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "region", acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZoneDataSource_wavelengthZone(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWavelengthZoneAvailable(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneZoneTypeDataSourceConfig("wavelength-zone"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z0-9][a-z0-9-]+$`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "network_border_group"),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_name"),
					resource.TestCheckResourceAttr(dataSourceName, "region", acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "wavelength-zone"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZoneDataSource_zoneID(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZoneZoneIDDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_name", acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptInNotRequired),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "region", acctest.Region()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func testAccPreCheckLocalZoneAvailable(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeAvailabilityZonesInput{
		Filters: tfec2.BuildAttributeFilterList(map[string]string{
			"zone-type":     "local-zone",
			"opt-in-status": "opted-in",
		}),
	}

	output, err := tfec2.FindAvailabilityZones(conn, input)

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

func testAccAvailabilityZoneAllAvailabilityZonesDataSourceConfig() string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		`
data "aws_availability_zone" "test" {
  all_availability_zones = true
  name                   = data.aws_availability_zones.available.names[0]
}
`)
}

func testAccAvailabilityZoneFilterDataSourceConfig() string {
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

func testAccAvailabilityZoneNameDataSourceConfig() string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		`
data "aws_availability_zone" "test" {
  name = data.aws_availability_zones.available.names[0]
}
`)
}

func testAccAvailabilityZoneZoneIDDataSourceConfig() string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		`
data "aws_availability_zone" "test" {
  zone_id = data.aws_availability_zones.available.zone_ids[0]
}
`)
}

func testAccAvailabilityZoneZoneTypeDataSourceConfig(zoneType string) string {
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
