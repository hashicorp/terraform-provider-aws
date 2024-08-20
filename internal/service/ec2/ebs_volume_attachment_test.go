// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func TestAccEC2EBSVolumeAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeviceName, "/dev/sdh"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVolumeAttachmentImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2EBSVolumeAttachment_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeviceName, "/dev/sdh"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVolumeAttachmentImportStateIDFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrSkipDestroy, // attribute only used on resource deletion
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeAttachment_attachStopped(t *testing.T) {
	ctx := acctest.Context(t)
	var i awstypes.Instance
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	stopInstance := func() {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		err := tfec2.StopEBSVolumeAttachmentInstance(ctx, conn, aws.ToString(i.InstanceId), false, 10*time.Minute)

		if err != nil {
			t.Fatal(err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_base(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachmentInstanceExists(ctx, "aws_instance.test", &i),
				),
			},
			{
				PreConfig: stopInstance,
				Config:    testAccEBSVolumeAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeviceName, "/dev/sdh"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVolumeAttachmentImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2EBSVolumeAttachment_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_update(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "force_detach", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVolumeAttachmentImportStateIDFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_detach",        // attribute only used on resource deletion
					names.AttrSkipDestroy, // attribute only used on resource deletion
				},
			},
			{
				Config: testAccEBSVolumeAttachmentConfig_update(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "force_detach", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVolumeAttachmentImportStateIDFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_detach",        // attribute only used on resource deletion
					names.AttrSkipDestroy, // attribute only used on resource deletion
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var i awstypes.Instance
	var v awstypes.Volume
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachmentInstanceExists(ctx, "aws_instance.test", &i),
					testAccCheckVolumeExists(ctx, "aws_ebs_volume.test", &v),
					testAccCheckVolumeAttachmentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVolumeAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EBSVolumeAttachment_stopInstance(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_stopInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeviceName, "/dev/sdh"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVolumeAttachmentImportStateIDFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"stop_instance_before_detaching",
				},
			},
		},
	})
}

func testAccCheckVolumeAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EBS Volume Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindEBSVolumeAttachment(ctx, conn, rs.Primary.Attributes["volume_id"], rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes[names.AttrDeviceName])

		return err
	}
}

func testAccCheckVolumeAttachmentInstanceExists(ctx context.Context, n string, v *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVolumeAttachmentInstanceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVolumeAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_volume_attachment" {
				continue
			}

			_, err := tfec2.FindEBSVolumeAttachment(ctx, conn, rs.Primary.Attributes["volume_id"], rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes[names.AttrDeviceName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EBS Volume Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVolumeAttachmentInstanceOnlyBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSVolumeAttachmentConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccVolumeAttachmentInstanceOnlyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSVolumeAttachmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEBSVolumeAttachmentConfig_base(rName), `
resource "aws_volume_attachment" "test" {
  device_name = "/dev/sdh"
  volume_id   = aws_ebs_volume.test.id
  instance_id = aws_instance.test.id
}
`)
}

func testAccEBSVolumeAttachmentConfig_stopInstance(rName string) string {
	return acctest.ConfigCompose(testAccVolumeAttachmentInstanceOnlyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1000

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "test" {
  device_name                    = "/dev/sdh"
  volume_id                      = aws_ebs_volume.test.id
  instance_id                    = aws_instance.test.id
  stop_instance_before_detaching = "true"
}
`, rName))
}

func testAccEBSVolumeAttachmentConfig_skipDestroy(rName string) string {
	return acctest.ConfigCompose(testAccEBSVolumeAttachmentConfig_base(rName), fmt.Sprintf(`
data "aws_ebs_volume" "test" {
  filter {
    name   = "size"
    values = [aws_ebs_volume.test.size]
  }

  filter {
    name   = "availability-zone"
    values = [aws_ebs_volume.test.availability_zone]
  }

  filter {
    name   = "tag:Name"
    values = ["%[1]s"]
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/sdh"
  volume_id    = data.aws_ebs_volume.test.id
  instance_id  = aws_instance.test.id
  skip_destroy = true
}
`, rName))
}

func testAccEBSVolumeAttachmentConfig_update(rName string, detach bool) string {
	return acctest.ConfigCompose(testAccEBSVolumeAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_volume_attachment" "test" {
  device_name  = "/dev/sdh"
  volume_id    = aws_ebs_volume.test.id
  instance_id  = aws_instance.test.id
  force_detach = %[1]t
  skip_destroy = %[1]t
}
`, detach))
}

func testAccVolumeAttachmentImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s:%s", rs.Primary.Attributes[names.AttrDeviceName], rs.Primary.Attributes["volume_id"], rs.Primary.Attributes[names.AttrInstanceID]), nil
	}
}
