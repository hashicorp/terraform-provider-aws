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

func TestAccAMPRuleGroupNamespace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_prometheus_rule_group_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(prometheusservice.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupNamespaceConfig_basic(defaultRuleGroupNamespace()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data", defaultRuleGroupNamespace()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleGroupNamespaceConfig_basic(anotherRuleGroupNamespace()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data", anotherRuleGroupNamespace()),
				),
			},
			{
				Config: testAccRuleGroupNamespaceConfig_basic(defaultRuleGroupNamespace()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data", defaultRuleGroupNamespace()),
				),
			},
		},
	})
}

func TestAccAMPRuleGroupNamespace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_prometheus_rule_group_namespace.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(prometheusservice.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupNamespaceConfig_basic(defaultRuleGroupNamespace()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfamp.ResourceRuleGroupNamespace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRuleGroupNamespaceExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Prometheus Rule Group namspace ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPConn()

		_, err := tfamp.FindRuleGroupNamespaceByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckRuleGroupNamespaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_prometheus_rule_group_namespace" {
				continue
			}

			_, err := tfamp.FindRuleGroupNamespaceByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Prometheus Rule Group Namespace %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func defaultRuleGroupNamespace() string {
	return `
groups:
  - name: test
    rules:
    - record: metric:recording_rule
      expr: avg(rate(container_cpu_usage_seconds_total[5m]))
  - name: alert-test
    rules:
    - alert: metric:alerting_rule
      expr: avg(rate(container_cpu_usage_seconds_total[5m])) > 0
      for: 2m
`
}

func anotherRuleGroupNamespace() string {
	return `
groups:
  - name: test
    rules:
    - record: metric:recording_rule
      expr: avg(rate(container_cpu_usage_seconds_total[5m]))
`
}

func testAccRuleGroupNamespaceConfig_basic(data string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {}

resource "aws_prometheus_rule_group_namespace" "test" {
  workspace_id = aws_prometheus_workspace.test.id
  name         = "rules"
  data         = %[1]q
}
`, data)
}
