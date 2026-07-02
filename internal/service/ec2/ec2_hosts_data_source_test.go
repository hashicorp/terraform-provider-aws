// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2HostsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_hosts.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", 0),
				),
			},
		},
	})
}

func TestAccEC2HostsDataSource_outpostARN(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_hosts.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostsDataSourceConfig_outpostARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
				),
			},
		},
	})
}

func TestAccEC2HostsDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_hosts.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostsDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
				),
			},
		},
	})
}

func testAccHostsDataSourceConfig_outpostARN(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_ec2_host" "test" {
  instance_family   = "r5d"
  availability_zone = data.aws_availability_zones.available.names[1]
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_hosts" "test" {
  outpost_arn = data.aws_outposts_outpost.test.arn

  filter {
    name   = "instance-type"
    values = ["r5d.*"]
  }

  depends_on = [aws_ec2_host.test]
}
`, rName))
}

func testAccHostsDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "c5.large"

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_hosts" "test" {
  filter {
    name   = "instance-type"
    values = ["c5.large"]
  }

  filter {
    name   = "availability-zone"
    values = [aws_ec2_host.test.availability_zone]
  }
}
`, rName))
}

func testAccHostsDataSourceConfig_tags(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "c5.large"

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_hosts" "test" {
  tags = {
    Name = aws_ec2_host.test.tags["Name"]
  }
}
`, rName))
}
