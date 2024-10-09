// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSVolumeDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_volume.test"
	dataSourceName := "data.aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeIDDataSource(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrSize, resourceName, names.AttrSize),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
					resource.TestCheckResourceAttrPair(dataSourceName, "outpost_arn", resourceName, "outpost_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_attach_enabled", resourceName, "multi_attach_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrThroughput, resourceName, names.AttrThroughput),
				),
			},
		},
	})
}

func TestAccEC2EBSVolumeDataSource_multipleFilters(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_volume.test"
	dataSourceName := "data.aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeDataSourceConfig_multipleFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeIDDataSource(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrSize, resourceName, names.AttrSize),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVolumeType, "gp2"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
				),
			},
		},
	})
}

func testAccCheckEBSVolumeIDDataSource(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Volume data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Volume data source ID not set")
		}
		return nil
	}
}

func testAccEBSVolumeDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 40

  tags = {
    Name = %[1]q
  }
}

data "aws_ebs_volume" "test" {
  most_recent = true

  filter {
    name   = "tag:Name"
    values = [%[1]q]
  }

  filter {
    name   = "volume-type"
    values = [aws_ebs_volume.test.type]
  }
}
`, rName))
}

func testAccEBSVolumeDataSourceConfig_multipleFilters(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 10

  tags = {
    Name = %[1]q
  }
}

data "aws_ebs_volume" "test" {
  most_recent = true

  filter {
    name   = "tag:Name"
    values = [%[1]q]
  }

  filter {
    name   = "size"
    values = [aws_ebs_volume.test.size]
  }

  filter {
    name   = "volume-type"
    values = [aws_ebs_volume.test.type]
  }
}
`, rName))
}
