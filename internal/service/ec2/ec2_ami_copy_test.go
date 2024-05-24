// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2AMICopy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var image awstypes.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMICopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &image),
					testAccCheckAMICopyAttributes(&image, rName),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "usage_operation", "RunInstances"),
					resource.TestCheckResourceAttr(resourceName, "platform_details", "Linux/UNIX"),
					resource.TestCheckResourceAttr(resourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(resourceName, "hypervisor", "xen"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
				),
			},
		},
	})
}

func TestAccEC2AMICopy_description(t *testing.T) {
	ctx := acctest.Context(t)
	var image awstypes.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMICopyConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccAMICopyConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccEC2AMICopy_enaSupport(t *testing.T) {
	ctx := acctest.Context(t)
	var image awstypes.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMICopyConfig_enaSupport(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "ena_support", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccEC2AMICopy_destinationOutpost(t *testing.T) {
	ctx := acctest.Context(t)
	var image awstypes.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMICopyConfig_destOutpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttrPair(resourceName, "destination_outpost_arn", outpostDataSourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccEC2AMICopy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ami awstypes.Image
	resourceName := "aws_ami_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMICopyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &ami),
					testAccCheckAMICopyAttributes(&ami, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccAMICopyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &ami),
					testAccCheckAMICopyAttributes(&ami, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAMICopyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &ami),
					testAccCheckAMICopyAttributes(&ami, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckAMICopyAttributes(image *awstypes.Image, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if expected := awstypes.ImageStateAvailable; image.State != expected {
			return fmt.Errorf("invalid image state; expected %s, got %s", expected, string(image.State))
		}
		if expected := awstypes.ImageTypeValuesMachine; image.ImageType != expected {
			return fmt.Errorf("wrong image type; expected %s, got %s", expected, string(image.ImageType))
		}
		if expected := expectedName; aws.ToString(image.Name) != expected {
			return fmt.Errorf("wrong name; expected %s, got %s", expected, aws.ToString(image.Name))
		}

		snapshots := []string{}
		for _, bdm := range image.BlockDeviceMappings {
			// The snapshot ID might not be set,
			// even for a block device that is an
			// EBS volume.
			if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
				snapshots = append(snapshots, aws.ToString(bdm.Ebs.SnapshotId))
			}
		}

		if expected := 1; len(snapshots) != expected {
			return fmt.Errorf("wrong number of snapshots; expected %v, got %v", expected, len(snapshots))
		}

		return nil
	}
}

func testAccAMICopyBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAMICopyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAMICopyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ami" "test" {
  name                = %[1]q
  virtualization_type = "hvm"
  root_device_name    = "/dev/sda1"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}

resource "aws_ami_copy" "test" {
  name              = %[1]q
  source_ami_id     = aws_ami.test.id
  source_ami_region = data.aws_region.current.name

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAMICopyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAMICopyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ami" "test" {
  name                = %[1]q
  virtualization_type = "hvm"
  root_device_name    = "/dev/sda1"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}

resource "aws_ami_copy" "test" {
  name              = %[1]q
  source_ami_id     = aws_ami.test.id
  source_ami_region = data.aws_region.current.name

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAMICopyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAMICopyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ami" "test" {
  name                = "%s-source"
  virtualization_type = "hvm"
  root_device_name    = "/dev/sda1"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}

resource "aws_ami_copy" "test" {
  name              = %q
  source_ami_id     = aws_ami.test.id
  source_ami_region = data.aws_region.current.name
}
`, rName, rName))
}

func testAccAMICopyConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccAMICopyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ami" "test" {
  name                = "%s-source"
  virtualization_type = "hvm"
  root_device_name    = "/dev/sda1"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}

resource "aws_ami_copy" "test" {
  description       = %q
  name              = %q
  source_ami_id     = aws_ami.test.id
  source_ami_region = data.aws_region.current.name
}
`, rName, description, rName))
}

func testAccAMICopyConfig_enaSupport(rName string) string {
	return acctest.ConfigCompose(testAccAMICopyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = "%s-source"
  virtualization_type = "hvm"
  root_device_name    = "/dev/sda1"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}

resource "aws_ami_copy" "test" {
  name              = "%s-copy"
  source_ami_id     = aws_ami.test.id
  source_ami_region = data.aws_region.current.name
}
`, rName, rName))
}

func testAccAMICopyConfig_destOutpost(rName string) string {
	return acctest.ConfigCompose(testAccAMICopyBaseConfig(rName), fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_ami" "test" {
  ena_support         = true
  name                = "%s-source"
  virtualization_type = "hvm"
  root_device_name    = "/dev/sda1"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}

resource "aws_ami_copy" "test" {
  name                    = "%s-copy"
  source_ami_id           = aws_ami.test.id
  source_ami_region       = data.aws_region.current.name
  destination_outpost_arn = data.aws_outposts_outpost.test.arn
}
`, rName, rName))
}
