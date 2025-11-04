// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2HostDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_host.test"
	resourceName := "aws_ec2_host.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_placement", resourceName, "auto_placement"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAvailabilityZone, resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(dataSourceName, "cores"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_recovery", resourceName, "host_recovery"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_family", resourceName, "instance_family"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrPair(dataSourceName, "outpost_arn", resourceName, "outpost_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(dataSourceName, "sockets"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrSet(dataSourceName, "total_vcpus"),
				),
			},
		},
	})
}

func TestAccEC2HostDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_host.test"
	resourceName := "aws_ec2_host.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "asset_id", resourceName, "asset_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_placement", resourceName, "auto_placement"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAvailabilityZone, resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(dataSourceName, "cores"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_recovery", resourceName, "host_recovery"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_family", resourceName, "instance_family"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrPair(dataSourceName, "outpost_arn", resourceName, "outpost_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(dataSourceName, "sockets"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrSet(dataSourceName, "total_vcpus"),
				),
			},
		},
	})
}

func testAccHostDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "c5.large"

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_host" "test" {
  host_id = aws_ec2_host.test.id
}
`, rName))
}

func testAccHostDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "c5.large"

  tags = {
    %[1]q = "True"
  }
}

data "aws_ec2_host" "test" {
  filter {
    name   = "availability-zone"
    values = [aws_ec2_host.test.availability_zone]
  }

  filter {
    name   = "instance-type"
    values = [aws_ec2_host.test.instance_type]
  }

  filter {
    name   = "tag-key"
    values = [%[1]q]
  }
}
`, rName))
}
