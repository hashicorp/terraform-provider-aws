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
	"strconv"
	"testing"
	"time"

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
func testAccInspector2Filter_basic(t *testing.T) {
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
				Config: testAccFilterConfig_basic(rName, action_1, description_1, reason_1, comparison_1, account_id_1),
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
		},
	})
}

func testAccInspector2Filter_update(t *testing.T) {
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
				Config: testAccFilterConfig_basic(rName, action_1, description_1, reason_1, comparison_1, account_id_1),
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
				Config: testAccFilterConfig_basic(rName, action_2, description_2, reason_2, comparison_2, account_id_2),
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

func testAccInspector2Filter_stringFilters(t *testing.T) {
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
	value_1 := "TestValue1"
	reason_1 := "TestReason_1"

	action_2 := string(awstypes.FilterActionSuppress)
	description_2 := "TestDescription_2"
	comparison_2 := string(awstypes.StringComparisonNotEquals)
	value_2 := "TestValue_2"
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
				Config: testAccFilterConfig_stringFilters(rName, action_1, description_1, reason_1, comparison_1, value_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, "action", action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.code_vulnerability_detector_name.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.code_vulnerability_detector_name.*", map[string]string{
						"comparison": comparison_1,
						"value":      value_1,
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
				Config: testAccFilterConfig_stringFilters(rName, action_2, description_2, reason_2, comparison_2, value_2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, "action", action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.code_vulnerability_detector_name.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.code_vulnerability_detector_name.*", map[string]string{
						"comparison": comparison_2,
						"value":      value_2,
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

func testAccInspector2Filter_numberFilters(t *testing.T) {
	ctx := acctest.Context(t)
	// fmt.Printf("Account is: " + acctest.AccountID(ctx))
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	lower_inclusive_value_1 := "0"
	upper_inclusive_value_1 := "100"
	reason_1 := "TestReason_1"

	action_2 := string(awstypes.FilterActionSuppress)
	description_2 := "TestDescription_2"
	lower_inclusive_value_2 := "1"
	upper_inclusive_value_2 := "101"
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
				Config: testAccFilterConfig_numberFilters(rName, action_1, description_1, reason_1, lower_inclusive_value_1, upper_inclusive_value_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, "action", action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.epss_score.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.epss_score.*", map[string]string{
						"lower_inclusive": lower_inclusive_value_1,
						"upper_inclusive": upper_inclusive_value_1,
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
				Config: testAccFilterConfig_numberFilters(rName, action_2, description_2, reason_2, lower_inclusive_value_2, upper_inclusive_value_2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, "action", action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.epss_score.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.epss_score.*", map[string]string{
						"lower_inclusive": lower_inclusive_value_2,
						"upper_inclusive": upper_inclusive_value_2,
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

func testAccInspector2Filter_dateFilters(t *testing.T) {
	ctx := acctest.Context(t)
	// fmt.Printf("Account is: " + acctest.AccountID(ctx))
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	start_inclusive_value_1 := time.Now().Format(time.RFC3339)
	end_inclusive_value_1 := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	reason_1 := "TestReason_1"

	action_2 := string(awstypes.FilterActionSuppress)
	description_2 := "TestDescription_2"
	start_inclusive_value_2 := time.Now().Add(6 * time.Minute).Format(time.RFC3339)
	end_inclusive_value_2 := time.Now().Add(10 * time.Minute).Format(time.RFC3339)
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
				Config: testAccFilterConfig_dateFilters(rName, action_1, description_1, reason_1, start_inclusive_value_1, end_inclusive_value_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, "action", action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.ecr_image_pushed_at.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.ecr_image_pushed_at.*", map[string]string{
						"start_inclusive": start_inclusive_value_1,
						"end_inclusive":   end_inclusive_value_1,
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
				Config: testAccFilterConfig_dateFilters(rName, action_2, description_2, reason_2, start_inclusive_value_2, end_inclusive_value_2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, "action", action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.ecr_image_pushed_at.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.ecr_image_pushed_at.*", map[string]string{
						"start_inclusive": start_inclusive_value_2,
						"end_inclusive":   end_inclusive_value_2,
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

func testAccInspector2Filter_mapFilters(t *testing.T) {
	ctx := acctest.Context(t)
	// fmt.Printf("Account is: " + acctest.AccountID(ctx))
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	comparison := string(awstypes.MapComparisonEquals)

	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	key_1 := "some_key_1"
	value_1 := "some_value_1"
	reason_1 := "TestReason_1"

	action_2 := string(awstypes.FilterActionSuppress)
	description_2 := "TestDescription_2"
	key_2 := "some_key_2"
	value_2 := "some_value_2"
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
				Config: testAccFilterConfig_mapFilters(rName, action_1, description_1, reason_1, comparison, key_1, value_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, "action", action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.resource_tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.resource_tags.*", map[string]string{
						"comparison": comparison,
						"key":        key_1,
						"value":      value_1,
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
				Config: testAccFilterConfig_mapFilters(rName, action_2, description_2, reason_2, comparison, key_2, value_2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, "action", action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.resource_tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.resource_tags.*", map[string]string{
						"comparison": comparison,
						"key":        key_2,
						"value":      value_2,
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

func testAccInspector2Filter_portRangeFilters(t *testing.T) {
	ctx := acctest.Context(t)
	// fmt.Printf("Account is: " + acctest.AccountID(ctx))
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	begin_inclusive_value_1 := 0
	end_inclusive_value_1 := 100
	reason_1 := "TestReason_1"

	action_2 := string(awstypes.FilterActionSuppress)
	description_2 := "TestDescription_2"
	begin_inclusive_value_2 := 101
	end_inclusive_value_2 := 200
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
				Config: testAccFilterConfig_portRangeFilters(rName, action_1, description_1, reason_1, begin_inclusive_value_1, end_inclusive_value_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, "action", action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.0.begin_inclusive", strconv.Itoa(begin_inclusive_value_1)),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.0.end_inclusive", strconv.Itoa(end_inclusive_value_1)),
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
				Config: testAccFilterConfig_portRangeFilters(rName, action_2, description_2, reason_2, begin_inclusive_value_2, end_inclusive_value_2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, "action", action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.0.begin_inclusive", strconv.Itoa(begin_inclusive_value_2)),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.0.end_inclusive", strconv.Itoa(end_inclusive_value_2)),
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

func testAccInspector2Filter_packageFilters(t *testing.T) {
	ctx := acctest.Context(t)
	// fmt.Printf("Account is: " + acctest.AccountID(ctx))
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	comparison := string(awstypes.MapComparisonEquals)

	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	reason_1 := "TestReason_1"

	action_2 := string(awstypes.FilterActionSuppress)
	description_2 := "TestDescription_2"
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
				Config: testAccFilterConfig_packageFilter_1(rName, action_1, description_1, reason_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, "action", action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.architecture.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.architecture.*", map[string]string{
						"comparison": comparison,
						"value":      "arch_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.epoch.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.epoch.*", map[string]string{
						"lower_inclusive": "10",
						"upper_inclusive": "20",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.file_path.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.file_path.*", map[string]string{
						"comparison": comparison,
						"value":      "file_path_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.name.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.name.*", map[string]string{
						"comparison": comparison,
						"value":      "name_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.release.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.release.*", map[string]string{
						"comparison": comparison,
						"value":      "release_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.source_lambda_layer_arn.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.source_lambda_layer_arn.*", map[string]string{
						"comparison": comparison,
						"value":      "source_lambda_layer_arn_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.source_layer_hash.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.source_layer_hash.*", map[string]string{
						"comparison": comparison,
						"value":      "source_layer_hash_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.version.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.version.*", map[string]string{
						"comparison": comparison,
						"value":      "version_1",
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
				Config: testAccFilterConfig_packageFilter_2(rName, action_2, description_2, reason_2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "description", description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, "action", action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.architecture.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.architecture.*", map[string]string{
						"comparison": comparison,
						"value":      "arch_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.epoch.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.epoch.*", map[string]string{
						"lower_inclusive": "21",
						"upper_inclusive": "30",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.file_path.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.file_path.*", map[string]string{
						"comparison": comparison,
						"value":      "file_path_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.name.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.name.*", map[string]string{
						"comparison": comparison,
						"value":      "name_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.release.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.release.*", map[string]string{
						"comparison": comparison,
						"value":      "release_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.source_lambda_layer_arn.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.source_lambda_layer_arn.*", map[string]string{
						"comparison": comparison,
						"value":      "source_lambda_layer_arn_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.source_layer_hash.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.source_layer_hash.*", map[string]string{
						"comparison": comparison,
						"value":      "source_layer_hash_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.version.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.version.*", map[string]string{
						"comparison": comparison,
						"value":      "version_2",
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

func testAccInspector2Filter_disappears(t *testing.T) {
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
				Config: testAccFilterConfig_basic(rName, action_1, description_1, reason_1, comparison_1, account_id_1),
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

func testAccFilterConfig_basic(rName, action, description, reason, comparison, value string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name             	= %[1]q
  action			= %[2]q
  description 		= %[3]q
  reason 			= %[4]q
  filter_criteria {
    aws_account_id {
      comparison = %[5]q
      value      = %[6]q
    }
  }
}
`, rName, action, description, reason, comparison, value)
}

func testAccFilterConfig_stringFilters(rName, action, description, reason, comparison, value string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name             	= %[1]q
  action			= %[2]q
  description 		= %[3]q
  reason 			= %[4]q
  filter_criteria {
    code_vulnerability_detector_name {
      comparison = %[5]q
      value      = %[6]q
    }
  }
}
`, rName, action, description, reason, comparison, value)
}

func testAccFilterConfig_numberFilters(rName, action, description, reason, lower_inclusive_value, upper_inclusive_value string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name             	= %[1]q
  action			= %[2]q
  description 		= %[3]q
  reason 			= %[4]q
  filter_criteria {
    epss_score {
      lower_inclusive = %[5]q
      upper_inclusive      = %[6]q
    }
  }
}
`, rName, action, description, reason, lower_inclusive_value, upper_inclusive_value)
}

func testAccFilterConfig_dateFilters(rName, action, description, reason, start_inclusive_value, end_inclusive_value string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name             	= %[1]q
  action			= %[2]q
  description 		= %[3]q
  reason 			= %[4]q
  filter_criteria {
    ecr_image_pushed_at {
      start_inclusive    = %[5]q
      end_inclusive      = %[6]q
    }
  }
}
`, rName, action, description, reason, start_inclusive_value, end_inclusive_value)
}

func testAccFilterConfig_mapFilters(rName, action, description, reason, comparison, key, value string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name             	= %[1]q
  action			= %[2]q
  description 		= %[3]q
  reason 			= %[4]q
  filter_criteria {
    resource_tags {
	  comparison = %[5]q
      key = %[6]q
      value = %[7]q
    }
  }
}
`, rName, action, description, reason, comparison, key, value)
}

func testAccFilterConfig_portRangeFilters(rName, action, description, reason string, begin_inclusive_value, end_inclusive_value int) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name             	= %[1]q
  action			= %[2]q
  description 		= %[3]q
  reason 			= %[4]q
  filter_criteria {
    port_range {
      begin_inclusive    = %[5]d
      end_inclusive      = %[6]d
    }
  }
}
`, rName, action, description, reason, begin_inclusive_value, end_inclusive_value)
}

func testAccFilterConfig_packageFilter_1(rName, action, description, reason string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name             	= %[1]q
  action			= %[2]q
  description 		= %[3]q
  reason 			= %[4]q
  filter_criteria {
  	vulnerable_packages{
		architecture {
			comparison = "EQUALS"
			value      = "arch_1"
		}
		epoch {
			lower_inclusive = "10"
			upper_inclusive = "20"
		}
		file_path {
			comparison = "EQUALS"
			value      = "file_path_1"
		}
		name {
			comparison = "EQUALS"
			value      = "name_1"
		}
		release {
			comparison = "EQUALS"
			value      = "release_1"
		}
		source_lambda_layer_arn {
			comparison = "EQUALS"
			value      = "source_lambda_layer_arn_1"
		}
		source_layer_hash {
			comparison = "EQUALS"
			value      = "source_layer_hash_1"
		}
		version {
			comparison = "EQUALS"
			value      = "version_1"
		}
	}
  }
}
`, rName, action, description, reason)
}

func testAccFilterConfig_packageFilter_2(rName, action, description, reason string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name             	= %[1]q
  action			= %[2]q
  description 		= %[3]q
  reason 			= %[4]q
	filter_criteria {
		vulnerable_packages{
			architecture {
				comparison = "EQUALS"
				value      = "arch_2"
			}
			epoch {
				lower_inclusive = "21"
				upper_inclusive = "30"
			}
			file_path {
				comparison = "EQUALS"
				value      = "file_path_2"
			}
			name {
				comparison = "EQUALS"
				value      = "name_2"
			}
			release {
				comparison = "EQUALS"
				value      = "release_2"
			}
			source_lambda_layer_arn {
				comparison = "EQUALS"
				value      = "source_lambda_layer_arn_2"
			}
			source_layer_hash {
				comparison = "EQUALS"
				value      = "source_layer_hash_2"
			}
			version {
				comparison = "EQUALS"
				value      = "version_2"
			}
		}
  	}
}
`, rName, action, description, reason)
}
