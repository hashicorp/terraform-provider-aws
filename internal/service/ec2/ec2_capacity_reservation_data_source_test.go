// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2CapacityReservationDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_capacity_reservation.test"
	resourceName := "aws_ec2_capacity_reservation.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationDataSourceConfig_id,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAvailabilityZone, resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(dataSourceName, "ebs_optimized", resourceName, "ebs_optimized"),
					resource.TestCheckResourceAttrPair(dataSourceName, "end_date_type", resourceName, "end_date_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrInstanceCount, resourceName, names.AttrInstanceCount),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_match_criteria", resourceName, "instance_match_criteria"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_platform", resourceName, "instance_platform"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrState, "active"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tenancy", resourceName, "tenancy"),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservationDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_capacity_reservation.test"
	resourceName := "aws_ec2_capacity_reservation.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationDataSourceConfig_filter,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
				),
			},
		},
	})
}

const testAccCapacityReservationDataSourceConfig_id = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t3.micro"
}

data "aws_ec2_capacity_reservation" "test" {
  id = aws_ec2_capacity_reservation.test.id
}
`

const testAccCapacityReservationDataSourceConfig_filter = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_caller_identity" "current" {}

resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "m5.xlarge"
}

data "aws_ec2_capacity_reservation" "test" {
  filter {
    name   = "instance-type"
    values = [aws_ec2_capacity_reservation.test.instance_type]
  }

  filter {
    name   = "state"
    values = ["active"]
  }

  filter {
    name   = "availability-zone"
    values = [aws_ec2_capacity_reservation.test.availability_zone]
  }

  filter {
    name   = "owner-id"
    values = [data.aws_caller_identity.current.account_id]
  }
}
`
