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

func TestAccAWSRoute53RecoveryControlConfigControlPanel_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoverycontrolconfig_control_panel.test"
	clusterArn := "someClusterArn"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAwsRoute53RecoveryControlConfigControlPanel(t) },
		ErrorCheck:        testAccErrorCheck(t, route53recoverycontrolconfig.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryControlConfigControlPanelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigControlPanelConfig(rName, clusterArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigControlPanelExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "control_panel_arn", "route53recoverycontrolconfig", regexp.MustCompile(`controlpanel/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "default_control_panel", "false"),
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

func testAccPreCheckAwsRoute53RecoveryControlConfigControlPanel(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.ListControlPanelsInput{}

	_, err := conn.ListControlPanels(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAwsRoute53RecoveryControlConfigControlPanelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoverycontrolconfig_control_panel" {
			continue
		}

		input := &route53recoverycontrolconfig.DescribeControlPanelInput{
			ControlPanelArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeControlPanel(input)
		if err == nil {
			return fmt.Errorf("Route53RecoveryControlConfig Control Panel (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsRoute53RecoveryControlConfigControlPanelConfig(rName, clusterArn string) string {
	return fmt.Sprintf(`
	resource "aws_route53recoverycontrolconfig_control_panel" "test" {
	  name = %[1]q
	  cluster_arn = %[2]q
	}
	`, rName, clusterArn)
}

func testAccCheckAwsRoute53RecoveryControlConfigControlPanelExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

		input := &route53recoverycontrolconfig.DescribeControlPanelInput{
			ControlPanelArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeControlPanel(input)

		return err
	}
}
