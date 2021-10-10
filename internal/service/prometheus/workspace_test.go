package prometheus_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfprometheus "github.com/hashicorp/terraform-provider-aws/internal/service/prometheus"
)

func TestAccPrometheusWorkspace_AMP_basic(t *testing.T) {
	workspaceAlias := sdkacctest.RandomWithPrefix("tf_amp_workspace")
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAMPWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMPWorkspaceWithAliasConfig(workspaceAlias),
				Check: resource.ComposeTestCheckFunc(
					testCheckAMPWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alias", workspaceAlias),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAMPWorkspaceWithoutAliasConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "alias", ""),
				),
			},
			{
				Config: testAccAMPWorkspaceWithAliasConfig(workspaceAlias),
				Check: resource.ComposeTestCheckFunc(
					testCheckAMPWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alias", workspaceAlias),
				),
			},
			{
				Config: testAWSAMPWorkspaceConfigWithAlertManagerDefinition("test"),
				Check: resource.ComposeTestCheckFunc(
					testCheckAWSAMPAlertManagerExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "alert_manager_definition"),
				),
			},
			{
				Config: testAWSAMPWorkspaceConfigWithAlertManagerDefinition("update"),
				Check: resource.ComposeTestCheckFunc(
					testCheckAWSAMPAlertManagerExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "alert_manager_definition"),
				),
			},
			{
				Config: testAccAMPWorkspaceWithAliasConfig(workspaceAlias),
				Check: resource.ComposeTestCheckFunc(
					testCheckAMPWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alias", workspaceAlias),
					resource.TestCheckResourceAttr(resourceName, "alert_manager_definition", ""),
				),
			},
		},
	})
}

func TestAccPrometheusWorkspace_AMP_disappears(t *testing.T) {
	resourceName := "aws_prometheus_workspace.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAMPWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMPWorkspaceWithoutAliasConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAMPWorkspaceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfprometheus.ResourceWorkspace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckAWSAMPAlertManagerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AMP Workspace ID is set")
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
			return fmt.Errorf("Got nil alertmanager ?!")
		}

		return nil
	}
}

func testCheckAMPWorkspaceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AMP Workspace ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PrometheusConn

		req := &prometheusservice.DescribeWorkspaceInput{
			WorkspaceId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.DescribeWorkspace(req)
		if err != nil {
			return err
		}
		if describe == nil {
			return fmt.Errorf("Got nil account ?!")
		}

		return nil
	}
}

func testAccCheckAMPWorkspaceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PrometheusConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_prometheus_workspace" {
			continue
		}

		_, err := conn.DescribeWorkspace(&prometheusservice.DescribeWorkspaceInput{
			WorkspaceId: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrMessageContains(err, prometheusservice.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Prometheus WorkSpace (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccAMPWorkspaceWithAliasConfig(randName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias = %q
}
`, randName)
}

func testAccAMPWorkspaceWithoutAliasConfig() string {
	return `
resource "aws_prometheus_workspace" "test" {
}
`
}

func testAWSAMPWorkspaceConfigWithAlertManagerDefinition(name string) string {
	definition := fmt.Sprintf(`alertmanager_config: |
  route:
    receiver: '%s'
  receivers:
    - name: '%s'
      sns_configs: 
      - topic_arn: arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:My-Topic
`, name, name)
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_prometheus_workspace" "test" {
  alert_manager_definition = <<EOD
%s
EOD
}`, definition)
}
