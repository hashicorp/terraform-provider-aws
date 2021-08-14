package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRoute53RecoveryControlConfigRoutingControl_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoverycontrolconfig_routing_control.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAwsRoute53RecoveryControlConfigRoutingControlConfig(t) },
		ErrorCheck:        testAccErrorCheck(t, route53recoverycontrolconfig.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryControlConfigRoutingControlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigRoutingControlConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigRoutingControlExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "routing_control_arn", "route53recoverycontrolconfig", regexp.MustCompile(`routingcontrol/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
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

func testAccPreCheckAwsRoute53RecoveryControlConfigRoutingControlConfig(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.ListRoutingControlsInput{}

	_, err := conn.ListRoutingControls(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAwsRoute53RecoveryControlConfigRoutingControlDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoverycontrolconfig_routing_control" {
			continue
		}

		input := &route53recoverycontrolconfig.DescribeRoutingControlInput{
			RoutingControlArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeRoutingControl(input)
		if err == nil {
			return fmt.Errorf("Route53RecoveryControlConfig Routing Control (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsRoute53RecoveryControlConfigRoutingControlConfig(rName string) string {
	return fmt.Sprintf(`
	resource "aws_route53recoverycontrolconfig_routing_control" "test" {
	  name = %q
	}
	`, rName)
}

func testAccCheckAwsRoute53RecoveryControlConfigRoutingControlExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

		input := &route53recoverycontrolconfig.DescribeRoutingControlInput{
			RoutingControlArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeRoutingControl(input)

		return err
	}
}
