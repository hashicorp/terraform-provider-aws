// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailInstancePublicPorts_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance_public_ports.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstancePublicPortsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstancePublicPortsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_info.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						names.AttrProtocol: "tcp",
						"from_port":        "80",
						"to_port":          "80",
					}),
				),
			},
		},
	})
}

func TestAccLightsailInstancePublicPorts_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance_public_ports.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstancePublicPortsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstancePublicPortsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_info.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						names.AttrProtocol: "tcp",
						"from_port":        "80",
						"to_port":          "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						names.AttrProtocol: "tcp",
						"from_port":        "443",
						"to_port":          "443",
					}),
				),
			},
		},
	})
}

func TestAccLightsailInstancePublicPorts_cidrs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance_public_ports.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstancePublicPortsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_cidrs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstancePublicPortsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_info.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						names.AttrProtocol: "tcp",
						"from_port":        "125",
						"to_port":          "125",
						"cidrs.#":          "2",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "port_info.*.cidrs.*", "1.1.1.1/32"),
					resource.TestCheckTypeSetElemAttr(resourceName, "port_info.*.cidrs.*", "192.168.1.0/24"),
				),
			},
		},
	})
}

func TestAccLightsailInstancePublicPorts_cidrListAliases(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance_public_ports.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstancePublicPortsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_cidrListAliases(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstancePublicPortsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_info.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						names.AttrProtocol:    "tcp",
						"from_port":           "22",
						"to_port":             "22",
						"cidr_list_aliases.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "port_info.*.cidr_list_aliases.*", "lightsail-connect"),
				),
			},
		},
	})
}

func TestAccLightsailInstancePublicPorts_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_instance_public_ports.test"
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstancePublicPortsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstancePublicPortsExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflightsail.ResourceInstancePublicPorts(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLightsailInstancePublicPorts_disappears_Instance(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	parentResourceName := "aws_lightsail_instance.test"
	resourceName := "aws_lightsail_instance_public_ports.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstancePublicPortsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstancePublicPortsExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflightsail.ResourceInstance(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLightsailInstancePublicPorts_icmpPing(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance_public_ports.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstancePublicPortsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_icmpPing(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstancePublicPortsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_info.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						"cidrs.0":          "0.0.0.0/0",
						names.AttrProtocol: "icmp",
						"from_port":        "8",
						"to_port":          "-1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						"ipv6_cidrs.0":     "::/0",
						names.AttrProtocol: "icmpv6",
						"from_port":        "128",
						"to_port":          "0",
					}),
				),
			},
		},
	})
}

func TestAccLightsailInstancePublicPorts_icmpAll(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_instance_public_ports.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstancePublicPortsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstancePublicPortsConfig_icmpAll(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstancePublicPortsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_info.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						"cidrs.0":          "0.0.0.0/0",
						names.AttrProtocol: "icmp",
						"from_port":        "-1",
						"to_port":          "-1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_info.*", map[string]string{
						"ipv6_cidrs.0":     "::/0",
						names.AttrProtocol: "icmpv6",
						"from_port":        "-1",
						"to_port":          "-1",
					}),
				),
			},
		},
	})
}

func testAccCheckInstancePublicPortsExists(ctx context.Context, t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

		input := &lightsail.GetInstancePortStatesInput{
			InstanceName: aws.String(rs.Primary.Attributes["instance_name"]),
		}

		_, err := conn.GetInstancePortStates(ctx, input)

		if err != nil {
			return fmt.Errorf("error getting Lightsail Instance Public Ports (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckInstancePublicPortsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_instance_public_ports" {
				continue
			}

			input := &lightsail.GetInstancePortStatesInput{
				InstanceName: aws.String(rs.Primary.Attributes["instance_name"]),
			}

			output, err := conn.GetInstancePortStates(ctx, input)

			if tflightsail.IsANotFoundError(err) {
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
}

func testAccInstancePublicPortsConfigBase(rName string) string {
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
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}
`, rName)
}

func testAccInstancePublicPortsConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInstancePublicPortsConfigBase(rName),
		`	
resource "aws_lightsail_instance_public_ports" "test" {
  instance_name = aws_lightsail_instance.test.name

  port_info {
    protocol  = "tcp"
    from_port = 80
    to_port   = 80
  }
}
`)
}

func testAccInstancePublicPortsConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccInstancePublicPortsConfigBase(rName),
		`
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
`)
}

func testAccInstancePublicPortsConfig_cidrs(rName string) string {
	return acctest.ConfigCompose(
		testAccInstancePublicPortsConfigBase(rName),
		`
resource "aws_lightsail_instance_public_ports" "test" {
  instance_name = aws_lightsail_instance.test.name

  port_info {
    protocol  = "tcp"
    from_port = 125
    to_port   = 125
    cidrs     = ["192.168.1.0/24", "1.1.1.1/32"]
  }
}
`)
}

func testAccInstancePublicPortsConfig_cidrListAliases(rName string) string {
	return acctest.ConfigCompose(
		testAccInstancePublicPortsConfigBase(rName),
		`
resource "aws_lightsail_instance_public_ports" "test" {
  instance_name = aws_lightsail_instance.test.name

  port_info {
    protocol          = "tcp"
    from_port         = 22
    to_port           = 22
    cidr_list_aliases = ["lightsail-connect"]
  }
}
`)
}

func testAccInstancePublicPortsConfig_icmpPing(rName string) string {
	return acctest.ConfigCompose(
		testAccInstancePublicPortsConfigBase(rName),
		`
resource "aws_lightsail_instance_public_ports" "test" {
  instance_name = aws_lightsail_instance.test.name

  port_info {
    cidrs     = ["0.0.0.0/0"]
    protocol  = "icmp"
    from_port = 8
    to_port   = -1
  }

  port_info {
    ipv6_cidrs = ["::/0"]
    protocol   = "icmpv6"
    from_port  = 128
    to_port    = 0
  }
}
`)
}

func testAccInstancePublicPortsConfig_icmpAll(rName string) string {
	return acctest.ConfigCompose(
		testAccInstancePublicPortsConfigBase(rName),
		`
resource "aws_lightsail_instance_public_ports" "test" {
  instance_name = aws_lightsail_instance.test.name

  port_info {
    cidrs     = ["0.0.0.0/0"]
    protocol  = "icmp"
    from_port = -1
    to_port   = -1
  }

  port_info {
    ipv6_cidrs = ["::/0"]
    protocol   = "icmpv6"
    from_port  = -1
    to_port    = -1
  }
}
`)
}
