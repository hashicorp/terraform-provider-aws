package amp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAMPAlertManagerDefinition_basic(t *testing.T) {
	resourceName := "aws_prometheus_alert_manager_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(prometheusservice.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAlertManagerDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertManagerDefinitionConfig_basic(defaultAlertManagerDefinition()),
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
				Config: testAccAlertManagerDefinitionConfig_basic(anotherAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "definition", anotherAlertManagerDefinition()),
				),
			},
			{
				Config: testAccAlertManagerDefinitionConfig_basic(defaultAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "definition", defaultAlertManagerDefinition()),
				),
			},
		},
	})
}

func TestAccAMPAlertManagerDefinition_disappears(t *testing.T) {
	resourceName := "aws_prometheus_alert_manager_definition.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(prometheusservice.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAlertManagerDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertManagerDefinitionConfig_basic(defaultAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfamp.ResourceAlertManagerDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAlertManagerDefinitionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Prometheus Alert Manager Definition ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPConn

		_, err := tfamp.FindAlertManagerDefinitionByID(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAlertManagerDefinitionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AMPConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_prometheus_alert_manager_definition" {
			continue
		}

		_, err := tfamp.FindAlertManagerDefinitionByID(context.TODO(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Prometheus Alert Manager Definition %s still exists", rs.Primary.ID)
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

func testAccAlertManagerDefinitionConfig_basic(definition string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
}
resource "aws_prometheus_alert_manager_definition" "test" {
  workspace_id = aws_prometheus_workspace.test.id
  definition   = %[1]q
}
`, definition)
}
