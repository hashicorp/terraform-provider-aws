// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceTypeOfferingDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckInstanceTypeOfferings(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrInstanceType),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeOfferingDataSource_locationType(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckInstanceTypeOfferings(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingDataSourceConfig_location(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrInstanceType),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeOfferingDataSource_preferredInstanceTypes(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckInstanceTypeOfferings(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingDataSourceConfig_preferreds(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrInstanceType, "t3.micro"),
				),
			},
		},
	})
}

func testAccInstanceTypeOfferingDataSourceConfig_filter() string {
	return `
# Rather than hardcode an instance type in the testing,
# use the first result from all available offerings.
data "aws_ec2_instance_type_offerings" "test" {}

data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
    values = [tolist(data.aws_ec2_instance_type_offerings.test.instance_types)[0]]
  }
}
`
}

func testAccInstanceTypeOfferingDataSourceConfig_location() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
# Rather than hardcode an instance type in the testing,
# use the first result from all available offerings.
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type = "availability-zone"
}

data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
    values = [tolist(data.aws_ec2_instance_type_offerings.test.instance_types)[0]]
  }

  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type = "availability-zone"
}
`)
}

func testAccInstanceTypeOfferingDataSourceConfig_preferreds() string {
	return `
data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
    values = ["t1.micro", "t2.micro", "t3.micro"]
  }

  preferred_instance_types = ["t3.micro", "t2.micro", "t1.micro"]
}
`
}
