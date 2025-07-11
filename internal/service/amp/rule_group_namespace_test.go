// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAMPRuleGroupNamespace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rgn types.RuleGroupsNamespaceDescription
	resourceName := "aws_prometheus_rule_group_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupNamespaceConfig_basic(defaultRuleGroupNamespace()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName, &rgn),
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
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName, &rgn),
					resource.TestCheckResourceAttr(resourceName, "data", anotherRuleGroupNamespace()),
				),
			},
			{
				Config: testAccRuleGroupNamespaceConfig_basic(defaultRuleGroupNamespace()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName, &rgn),
					resource.TestCheckResourceAttr(resourceName, "data", defaultRuleGroupNamespace()),
				),
			},
		},
	})
}

func TestAccAMPRuleGroupNamespace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_prometheus_rule_group_namespace.test"
	var rgn types.RuleGroupsNamespaceDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupNamespaceConfig_basic(defaultRuleGroupNamespace()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName, &rgn),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfamp.ResourceRuleGroupNamespace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAMPRuleGroupNamespace_Identity_ExistingResource(t *testing.T) {
	ctx := acctest.Context(t)
	var rgn types.RuleGroupsNamespaceDescription
	resourceName := "aws_prometheus_rule_group_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.AMPServiceID),
		CheckDestroy: testAccCheckRuleGroupNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccRuleGroupNamespaceConfig_basic(defaultRuleGroupNamespace()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName, &rgn),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.0.0",
					},
				},
				Config: testAccRuleGroupNamespaceConfig_basic(defaultRuleGroupNamespace()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName, &rgn),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: knownvalue.Null(),
					}),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccRuleGroupNamespaceConfig_basic(defaultRuleGroupNamespace()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupNamespaceExists(ctx, resourceName, &rgn),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNRegexp("aps", regexache.MustCompile(`rulegroupsnamespace/.+`)),
					}),
				},
			},
		},
	})
}

func testAccCheckRuleGroupNamespaceExists(ctx context.Context, n string, v *types.RuleGroupsNamespaceDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPClient(ctx)

		output, err := tfamp.FindRuleGroupNamespaceByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRuleGroupNamespaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPClient(ctx)

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
