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
)

func TestAccAWSAMPWorkspace_basic(t *testing.T) {
	workspaceAlias := sdkacctest.RandomWithPrefix("tf_amp_workspace")
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSAMPWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAWSAMPWorkspaceConfigWithAlias(workspaceAlias),
				Check: resource.ComposeTestCheckFunc(
					testCheckAWSAMPWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alias", workspaceAlias),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAWSAMPWorkspaceConfigWithoutAlias(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "alias", ""),
				),
			},
			{
				Config: testAWSAMPWorkspaceConfigWithAlias(workspaceAlias),
				Check: resource.ComposeTestCheckFunc(
					testCheckAWSAMPWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alias", workspaceAlias),
				),
			},
		},
	})
}

func TestAccAWSAMPWorkspace_disappears(t *testing.T) {
	resourceName := "aws_prometheus_workspace.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSAMPWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAWSAMPWorkspaceConfigWithoutAlias(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAWSAMPWorkspaceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceWorkspace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckAWSAMPWorkspaceExists(n string) resource.TestCheckFunc {
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

func testAccCheckAWSAMPWorkspaceDestroy(s *terraform.State) error {
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

func testAWSAMPWorkspaceConfigWithAlias(randName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias = %q
}
`, randName)
}

func testAWSAMPWorkspaceConfigWithoutAlias() string {
	return `
resource "aws_prometheus_workspace" "test" {
}
`
}
