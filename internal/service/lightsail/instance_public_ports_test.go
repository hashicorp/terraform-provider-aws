package lightsail_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccLightsailInstancePublicPorts_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance_public_ports.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstancePublicPortsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstancePublicPortsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_info.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						"protocol":  "tcp",
						"from_port": "80",
						"to_port":   "80",
					}),
				),
			},
		},
	})
}

func TestAccLightsailInstancePublicPorts_multiple(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance_public_ports.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstancePublicPortsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstancePublicPortsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_info.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						"protocol":  "tcp",
						"from_port": "80",
						"to_port":   "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						"protocol":  "tcp",
						"from_port": "443",
						"to_port":   "443",
					}),
				),
			},
		},
	})
}

func TestAccLightsailInstancePublicPorts_cidrs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance_public_ports.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstancePublicPortsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_cidrs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstancePublicPortsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_info.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						"protocol":  "tcp",
						"from_port": "125",
						"to_port":   "125",
						"cidrs.#":   "2",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "port_info.*.cidrs.*", "1.1.1.1/32"),
					resource.TestCheckTypeSetElemAttr(resourceName, "port_info.*.cidrs.*", "192.168.1.0/24"),
				),
			},
		},
	})
}

func testAccCheckInstancePublicPortsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		input := &lightsail.GetInstancePortStatesInput{
			InstanceName: aws.String(rs.Primary.Attributes["instance_name"]),
		}

		_, err := conn.GetInstancePortStates(input)

		if err != nil {
			return fmt.Errorf("error getting Lightsail Instance Public Ports (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckInstancePublicPortsDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_instance_public_ports" {
			continue
		}

		input := &lightsail.GetInstancePortStatesInput{
			InstanceName: aws.String(rs.Primary.Attributes["instance_name"]),
		}

		output, err := conn.GetInstancePortStates(input)

		if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Lightsail Instance Public Ports (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Lightsail Instance Public Ports (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccInstancePublicPortsConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}

resource "aws_lightsail_instance_public_ports" "test" {
  instance_name = aws_lightsail_instance.test.name

  port_info {
    protocol  = "tcp"
    from_port = 80
    to_port   = 80
  }
}
`, rName)
}

func testAccInstancePublicPortsConfig_multiple(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}

resource "aws_lightsail_instance_public_ports" "test" {
  instance_name = aws_lightsail_instance.test.name

  port_info {
    protocol  = "tcp"
    from_port = 80
    to_port   = 80
  }

  port_info {
    protocol  = "tcp"
    from_port = 443
    to_port   = 443
  }
}
`, rName)
}

func testAccInstancePublicPortsConfig_cidrs(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}

resource "aws_lightsail_instance_public_ports" "test" {
  instance_name = aws_lightsail_instance.test.name

  port_info {
    protocol  = "tcp"
    from_port = 125
    to_port   = 125
    cidrs     = ["192.168.1.0/24", "1.1.1.1/32"]
  }
}
`, rName)
}
