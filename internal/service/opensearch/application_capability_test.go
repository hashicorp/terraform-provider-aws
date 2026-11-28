// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchApplicationCapability_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var capability opensearch.GetCapabilityOutput
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")[:30]
	resourceName := "aws_opensearch_application_capability.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationCapabilityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationCapabilityConfig_ai(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationCapabilityExists(ctx, t, resourceName, &capability),
					resource.TestCheckResourceAttr(resourceName, "capability_name", "ai-capability"),
					resource.TestCheckResourceAttr(resourceName, "ai_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, "aws_opensearch_application.test", names.AttrID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccApplicationCapabilityImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrApplicationID,
			},
		},
	})
}

func TestAccOpenSearchApplicationCapability_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var capability opensearch.GetCapabilityOutput
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")[:30]
	resourceName := "aws_opensearch_application_capability.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationCapabilityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationCapabilityConfig_ai(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationCapabilityExists(ctx, t, resourceName, &capability),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfopensearch.ResourceApplicationCapability, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckApplicationCapabilityDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_application_capability" {
				continue
			}

			_, err := tfopensearch.FindApplicationCapabilityByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["capability_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Application Capability %s/%s still exists", rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["capability_name"])
		}

		return nil
	}
}

func testAccCheckApplicationCapabilityExists(ctx context.Context, t *testing.T, n string, v *opensearch.GetCapabilityOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

		output, err := tfopensearch.FindApplicationCapabilityByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["capability_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccApplicationCapabilityImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}
		return fmt.Sprintf("%s,%s", rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["capability_name"]), nil
	}
}

func testAccApplicationCapabilityConfig_ai(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_application" "test" {
  name = %[1]q
}

resource "aws_opensearch_application_capability" "test" {
  application_id  = aws_opensearch_application.test.id
  capability_name = "ai-capability"

  ai_config {}
}
`, rName)
}
