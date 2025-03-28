// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2_test

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
	// using the services/inspector2/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	// "github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
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
// Acceptance test access AWS and cost money to run.
func TestAccInspector2Filter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// fmt.Printf("Account is: " + acctest.AccountID(ctx))
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	comparison_1 := string(awstypes.StringComparisonEquals)
	account_id_1 := "111222333444"
	reason_1 := "TestReason_1"

	action_2 := string(awstypes.FilterActionSuppress)
	description_2 := "TestDescription_2"
	comparison_2 := string(awstypes.StringComparisonNotEquals)
	account_id_2 := "444333222111"
	reason_2 := "TestReason_2"

	var filter awstypes.Filter
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_basic(rName, action_1, description_1, comparison_1, account_id_1, reason_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, "action", action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.aws_account_id.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.aws_account_id.*", map[string]string{
						"comparison": comparison_1,
						"value":      account_id_1,
					}),
					// TIP: If the ARN can be partially or completely determined by the parameters passed, e.g. it contains the
					// value of `rName`, either include the values in the regex or check for an exact match using `acctest.CheckResourceAttrRegionalARN`
					// acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`filter:.+$`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/.+/filter/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccFilterImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccFilterConfig_basic(rName, action_2, description_2, comparison_2, account_id_2, reason_2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, "action", action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.aws_account_id.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.aws_account_id.*", map[string]string{
						"comparison": comparison_2,
						"value":      account_id_2,
					}),
					// TIP: If the ARN can be partially or completely determined by the parameters passed, e.g. it contains the
					// value of `rName`, either include the values in the regex or check for an exact match using `acctest.CheckResourceAttrRegionalARN`
					// acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`filter:.+$`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/.+/filter/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccFilterImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccFilterImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrARN], nil
	}
}

func TestAccInspector2Filter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	comparison_1 := string(awstypes.StringComparisonEquals)
	account_id_1 := "111222333444"
	reason_1 := "TestReason_1"

	var filter awstypes.Filter
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_basic(rName, action_1, description_1, comparison_1, account_id_1, reason_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceFilter = newResourceFilter
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfinspector2.ResourceFilter, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFilterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_filter" {
				continue
			}

			// TIP: ==== FINDERS ====
			// The find function should be exported. Since it won't be used outside of the package, it can be exported
			// in the `exports_test.go` file.
			_, err := tfinspector2.FindFilterByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameFilter, rs.Primary.Attributes[names.AttrARN], err)
			}

			return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameFilter, rs.Primary.Attributes[names.AttrARN], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFilterExists(ctx context.Context, name string, filter *awstypes.Filter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameFilter, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrARN] == "" {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameFilter, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

		resp, err := tfinspector2.FindFilterByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameFilter, rs.Primary.Attributes[names.AttrARN], err)
		}

		*filter = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

	input := &inspector2.ListFiltersInput{}

	_, err := conn.ListFilters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckFilterNotRecreated(before, after *awstypes.Filter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if beforeArn, afterArn := aws.ToString(before.Arn), aws.ToString(after.Arn); beforeArn != afterArn {
			return create.Error(names.Inspector2, create.ErrActionCheckingNotRecreated, tfinspector2.ResNameFilter, beforeArn, errors.New("recreated"))
		}

		return nil
	}
}

func testAccFilterConfig_basic(rName, action, description, comparison, account_id, reason string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name             	= %[1]q
  action			= %[2]q
  description 		= %[3]q
  filter_criteria {
    aws_account_id {
      comparison = %[4]q
      value      = %[5]q
    }
  }
  reason 			= %[6]q
}
`, rName, action, description, comparison, account_id, reason)
}
