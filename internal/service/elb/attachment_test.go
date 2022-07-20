package elb_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccELBAttachment_basic(t *testing.T) {
	var conf elb.LoadBalancerDescription
	resourceName := "aws_elb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccAttachmentCheckInstanceCount(&conf, 1),
				),
			},
			{
				Config: testAccAttachmentConfig_2(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccAttachmentCheckInstanceCount(&conf, 2),
				),
			},
			{
				Config: testAccAttachmentConfig_3(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccAttachmentCheckInstanceCount(&conf, 2),
				),
			},
			{
				Config: testAccAttachmentConfig_4(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccAttachmentCheckInstanceCount(&conf, 0),
				),
			},
		},
	})
}

// remove and instance and check that it's correctly re-attached.
func TestAccELBAttachment_drift(t *testing.T) {
	var conf elb.LoadBalancerDescription
	resourceName := "aws_elb.test"

	testAccAttachmentConfig_deregInstance := func() {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn

		deRegisterInstancesOpts := elb.DeregisterInstancesFromLoadBalancerInput{
			LoadBalancerName: conf.LoadBalancerName,
			Instances:        conf.Instances,
		}

		log.Printf("[DEBUG] deregistering instance %v from ELB", *conf.Instances[0].InstanceId)

		_, err := conn.DeregisterInstancesFromLoadBalancer(&deRegisterInstancesOpts)
		if err != nil {
			t.Fatalf("Failure deregistering instances from ELB: %s", err)
		}

	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccAttachmentCheckInstanceCount(&conf, 1),
				),
			},
			// remove an instance from the ELB, and make sure it gets re-added
			{
				Config:    testAccAttachmentConfig_1(),
				PreConfig: testAccAttachmentConfig_deregInstance,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccAttachmentCheckInstanceCount(&conf, 1),
				),
			},
		},
	})
}

func testAccAttachmentCheckInstanceCount(conf *elb.LoadBalancerDescription, expected int) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if actual := len(conf.Instances); actual != expected {
			return fmt.Errorf("instance count does not match: expected %d, got %d", expected, actual)
		}
		return nil
	}
}

// add one attachment
func testAccAttachmentConfig_1() string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "test" {
  availability_zones = data.aws_availability_zones.available.names

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_instance" "foo1" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}

resource "aws_elb_attachment" "foo1" {
  elb      = aws_elb.test.id
  instance = aws_instance.foo1.id
}
`)
}

// add a second attachment
func testAccAttachmentConfig_2() string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "test" {
  availability_zones = data.aws_availability_zones.available.names

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_instance" "foo1" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}

resource "aws_instance" "foo2" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}

resource "aws_elb_attachment" "foo1" {
  elb      = aws_elb.test.id
  instance = aws_instance.foo1.id
}

resource "aws_elb_attachment" "foo2" {
  elb      = aws_elb.test.id
  instance = aws_instance.foo2.id
}
`)
}

// swap attachments between resources
func testAccAttachmentConfig_3() string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "test" {
  availability_zones = data.aws_availability_zones.available.names

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_instance" "foo1" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}

resource "aws_instance" "foo2" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}

resource "aws_elb_attachment" "foo1" {
  elb      = aws_elb.test.id
  instance = aws_instance.foo2.id
}

resource "aws_elb_attachment" "foo2" {
  elb      = aws_elb.test.id
  instance = aws_instance.foo1.id
}
`)
}

// destroy attachments
func testAccAttachmentConfig_4() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "test" {
  availability_zones = data.aws_availability_zones.available.names

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}
`
}
