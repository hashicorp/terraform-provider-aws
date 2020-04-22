package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSVolumeAttachment_basic(t *testing.T) {
	rn := "aws_volume_attachment.ebs_att"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeAttachmentConfig,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSVolumeAttachmentImportStateIDFunc(rn),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSVolumeAttachment_skipDestroy(t *testing.T) {
	rn := "aws_volume_attachment.ebs_att"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeAttachmentConfigSkipDestroy,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSVolumeAttachmentImportStateIDFunc(rn),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"skip_destroy", // attribute only used on resource deletion
				},
			},
		},
	})
}

func TestAccAWSVolumeAttachment_attachStopped(t *testing.T) {
	var i ec2.Instance
	rn := "aws_volume_attachment.ebs_att"

	stopInstance := func() {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		_, err := conn.StopInstances(&ec2.StopInstancesInput{
			InstanceIds: []*string{i.InstanceId},
		})
		if err != nil {
			t.Fatalf("error stopping instance (%s): %s", aws.StringValue(i.InstanceId), err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"pending", "running", "stopping"},
			Target:     []string{"stopped"},
			Refresh:    InstanceStateRefreshFunc(conn, *i.InstanceId, []string{}),
			Timeout:    10 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			t.Fatalf("Error waiting for instance(%s) to stop: %s", *i.InstanceId, err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeAttachmentConfigInstanceOnly,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.web", &i),
				),
			},
			{
				PreConfig: stopInstance,
				Config:    testAccVolumeAttachmentConfig,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSVolumeAttachmentImportStateIDFunc(rn),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSVolumeAttachment_update(t *testing.T) {
	rn := "aws_volume_attachment.ebs_att"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeAttachmentConfig_update(false),
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSVolumeAttachmentImportStateIDFunc(rn),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_detach", // attribute only used on resource deletion
					"skip_destroy", // attribute only used on resource deletion
				},
			},
			{
				Config: testAccVolumeAttachmentConfig_update(true),
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSVolumeAttachmentImportStateIDFunc(rn),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_detach", // attribute only used on resource deletion
					"skip_destroy", // attribute only used on resource deletion
				},
			},
		},
	})
}

func testAccCheckVolumeAttachmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		log.Printf("\n\n----- This is never called")
		if rs.Type != "aws_volume_attachment" {
			continue
		}
	}
	return nil
}

const testAccVolumeAttachmentConfigInstanceOnly = `
resource "aws_instance" "web" {
  ami = "ami-21f78e11"
  availability_zone = "us-west-2a"
  instance_type = "t1.micro"
  tags = {
    Name = "HelloWorld"
  }
}
`

const testAccVolumeAttachmentConfig = `
resource "aws_instance" "web" {
  ami = "ami-21f78e11"
  availability_zone = "us-west-2a"
  instance_type = "t1.micro"
  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size = 1
}

resource "aws_volume_attachment" "ebs_att" {
  device_name = "/dev/sdh"
  volume_id = "${aws_ebs_volume.example.id}"
  instance_id = "${aws_instance.web.id}"
}
`

const testAccVolumeAttachmentConfigSkipDestroy = `
resource "aws_instance" "web" {
  ami = "ami-21f78e11"
  availability_zone = "us-west-2a"
  instance_type = "t1.micro"
  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size = 1
  tags = {
    Name = "TestVolume"
  }
}

data "aws_ebs_volume" "ebs_volume" {
  filter {
    name = "size"
    values = ["${aws_ebs_volume.example.size}"]
  }
  filter {
    name = "availability-zone"
    values = ["${aws_ebs_volume.example.availability_zone}"]
  }
  filter {
    name = "tag:Name"
    values = ["TestVolume"]
  }
}

resource "aws_volume_attachment" "ebs_att" {
  device_name = "/dev/sdh"
  volume_id = "${data.aws_ebs_volume.ebs_volume.id}"
  instance_id = "${aws_instance.web.id}"
  skip_destroy = true
}
`

func testAccVolumeAttachmentConfig_update(detach bool) string {
	return fmt.Sprintf(`
resource "aws_instance" "web" {
  ami               = "ami-21f78e11"
  availability_zone = "us-west-2a"
  instance_type     = "t1.micro"
}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size              = 1
}

resource "aws_volume_attachment" "ebs_att" {
  device_name  = "/dev/sdh"
  volume_id    = "${aws_ebs_volume.example.id}"
  instance_id  = "${aws_instance.web.id}"
  force_detach = %t
  skip_destroy = %t
}
`, detach, detach)
}

func testAccAWSVolumeAttachmentImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s:%s", rs.Primary.Attributes["device_name"], rs.Primary.Attributes["volume_id"], rs.Primary.Attributes["instance_id"]), nil
	}
}
