package route53recoverycontrolconfig_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53recoverycontrolconfig "github.com/hashicorp/terraform-provider-aws/internal/service/route53recoverycontrolconfig"
)

func testAccRoutingControl_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_routing_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(r53rcc.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, r53rcc.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoutingControlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingControlConfig_inDefaultPanel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingControlExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_arn", // not available in DescribeRoutingControlOutput
				},
			},
		},
	})
}

func testAccRoutingControl_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_routing_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(r53rcc.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, r53rcc.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoutingControlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingControlConfig_inDefaultPanel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingControlExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53recoverycontrolconfig.ResourceRoutingControl(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRoutingControl_nonDefaultControlPanel(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_routing_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(r53rcc.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, r53rcc.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoutingControlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingControlConfig_inNonDefaultPanel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingControlExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
				),
			},
		},
	})
}

func testAccCheckRoutingControlExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryControlConfigConn

		input := &r53rcc.DescribeRoutingControlInput{
			RoutingControlArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeRoutingControl(input)

		return err
	}
}

func testAccCheckRoutingControlDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryControlConfigConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoverycontrolconfig_routing_control" {
			continue
		}

		input := &r53rcc.DescribeRoutingControlInput{
			RoutingControlArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeRoutingControl(input)

		if err == nil {
			return fmt.Errorf("Route53RecoveryControlConfig Routing Control (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccClusterBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_cluster" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRoutingControlConfig_inDefaultPanel(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBase(rName), fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_routing_control" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}
`, rName))
}

func testAccControlPanelBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_control_panel" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}
`, rName)
}

func testAccRoutingControlConfig_inNonDefaultPanel(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBase(rName),
		testAccControlPanelBase(rName),
		fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_routing_control" "test" {
  name              = %[1]q
  cluster_arn       = aws_route53recoverycontrolconfig_cluster.test.arn
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test.arn
}
`, rName))
}
