// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package amp_test

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the service/amp/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance tests access AWS and cost money to run.
func TestAccAMPAnomalyDetector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_anomaly_detector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
			testAccPreCheckAnomalyDetector(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyDetectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					// TIP: If the ARN can be partially or completely determined by the parameters passed, e.g. it contains the
					// value of `rName`, either include the values in the regex or check for an exact match using `acctest.CheckResourceAttrRegionalARN`
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "aps", regexache.MustCompile(`anomalydetector:.+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccAMPAnomalyDetector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_anomaly_detector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
			testAccPreCheckAnomalyDetector(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyDetectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceAnomalyDetector = newAnomalyDetectorResource
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfamp.ResourceAnomalyDetector, resourceName),
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

func testAccCheckAnomalyDetectorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AMPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_prometheus_anomaly_detector" {
				continue
			}

			_, err := tfamp.FindAnomalyDetectorByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["workspace_id"])
			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return create.Error(names.AMP, create.ErrActionCheckingDestroyed, tfamp.ResNameAnomalyDetector, rs.Primary.ID, err)
			}

			return create.Error(names.AMP, create.ErrActionCheckingDestroyed, tfamp.ResNameAnomalyDetector, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAnomalyDetectorExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AMP, create.ErrActionCheckingExistence, tfamp.ResNameAnomalyDetector, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.AMP, create.ErrActionCheckingExistence, tfamp.ResNameAnomalyDetector, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).AMPClient(ctx)

		_, err := tfamp.FindAnomalyDetectorByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["workspace_id"])
		return create.Error(names.AMP, create.ErrActionCheckingExistence, tfamp.ResNameAnomalyDetector, rs.Primary.ID, err)
	}
}

func testAccPreCheckAnomalyDetector(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).AMPClient(ctx)

	input := &amp.ListAnomalyDetectorsInput{}

	_, err := conn.ListAnomalyDetectors(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAnomalyDetectorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {}

resource "aws_prometheus_anomaly_detector" "test" {
  alias = %[1]q
  workspace_id = aws_prometheus_workspace.test.id

  configuration {
	random_cut_forest {
	  query = avg(up)
	}
  }
}
`, rName)
}

