// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Acceptance test access AWS and cost money to run.
func TestAccSSMParameterVersionLabels_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var parameterversionlabels []string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_parameter_version_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterVersionLabelsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterVersionLabelsConfig_basic(rName, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckParameterVersionLabelsExists(ctx, resourceName, &parameterversionlabels),
					resource.TestCheckResourceAttr(resourceName, "labels.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.0", "label1"),
					resource.TestCheckResourceAttr(resourceName, "labels.1", "label2"),
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

func TestAccSSMParameterVersionLabels_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var parameterversionlabels []string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_parameter_version_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterVersionLabelsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterVersionLabelsConfig_basic(rName, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckParameterVersionLabelsExists(ctx, resourceName, &parameterversionlabels),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourceParameterVersionLabels(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckParameterVersionLabelsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_parameter_version_labels" {
				continue
			}

			name, version, err := tfssm.ParameterVersionLabelsParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}
			_, err = tfssm.FindParameterVersionLabels(ctx, conn, name, version)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Parameter Version Labels %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckParameterVersionLabelsExists(ctx context.Context, name string, parameterversionlabels *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		name, version, err := tfssm.ParameterVersionLabelsParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}
		resp, err := tfssm.FindParameterVersionLabels(ctx, conn, name, version)

		if err != nil {
			return err
		}

		*parameterversionlabels = append(*parameterversionlabels, resp...)

		return nil
	}
}

func testAccParameterVersionLabelsConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "String"
  value = "test value %[2]s"
}

resource "aws_ssm_parameter_version_labels" "test" {
  name    = aws_ssm_parameter.test.name
  version = %[2]q == "" ? 0 : tonumber(%[2]q)
  labels  = ["label1", "label2"]
}
`, rName, version)
}
