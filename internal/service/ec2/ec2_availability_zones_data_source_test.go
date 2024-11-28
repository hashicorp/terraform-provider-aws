// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2AvailabilityZonesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZonesDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesMeta("data.aws_availability_zones.availability_zones"),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZonesDataSource_allAvailabilityZones(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZonesDataSourceConfig_all(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesMeta(dataSourceName),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZonesDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZonesDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesMeta(dataSourceName),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZonesDataSource_excludeNames(t *testing.T) {
	ctx := acctest.Context(t)
	allDataSourceName := "data.aws_availability_zones.all"
	excludeDataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZonesDataSourceConfig_excludeNames(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesExcluded(allDataSourceName, excludeDataSourceName),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZonesDataSource_excludeZoneIDs(t *testing.T) {
	ctx := acctest.Context(t)
	allDataSourceName := "data.aws_availability_zones.all"
	excludeDataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZonesDataSourceConfig_excludeZoneIDs(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesExcluded(allDataSourceName, excludeDataSourceName),
				),
			},
		},
	})
}

func TestAccEC2AvailabilityZonesDataSource_stateFilter(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailabilityZonesDataSourceConfig_state,
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
		slices.Sort(expected)
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

		if _, ok := rs.Primary.Attributes[names.AttrState]; !ok {
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

const testAccAvailabilityZonesDataSourceConfig_basic = `
data "aws_availability_zones" "availability_zones" {}
`

func testAccAvailabilityZonesDataSourceConfig_all() string {
	return `
data "aws_availability_zones" "test" {
  all_availability_zones = true
}
`
}

func testAccAvailabilityZonesDataSourceConfig_filter() string {
	return `
data "aws_availability_zones" "test" {
  filter {
    name   = "state"
    values = ["available"]
  }
}
`
}

func testAccAvailabilityZonesDataSourceConfig_excludeNames() string {
	return `
data "aws_availability_zones" "all" {}

data "aws_availability_zones" "test" {
  exclude_names = [data.aws_availability_zones.all.names[0]]
}
`
}

func testAccAvailabilityZonesDataSourceConfig_excludeZoneIDs() string {
	return `
data "aws_availability_zones" "all" {}

data "aws_availability_zones" "test" {
  exclude_zone_ids = [data.aws_availability_zones.all.zone_ids[0]]
}
`
}

const testAccAvailabilityZonesDataSourceConfig_state = `
data "aws_availability_zones" "state_filter" {
  state = "available"
}
`
