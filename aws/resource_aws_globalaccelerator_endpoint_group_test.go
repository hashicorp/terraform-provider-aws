package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccAwsGlobalAcceleratorEndpointGroup_basic(t *testing.T) {
	resourceName := "aws_globalaccelerator_endpoint_group.example"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorEndpointGroup_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorEndpointGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/"),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
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

func TestAccAwsGlobalAcceleratorEndpointGroup_alb_clientip(t *testing.T) {
	resourceName := "aws_globalaccelerator_endpoint_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorEndpointGroup_alb_clientip(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorEndpointGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/"),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "endpoint_configuration.*", map[string]string{
						"client_ip_preservation_enabled": "true",
					}),
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

func TestAccAwsGlobalAcceleratorEndpointGroup_update(t *testing.T) {
	resourceName := "aws_globalaccelerator_endpoint_group.example"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorEndpointGroup_basic(rInt),
			},
			{
				Config: testAccGlobalAcceleratorEndpointGroup_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorEndpointGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/foo"),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
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

func testAccCheckGlobalAcceleratorEndpointGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).globalacceleratorconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		accelerator, err := resourceAwsGlobalAcceleratorEndpointGroupRetrieve(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if accelerator == nil {
			return fmt.Errorf("Global Accelerator endpoint group not found")
		}

		return nil
	}
}

func testAccCheckGlobalAcceleratorEndpointGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).globalacceleratorconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_globalaccelerator_endpoint_group" {
			continue
		}

		accelerator, err := resourceAwsGlobalAcceleratorEndpointGroupRetrieve(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if accelerator != nil {
			return fmt.Errorf("Global Accelerator endpoint group still exists")
		}
	}
	return nil
}

func testAccGlobalAcceleratorEndpointGroup_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "tf-%d"
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = aws_globalaccelerator_accelerator.example.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}

data "aws_region" "current" {}

resource "aws_eip" "example" {}

resource "aws_globalaccelerator_endpoint_group" "example" {
  listener_arn = aws_globalaccelerator_listener.example.id

  endpoint_configuration {
    endpoint_id = aws_eip.example.id
    weight      = 10
  }

  endpoint_group_region         = data.aws_region.current.name
  health_check_interval_seconds = 30
  health_check_path             = "/"
  health_check_port             = 80
  health_check_protocol         = "HTTP"
  threshold_count               = 3
  traffic_dial_percentage       = 100
}
`, rInt)
}

func testAccGlobalAcceleratorEndpointGroup_alb_clientip(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = false
  security_groups = ["${aws_security_group.test.id}"]
  subnets         = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "${element(var.subnets, count.index)}"
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = "${aws_vpc.test.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "test" {
  accelerator_arn = "${aws_globalaccelerator_accelerator.test.id}"
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}

resource "aws_globalaccelerator_endpoint_group" "test" {
  listener_arn = "${aws_globalaccelerator_listener.test.id}"

  endpoint_configuration {
    endpoint_id = "${aws_lb.test.id}"
    weight      = 20
    client_ip_preservation_enabled = true
  }

  health_check_interval_seconds = 30
  health_check_path             = "/"
  health_check_port             = 80
  health_check_protocol         = "HTTP"
  threshold_count               = 3
  traffic_dial_percentage       = 100
}
`, rName)
}

func testAccGlobalAcceleratorEndpointGroup_update(rInt int) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "tf-%d"
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = aws_globalaccelerator_accelerator.example.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}

data "aws_region" "current" {}

resource "aws_eip" "example" {}

resource "aws_globalaccelerator_endpoint_group" "example" {
  listener_arn = aws_globalaccelerator_listener.example.id

  endpoint_configuration {
    endpoint_id = aws_eip.example.id
    weight      = 20
  }

  endpoint_group_region         = data.aws_region.current.name
  health_check_interval_seconds = 10
  health_check_path             = "/foo"
  health_check_port             = 8080
  health_check_protocol         = "HTTPS"
  threshold_count               = 1
  traffic_dial_percentage       = 0
}
`, rInt)
}
