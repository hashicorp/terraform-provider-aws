package globalaccelerator_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGlobalAcceleratorEndpointGroup_basic(t *testing.T) {
	var v globalaccelerator.EndpointGroup
	resourceName := "aws_globalaccelerator_endpoint_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+/endpoint-group/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", ""),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "TCP"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "listener_arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
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

func TestAccGlobalAcceleratorEndpointGroup_disappears(t *testing.T) {
	var v globalaccelerator.EndpointGroup
	resourceName := "aws_globalaccelerator_endpoint_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfglobalaccelerator.ResourceEndpointGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlobalAcceleratorEndpointGroup_ALBEndpoint_clientIP(t *testing.T) {
	var v globalaccelerator.EndpointGroup
	var vpc ec2.Vpc
	resourceName := "aws_globalaccelerator_endpoint_group.test"
	albResourceName := "aws_lb.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointGroupConfig_albClientIP(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+/endpoint-group/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "endpoint_configuration.*", map[string]string{
						"client_ip_preservation_enabled": "false",
						"weight":                         "20",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.*.endpoint_id", albResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/"),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "HTTP"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "listener_arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEndpointGroupConfig_albClientIP(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+/endpoint-group/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "endpoint_configuration.*", map[string]string{
						"client_ip_preservation_enabled": "true",
						"weight":                         "20",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.*.endpoint_id", albResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/"),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "HTTP"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "listener_arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
				),
			},
			{
				Config: testAccEndpointGroupConfig_baseVPC(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(vpcResourceName, &vpc),
					testAccCheckEndpointGroupDeleteSecurityGroup(&vpc),
				),
			},
		},
	})
}

func TestAccGlobalAcceleratorEndpointGroup_instanceEndpoint(t *testing.T) {
	var v globalaccelerator.EndpointGroup
	var vpc ec2.Vpc
	resourceName := "aws_globalaccelerator_endpoint_group.test"
	instanceResourceName := "aws_instance.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointGroupConfig_instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+/endpoint-group/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "endpoint_configuration.*", map[string]string{
						"client_ip_preservation_enabled": "true",
						"weight":                         "20",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.*.endpoint_id", instanceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/"),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "HTTP"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "listener_arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEndpointGroupConfig_baseVPC(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(vpcResourceName, &vpc),
					testAccCheckEndpointGroupDeleteSecurityGroup(&vpc),
				),
			},
		},
	})
}

func TestAccGlobalAcceleratorEndpointGroup_multiRegion(t *testing.T) {
	var providers []*schema.Provider
	var v globalaccelerator.EndpointGroup
	resourceName := "aws_globalaccelerator_endpoint_group.test"
	eipResourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckMultipleRegion(t, 2); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointGroupConfig_multiRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+/endpoint-group/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "endpoint_configuration.*", map[string]string{
						"client_ip_preservation_enabled": "false",
						"weight":                         "20",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.*.endpoint_id", eipResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/foo"),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "HTTPS"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "listener_arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "0"),
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

func TestAccGlobalAcceleratorEndpointGroup_portOverrides(t *testing.T) {
	var v globalaccelerator.EndpointGroup
	resourceName := "aws_globalaccelerator_endpoint_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointGroupConfig_portOverrides(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+/endpoint-group/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", ""),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "TCP"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "listener_arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"endpoint_port": "8081",
						"listener_port": "81",
					}),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
				),
			},
			{
				Config: testAccEndpointGroupConfig_portOverridesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+/endpoint-group/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", ""),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "TCP"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "listener_arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"endpoint_port": "8081",
						"listener_port": "81",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"endpoint_port": "9090",
						"listener_port": "90",
					}),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
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

func TestAccGlobalAcceleratorEndpointGroup_tcpHealthCheckProtocol(t *testing.T) {
	var v globalaccelerator.EndpointGroup
	resourceName := "aws_globalaccelerator_endpoint_group.test"
	eipResourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointGroupConfig_tcpHealthCheckProtocol(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+/endpoint-group/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "endpoint_configuration.*", map[string]string{
						"client_ip_preservation_enabled": "false",
						"weight":                         "10",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.*.endpoint_id", eipResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "1234"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "TCP"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "listener_arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
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

func TestAccGlobalAcceleratorEndpointGroup_update(t *testing.T) {
	var v globalaccelerator.EndpointGroup
	resourceName := "aws_globalaccelerator_endpoint_group.test"
	eipResourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+/endpoint-group/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", ""),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "TCP"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "listener_arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "100"),
				),
			},
			{
				Config: testAccEndpointGroupConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointGroupExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+/endpoint-group/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "endpoint_configuration.*", map[string]string{
						"client_ip_preservation_enabled": "false",
						"weight":                         "20",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.*.endpoint_id", eipResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "health_check_interval_seconds", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/foo"),
					resource.TestCheckResourceAttr(resourceName, "health_check_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "health_check_protocol", "HTTPS"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "listener_arn", "globalaccelerator", regexp.MustCompile(`accelerator/[^/]+/listener/[^/]+`)),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "threshold_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_dial_percentage", "0"),
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

func testAccCheckEndpointGroupExists(name string, v *globalaccelerator.EndpointGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Global Accelerator endpoint group ID is set")
		}

		endpointGroup, err := tfglobalaccelerator.FindEndpointGroupByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *endpointGroup

		return nil
	}
}

func testAccCheckEndpointGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_globalaccelerator_endpoint_group" {
			continue
		}

		_, err := tfglobalaccelerator.FindEndpointGroupByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Global Accelerator endpoint group %s still exists", rs.Primary.ID)
	}
	return nil
}

// testAccCheckEndpointGroupDeleteSecurityGroup deletes the security group
// placed into the VPC when Global Accelerator client IP address preservation is enabled.
func testAccCheckEndpointGroupDeleteSecurityGroup(vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		sg, err := tfec2.FindSecurityGroupByNameAndVPCID(conn, "GlobalAccelerator", aws.StringValue(vpc.VpcId))
		if tfresource.NotFound(err) {
			// Already gone.
			return nil
		}
		if err != nil {
			return err
		}

		_, err = conn.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
			GroupId: sg.GroupId,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccEndpointGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "test" {
  accelerator_arn = aws_globalaccelerator_accelerator.test.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}

resource "aws_globalaccelerator_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_listener.test.id
}
`, rName)
}

func testAccEndpointGroupConfig_baseVPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccEndpointGroupConfig_albClientIP(rName string, clientIP bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		testAccEndpointGroupConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = false
  security_groups = [aws_security_group.test.id]
  subnets         = [aws_subnet.test.*.id[0], aws_subnet.test.*.id[1]]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list
}

resource "aws_subnet" "test" {
  count             = length(var.subnets)
  vpc_id            = aws_vpc.test.id
  cidr_block        = element(var.subnets, count.index)
  availability_zone = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

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
  vpc_id = aws_vpc.test.id

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
  accelerator_arn = aws_globalaccelerator_accelerator.test.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}

resource "aws_globalaccelerator_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_listener.test.id

  endpoint_configuration {
    endpoint_id                    = aws_lb.test.id
    weight                         = 20
    client_ip_preservation_enabled = %[2]t
  }

  health_check_interval_seconds = 30
  health_check_path             = "/"
  health_check_port             = 80
  health_check_protocol         = "HTTP"
  threshold_count               = 3
  traffic_dial_percentage       = 100
}
`, rName, clientIP))
}

func testAccEndpointGroupConfig_instance(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccEndpointGroupConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

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
  accelerator_arn = aws_globalaccelerator_accelerator.test.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}

resource "aws_globalaccelerator_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_listener.test.id

  endpoint_configuration {
    endpoint_id                    = aws_instance.test.id
    weight                         = 20
    client_ip_preservation_enabled = true
  }

  health_check_interval_seconds = 30
  health_check_path             = "/"
  health_check_port             = 80
  health_check_protocol         = "HTTP"
  threshold_count               = 3
  traffic_dial_percentage       = 100
}
`, rName))
}

func testAccEndpointGroupConfig_multiRegion(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "test" {
  accelerator_arn = aws_globalaccelerator_accelerator.test.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}

resource "aws_eip" "test" {
  provider = "awsalternate"

  vpc = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_globalaccelerator_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_listener.test.id

  endpoint_configuration {
    endpoint_id = aws_eip.test.id
    weight      = 20
  }

  endpoint_group_region         = %[2]q
  health_check_interval_seconds = 10
  health_check_path             = "/foo"
  health_check_port             = 8080
  health_check_protocol         = "HTTPS"
  threshold_count               = 1
  traffic_dial_percentage       = 0
}
`, rName, acctest.AlternateRegion()))
}

func testAccEndpointGroupConfig_portOverrides(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "test" {
  accelerator_arn = aws_globalaccelerator_accelerator.test.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 90
  }
}

resource "aws_globalaccelerator_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_listener.test.id

  health_check_port = 80

  port_override {
    endpoint_port = 8081
    listener_port = 81
  }
}
`, rName)
}

func testAccEndpointGroupConfig_portOverridesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "test" {
  accelerator_arn = aws_globalaccelerator_accelerator.test.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 90
  }
}

resource "aws_globalaccelerator_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_listener.test.id

  port_override {
    endpoint_port = 8081
    listener_port = 81
  }

  port_override {
    endpoint_port = 9090
    listener_port = 90
  }
}
`, rName)
}

func testAccEndpointGroupConfig_tcpHealthCheckProtocol(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "test" {
  accelerator_arn = aws_globalaccelerator_accelerator.test.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}

resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_globalaccelerator_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_listener.test.id

  endpoint_configuration {
    endpoint_id = aws_eip.test.id
    weight      = 10
  }

  endpoint_group_region         = data.aws_region.current.name
  health_check_interval_seconds = 30
  health_check_port             = 1234
  health_check_protocol         = "TCP"
  threshold_count               = 3
  traffic_dial_percentage       = 100
}
`, rName)
}

func testAccEndpointGroupConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "test" {
  accelerator_arn = aws_globalaccelerator_accelerator.test.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}

resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_globalaccelerator_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_listener.test.id

  endpoint_configuration {
    endpoint_id = aws_eip.test.id
    weight      = 20
  }

  health_check_interval_seconds = 10
  health_check_path             = "/foo"
  health_check_port             = 8080
  health_check_protocol         = "HTTPS"
  threshold_count               = 1
  traffic_dial_percentage       = 0
}
`, rName)
}
