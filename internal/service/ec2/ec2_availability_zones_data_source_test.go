package ec2_test

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAvailabilityZonesSort(t *testing.T) {
	azs := []*ec2.AvailabilityZone{
		{
			ZoneName: aws.String("name_YYY"),
			ZoneId:   aws.String("id_YYY"),
		},
		{
			ZoneName: aws.String("name_AAA"),
			ZoneId:   aws.String("id_AAA"),
		},
		{
			ZoneName: aws.String("name_ZZZ"),
			ZoneId:   aws.String("id_ZZZ"),
		},
		{
			ZoneName: aws.String("name_BBB"),
			ZoneId:   aws.String("id_BBB"),
		},
	}
	sort.Slice(azs, func(i, j int) bool {
		return aws.StringValue(azs[i].ZoneName) < aws.StringValue(azs[j].ZoneName)
	})

	cases := []struct {
		Index    int
		ZoneName string
		ZoneId   string
	}{
		{
			Index:    0,
			ZoneName: "name_AAA",
			ZoneId:   "id_AAA",
		},
		{
			Index:    1,
			ZoneName: "name_BBB",
			ZoneId:   "id_BBB",
		},
		{
			Index:    2,
			ZoneName: "name_YYY",
			ZoneId:   "id_YYY",
		},
		{
			Index:    3,
			ZoneName: "name_ZZZ",
			ZoneId:   "id_ZZZ",
		},
	}
	for _, tc := range cases {
		az := azs[tc.Index]
		if aws.StringValue(az.ZoneName) != tc.ZoneName {
			t.Fatalf("AvailabilityZones index %d got zone name %s, expected %s", tc.Index, aws.StringValue(az.ZoneName), tc.ZoneName)
		}
		if aws.StringValue(az.ZoneId) != tc.ZoneId {
			t.Fatalf("AvailabilityZones index %d got zone ID %s, expected %s", tc.Index, aws.StringValue(az.ZoneId), tc.ZoneId)
		}
	}
}

func TestAccEC2AvailabilityZonesDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSAvailabilityZonesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesMeta("data.aws_availability_zones.availability_zones"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZonesDataSource_allAvailabilityZones(t *testing.T) {
	dataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAvailabilityZonesAllAvailabilityZonesConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesMeta(dataSourceName),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZonesDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAvailabilityZonesFilterConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesMeta(dataSourceName),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZonesDataSource_excludeNames(t *testing.T) {
	allDataSourceName := "data.aws_availability_zones.all"
	excludeDataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAvailabilityZonesExcludeNamesConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesExcluded(allDataSourceName, excludeDataSourceName),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZonesDataSource_excludeZoneIDs(t *testing.T) {
	allDataSourceName := "data.aws_availability_zones.all"
	excludeDataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAvailabilityZonesExcludeZoneIDsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesExcluded(allDataSourceName, excludeDataSourceName),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZonesDataSource_stateFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSAvailabilityZonesStateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZoneState("data.aws_availability_zones.state_filter"),
				),
			},
		},
	})
}

func testAccCheckAvailabilityZonesMeta(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find AZ resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AZ resource ID not set.")
		}

		actual, err := testAccCheckAvailabilityZonesBuildAvailable(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		expected := actual
		sort.Strings(expected)
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("AZs not sorted - expected %v, got %v", expected, actual)
		}
		return nil
	}
}

func testAccCheckAvailabilityZonesExcluded(allDataSourceName, excludeDataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		allResourceState, ok := s.RootModule().Resources[allDataSourceName]
		if !ok {
			return fmt.Errorf("Resource does not exist: %s", allDataSourceName)
		}

		excludeResourceState, ok := s.RootModule().Resources[excludeDataSourceName]
		if !ok {
			return fmt.Errorf("Resource does not exist: %s", excludeDataSourceName)
		}

		for _, attribute := range []string{"names.#", "zone_ids.#"} {
			allValue, ok := allResourceState.Primary.Attributes[attribute]

			if !ok {
				return fmt.Errorf("cannot find %s in %s resource state attributes: %+v", attribute, allDataSourceName, allResourceState.Primary.Attributes)
			}

			excludeValue, ok := excludeResourceState.Primary.Attributes[attribute]

			if !ok {
				return fmt.Errorf("cannot find %s in %s resource state attributes: %+v", attribute, excludeDataSourceName, excludeResourceState.Primary.Attributes)
			}

			if allValue == excludeValue {
				return fmt.Errorf("expected %s attribute value difference, got: %s", attribute, allValue)
			}
		}

		return nil
	}
}

func testAccCheckAvailabilityZoneState(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find AZ resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AZ resource ID not set.")
		}

		if _, ok := rs.Primary.Attributes["state"]; !ok {
			return fmt.Errorf("AZs state filter is missing, should be set.")
		}

		_, err := testAccCheckAvailabilityZonesBuildAvailable(rs.Primary.Attributes)
		return err
	}
}

func testAccCheckAvailabilityZonesBuildAvailable(attrs map[string]string) ([]string, error) {
	groupNames, groupNamesOk := attrs["group_names.#"]

	if !groupNamesOk {
		return nil, fmt.Errorf("Availability Zone Group names list is missing.")
	}

	groupNamesQty, err := strconv.Atoi(groupNames)

	if err != nil {
		return nil, err
	}

	if groupNamesQty < 1 {
		return nil, fmt.Errorf("No Availability Zone Groups found in region, this is probably a bug.")
	}

	v, ok := attrs["names.#"]
	if !ok {
		return nil, fmt.Errorf("Available AZ name list is missing.")
	}
	qty, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}
	if qty < 1 {
		return nil, fmt.Errorf("No AZs found in region, this is probably a bug.")
	}
	_, ok = attrs["zone_ids.#"]
	if !ok {
		return nil, fmt.Errorf("Available AZ ID list is missing.")
	}
	zones := make([]string, qty)
	for n := range zones {
		zone, ok := attrs["names."+strconv.Itoa(n)]
		if !ok {
			return nil, fmt.Errorf("AZ list corrupt, this is definitely a bug.")
		}
		zones[n] = zone
	}
	return zones, nil
}

const testAccCheckAWSAvailabilityZonesConfig = `
data "aws_availability_zones" "availability_zones" {}
`

func testAccCheckAvailabilityZonesAllAvailabilityZonesConfig() string {
	return `
data "aws_availability_zones" "test" {
  all_availability_zones = true
}
`
}

func testAccCheckAvailabilityZonesFilterConfig() string {
	return `
data "aws_availability_zones" "test" {
  filter {
    name   = "state"
    values = ["available"]
  }
}
`
}

func testAccCheckAvailabilityZonesExcludeNamesConfig() string {
	return `
data "aws_availability_zones" "all" {}

data "aws_availability_zones" "test" {
  exclude_names = [data.aws_availability_zones.all.names[0]]
}
`
}

func testAccCheckAvailabilityZonesExcludeZoneIDsConfig() string {
	return `
data "aws_availability_zones" "all" {}

data "aws_availability_zones" "test" {
  exclude_zone_ids = [data.aws_availability_zones.all.zone_ids[0]]
}
`
}

const testAccCheckAWSAvailabilityZonesStateConfig = `
data "aws_availability_zones" "state_filter" {
  state = "available"
}
`
