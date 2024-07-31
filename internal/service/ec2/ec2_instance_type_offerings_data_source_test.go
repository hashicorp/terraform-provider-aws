// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceTypeOfferingsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckInstanceTypeOfferings(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingsDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "instance_types.#", 0),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "locations.#", 0),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "location_types.#", 0),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeOfferingsDataSource_locationType(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckInstanceTypeOfferings(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingsDataSourceConfig_location(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "instance_types.#", 0),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "locations.#", 0),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "location_types.#", 0),
				),
			},
		},
	})
}

func testAccPreCheckInstanceTypeOfferings(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeInstanceTypeOfferingsInput{
		MaxResults: aws.Int32(5),
	}

	_, err := conn.DescribeInstanceTypeOfferings(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccInstanceTypeOfferingsDataSourceConfig_filter() string {
	return `
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "instance-type"
    values = ["t2.micro", "t3.micro"]
  }
}
`
}

func testAccInstanceTypeOfferingsDataSourceConfig_location() string {
	return acctest.ConfigAvailableAZsNoOptIn() + `
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type = "availability-zone"
}
`
}
