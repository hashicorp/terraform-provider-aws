package prometheus_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfprometheus "github.com/hashicorp/terraform-provider-aws/internal/service/prometheus"
)

func TestAccPrometheusAlertManagerDefinition_basic(t *testing.T) {
	resourceName := "aws_prometheus_alert_manager_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAMPAlertManagerDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMPAlertManagerDefinition(defaultAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "definition", defaultAlertManagerDefinition()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAMPAlertManagerDefinition(anotherAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "definition", anotherAlertManagerDefinition()),
				),
			},
			{
				Config: testAccAMPAlertManagerDefinition(defaultAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "definition", defaultAlertManagerDefinition()),
				),
			},
		},
	})
}

func TestAccPrometheusAlertManagerDefinition_disappears(t *testing.T) {
	resourceName := "aws_prometheus_alert_manager_definition.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAMPAlertManagerDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMPAlertManagerDefinition(defaultAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfprometheus.ResourceAlertManagerDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAMPAlertManagerDefinitionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PrometheusConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_prometheus_alert_manager_definition" {
			continue
		}

		_, err := conn.DescribeAlertManagerDefinition(&prometheusservice.DescribeAlertManagerDefinitionInput{
			WorkspaceId: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrMessageContains(err, prometheusservice.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Prometheus Alert manager definition (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func defaultAlertManagerDefinition() string {
	return `
alertmanager_config: |
  route:
    receiver: 'default'
  receivers:
    - name: 'default'
`
}

func anotherAlertManagerDefinition() string {
	return `
alertmanager_config: |
  route:
    receiver: 'default2'
  receivers:
    - name: 'default2'
`
}

func testAccAMPAlertManagerDefinition(definition string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
}
resource "aws_prometheus_alert_manager_definition" "test" {
  workspace_id = aws_prometheus_workspace.test.id
  definition   = <<EOF
%sEOF
}
`, definition)
}

func testAccCheckAlertManagerDefinitionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AMP alert manager definition id is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PrometheusConn

		req := &prometheusservice.DescribeAlertManagerDefinitionInput{
			WorkspaceId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.DescribeAlertManagerDefinition(req)
		if err != nil {
			return err
		}
		if describe == nil {
			return fmt.Errorf("Got nil account ?!")
		}

		return nil
	}
}
