package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	resourceName := "aws_globalaccelerator_endpoint_group.example"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorEndpointGroup_alb_clientip(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorEndpointGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/"),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					testAccCheckGlobalAcceleratorEndpointGroupConfig(resourceName, "client_ip_preservation_enabled",
						"false"),
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

func testAccGlobalAcceleratorEndpointGroup_alb_clientip(rInt int) string {
	return fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%d"
  internal        = false
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_internet_gateway" "example" {
  vpc_id = "${aws_vpc.alb_test.id}"
}

resource "aws_globalaccelerator_accelerator" "example" {
  name            = "tf-%d"
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = "${aws_globalaccelerator_accelerator.example.id}"
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}

resource "aws_globalaccelerator_endpoint_group" "example" {
  listener_arn = "${aws_globalaccelerator_listener.example.id}"

  endpoint_configuration {
    endpoint_id = "${aws_lb.lb_test.id}"
    weight      = 20
	client_ip_preservation_enabled = false
  }

  health_check_interval_seconds = 30
  health_check_path             = "/"
  health_check_port             = 80
  health_check_protocol         = "HTTP"
  threshold_count               = 3
  traffic_dial_percentage       = 100
}
`, rInt, rInt)
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

func testAccCheckGlobalAcceleratorEndpointGroupConfig(n, k, v string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		r := fmt.Sprintf(`endpoint_configuration.\d+.%s`, k)
		reg, err := regexp.Compile(r)
		if err != nil {
			return fmt.Errorf("Regular Express not correct err: %+v", err)
		}
		for configKey, configValue := range rs.Primary.Attributes {
			if reg.MatchString(configKey) {
				if configValue == v {
					return nil
				} else {
					return fmt.Errorf("endpoint_configuration key: %s value does not match.  Expected: %s,"+
						" Got: %s", configKey, v, configValue)
				}
			}
		}

		// Failed to find value
		return fmt.Errorf("endpoint_configuration is missing key: %s", k)
	}
}
