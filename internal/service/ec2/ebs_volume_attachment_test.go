package ec2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2EBSVolumeAttachment_basic(t *testing.T) {
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "device_name", "/dev/sdh"),
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
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "device_name", "/dev/sdh"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVolumeAttachmentImportStateIDFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"skip_destroy", // attribute only used on resource deletion
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeAttachment_attachStopped(t *testing.T) {
	var i ec2.Instance
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	stopInstance := func() {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		err := tfec2.StopInstance(conn, aws.StringValue(i.InstanceId), 10*time.Minute)

		if err != nil {
			t.Fatal(err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_base(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.test", &i),
				),
			},
			{
				PreConfig: stopInstance,
				Config:    testAccEBSVolumeAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "device_name", "/dev/sdh"),
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
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_update(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "force_detach", "false"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVolumeAttachmentImportStateIDFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_detach", // attribute only used on resource deletion
					"skip_destroy", // attribute only used on resource deletion
				},
			},
			{
				Config: testAccEBSVolumeAttachmentConfig_update(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "force_detach", "true"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVolumeAttachmentImportStateIDFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_detach", // attribute only used on resource deletion
					"skip_destroy", // attribute only used on resource deletion
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeAttachment_disappears(t *testing.T) {
	var i ec2.Instance
	var v ec2.Volume
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.test", &i),
					testAccCheckVolumeExists("aws_ebs_volume.test", &v),
					testAccCheckVolumeAttachmentExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVolumeAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EBSVolumeAttachment_stopInstance(t *testing.T) {
	resourceName := "aws_volume_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeAttachmentConfig_stopInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "device_name", "/dev/sdh"),
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

func testAccCheckVolumeAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EBS Volume Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		_, err := tfec2.FindEBSVolumeAttachment(conn, rs.Primary.Attributes["volume_id"], rs.Primary.Attributes["instance_id"], rs.Primary.Attributes["device_name"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckVolumeAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_volume_attachment" {
			continue
		}

		_, err := tfec2.FindEBSVolumeAttachment(conn, rs.Primary.Attributes["volume_id"], rs.Primary.Attributes["instance_id"], rs.Primary.Attributes["device_name"])

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

func testAccVolumeAttachmentInstanceOnlyBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		return fmt.Sprintf("%s:%s:%s", rs.Primary.Attributes["device_name"], rs.Primary.Attributes["volume_id"], rs.Primary.Attributes["instance_id"]), nil
	}
}
