package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2AMICopy_basic(t *testing.T) {
	var image ec2.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMICopyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &image),
					testAccCheckAMICopyAttributes(&image, rName),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "usage_operation", "RunInstances"),
					resource.TestCheckResourceAttr(resourceName, "platform_details", "Linux/UNIX"),
					resource.TestCheckResourceAttr(resourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(resourceName, "hypervisor", "xen"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccEC2AMICopy_description(t *testing.T) {
	var image ec2.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMICopyDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAMICopyDescriptionConfig(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccEC2AMICopy_enaSupport(t *testing.T) {
	var image ec2.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMICopyENASupportConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
				),
			},
		},
	})
}

func TestAccEC2AMICopy_destinationOutpost(t *testing.T) {
	var image ec2.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMICopyDestOutpostConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &image),
					resource.TestCheckResourceAttrPair(resourceName, "destination_outpost_arn", outpostDataSourceName, "arn"),
				),
			},
		},
	})
}

func TestAccEC2AMICopy_tags(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMICopyTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					testAccCheckAMICopyAttributes(&ami, rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAMICopyTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					testAccCheckAMICopyAttributes(&ami, rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAMICopyTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					testAccCheckAMICopyAttributes(&ami, rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAMICopyAttributes(image *ec2.Image, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if expected := ec2.ImageStateAvailable; aws.StringValue(image.State) != expected {
			return fmt.Errorf("invalid image state; expected %s, got %s", expected, aws.StringValue(image.State))
		}
		if expected := ec2.ImageTypeValuesMachine; aws.StringValue(image.ImageType) != expected {
			return fmt.Errorf("wrong image type; expected %s, got %s", expected, aws.StringValue(image.ImageType))
		}
		if expected := expectedName; aws.StringValue(image.Name) != expected {
			return fmt.Errorf("wrong name; expected %s, got %s", expected, aws.StringValue(image.Name))
		}

		snapshots := []string{}
		for _, bdm := range image.BlockDeviceMappings {
			// The snapshot ID might not be set,
			// even for a block device that is an
			// EBS volume.
			if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
				snapshots = append(snapshots, aws.StringValue(bdm.Ebs.SnapshotId))
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

func testAccAMICopyTags1Config(rName, tagKey1, tagValue1 string) string {
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

func testAccAMICopyTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccAMICopyConfig(rName string) string {
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

func testAccAMICopyDescriptionConfig(rName, description string) string {
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

func testAccAMICopyENASupportConfig(rName string) string {
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

func testAccAMICopyDestOutpostConfig(rName string) string {
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
