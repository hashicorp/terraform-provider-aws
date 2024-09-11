// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2Host_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var host awstypes.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(ctx, resourceName, &host),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "on"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "off"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "c5.large"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2Host_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var host awstypes.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(ctx, resourceName, &host),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceHost(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2Host_instanceFamily(t *testing.T) {
	ctx := acctest.Context(t)
	var host awstypes.Host
	resourceName := "aws_ec2_host.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_instanceFamily(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(ctx, resourceName, &host),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "off"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "on"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", "c5"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHostConfig_instanceType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(ctx, resourceName, &host),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "on"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "off"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "c5.xlarge"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccEC2Host_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var host awstypes.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(ctx, resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHostConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(ctx, resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccHostConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(ctx, resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEC2Host_outpostAssetId(t *testing.T) {
	ctx := acctest.Context(t)
	var host awstypes.Host
	resourceName := "aws_ec2_host.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_outpostAssetId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(ctx, resourceName, &host),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
				),
			},
		},
	})
}

func TestAccEC2Host_outpost(t *testing.T) {
	ctx := acctest.Context(t)
	var host awstypes.Host
	resourceName := "aws_ec2_host.test"
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(ctx, resourceName, &host),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckHostExists(ctx context.Context, n string, v *awstypes.Host) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Host ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindHostByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckHostDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_host" {
				continue
			}

			_, err := tfec2.FindHostByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Host %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccHostConfig_basic() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[1]
  instance_type     = "c5.large"
}
`)
}

func testAccHostConfig_instanceFamily(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  auto_placement    = "off"
  availability_zone = data.aws_availability_zones.available.names[0]
  host_recovery     = "on"
  instance_family   = "c5"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccHostConfig_instanceType(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  auto_placement    = "on"
  availability_zone = data.aws_availability_zones.available.names[0]
  host_recovery     = "off"
  instance_type     = "c5.xlarge"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccHostConfig_tags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "c5.large"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccHostConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "c5.large"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccHostConfig_outpostAssetId(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_outposts_assets" "test" {
  arn = data.aws_outposts_outpost.test.arn
}

resource "aws_ec2_host" "test" {
  asset_id          = tolist(data.aws_outposts_assets.test.asset_ids)[3]
  instance_family   = "m5d"
  availability_zone = data.aws_availability_zones.available.names[0]
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccHostConfig_outpost(rName string) string {
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
`, rName))
}
