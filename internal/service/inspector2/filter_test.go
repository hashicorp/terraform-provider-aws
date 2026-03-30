// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package inspector2_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspector2Filter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	comparison_1 := string(awstypes.StringComparisonEquals)
	account_id_1 := "111222333444"
	reason_1 := "TestReason_1"
	var filter awstypes.Filter
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_basic(rName, action_1, description_1, reason_1, comparison_1, account_id_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.aws_account_id.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.aws_account_id.*", map[string]string{
						"comparison":    comparison_1,
						names.AttrValue: account_id_1,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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

func TestAccInspector2Filter_update(t *testing.T) {
	ctx := acctest.Context(t)
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_basic(rName, action_1, description_1, reason_1, comparison_1, account_id_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.aws_account_id.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.aws_account_id.*", map[string]string{
						"comparison":    comparison_1,
						names.AttrValue: account_id_1,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.aws_account_id.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.aws_account_id.*", map[string]string{
						"comparison":    comparison_2,
						names.AttrValue: account_id_2,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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

func TestAccInspector2Filter_stringFilters(t *testing.T) {
	ctx := acctest.Context(t)
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_stringFilters(rName, action_1, description_1, reason_1, comparison_1, value_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.code_repository_project_name.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.code_repository_project_name.*", map[string]string{
						"comparison":    comparison_1,
						names.AttrValue: value_1,
					}),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.code_repository_provider_type.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.code_repository_provider_type.*", map[string]string{
						"comparison":    comparison_1,
						names.AttrValue: value_1,
					}),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.code_vulnerability_detector_name.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.code_vulnerability_detector_name.*", map[string]string{
						"comparison":    comparison_1,
						names.AttrValue: value_1,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.code_vulnerability_detector_name.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.code_repository_project_name.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.code_repository_project_name.*", map[string]string{
						"comparison":    comparison_2,
						names.AttrValue: value_2,
					}),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.code_repository_provider_type.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.code_repository_provider_type.*", map[string]string{
						"comparison":    comparison_2,
						names.AttrValue: value_2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.code_vulnerability_detector_name.*", map[string]string{
						"comparison":    comparison_2,
						names.AttrValue: value_2,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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

func TestAccInspector2Filter_numberFilters(t *testing.T) {
	ctx := acctest.Context(t)
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_numberFilters(rName, action_1, description_1, reason_1, lower_inclusive_value_1, upper_inclusive_value_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.ecr_image_in_use_count.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.ecr_image_in_use_count.*", map[string]string{
						"lower_inclusive": lower_inclusive_value_1,
						"upper_inclusive": upper_inclusive_value_1,
					}),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.epss_score.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.epss_score.*", map[string]string{
						"lower_inclusive": lower_inclusive_value_1,
						"upper_inclusive": upper_inclusive_value_1,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.ecr_image_in_use_count.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.ecr_image_in_use_count.*", map[string]string{
						"lower_inclusive": lower_inclusive_value_2,
						"upper_inclusive": upper_inclusive_value_2,
					}),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.epss_score.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.epss_score.*", map[string]string{
						"lower_inclusive": lower_inclusive_value_2,
						"upper_inclusive": upper_inclusive_value_2,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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

func TestAccInspector2Filter_dateFilters(t *testing.T) {
	ctx := acctest.Context(t)
	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	start_inclusive_value_1 := time.Now().In(time.UTC).Format(time.RFC3339)
	end_inclusive_value_1 := time.Now().In(time.UTC).Add(5 * time.Minute).Format(time.RFC3339)
	reason_1 := "TestReason_1"
	action_2 := string(awstypes.FilterActionSuppress)
	description_2 := "TestDescription_2"
	start_inclusive_value_2 := time.Now().In(time.UTC).Add(6 * time.Minute).Format(time.RFC3339)
	end_inclusive_value_2 := time.Now().In(time.UTC).Add(10 * time.Minute).Format(time.RFC3339)
	reason_2 := "TestReason_2"
	var filter awstypes.Filter
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_dateFilters(rName, action_1, description_1, reason_1, start_inclusive_value_1, end_inclusive_value_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.ecr_image_last_in_use_at.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.ecr_image_last_in_use_at.*", map[string]string{
						"start_inclusive": start_inclusive_value_1,
						"end_inclusive":   end_inclusive_value_1,
					}),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.ecr_image_pushed_at.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.ecr_image_pushed_at.*", map[string]string{
						"start_inclusive": start_inclusive_value_1,
						"end_inclusive":   end_inclusive_value_1,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.ecr_image_last_in_use_at.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.ecr_image_last_in_use_at.*", map[string]string{
						"start_inclusive": start_inclusive_value_2,
						"end_inclusive":   end_inclusive_value_2,
					}),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.ecr_image_pushed_at.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.ecr_image_pushed_at.*", map[string]string{
						"start_inclusive": start_inclusive_value_2,
						"end_inclusive":   end_inclusive_value_2,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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

func TestAccInspector2Filter_mapFilters(t *testing.T) {
	ctx := acctest.Context(t)
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_mapFilters(rName, action_1, description_1, reason_1, comparison, key_1, value_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.resource_tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.resource_tags.*", map[string]string{
						"comparison":    comparison,
						names.AttrKey:   key_1,
						names.AttrValue: value_1,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.resource_tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.resource_tags.*", map[string]string{
						"comparison":    comparison,
						names.AttrKey:   key_2,
						names.AttrValue: value_2,
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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

func TestAccInspector2Filter_portRangeFilters(t *testing.T) {
	ctx := acctest.Context(t)
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_portRangeFilters(rName, action_1, description_1, reason_1, begin_inclusive_value_1, end_inclusive_value_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.0.begin_inclusive", strconv.Itoa(begin_inclusive_value_1)),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.0.end_inclusive", strconv.Itoa(end_inclusive_value_1)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.0.begin_inclusive", strconv.Itoa(begin_inclusive_value_2)),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.port_range.0.end_inclusive", strconv.Itoa(end_inclusive_value_2)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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

func TestAccInspector2Filter_packageFilters(t *testing.T) {
	ctx := acctest.Context(t)
	comparison := string(awstypes.MapComparisonEquals)
	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	reason_1 := "TestReason_1"
	action_2 := string(awstypes.FilterActionSuppress)
	description_2 := "TestDescription_2"
	reason_2 := "TestReason_2"
	var filter awstypes.Filter
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_packageFilter_1(rName, action_1, description_1, reason_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_1),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_1),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.architecture.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.architecture.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "arch_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.epoch.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.epoch.*", map[string]string{
						"lower_inclusive": "10",
						"upper_inclusive": "20",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.file_path.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.file_path.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "file_path_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.name.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.name.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "name_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.release.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.release.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "release_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.source_lambda_layer_arn.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.source_lambda_layer_arn.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "source_lambda_layer_arn_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.source_layer_hash.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.source_layer_hash.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "source_layer_hash_1",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.version.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.version.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "version_1",
					}),

					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description_2),
					resource.TestCheckResourceAttr(resourceName, "reason", reason_2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, action_2),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.architecture.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.architecture.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "arch_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.epoch.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.epoch.*", map[string]string{
						"lower_inclusive": "21",
						"upper_inclusive": "30",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.file_path.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.file_path.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "file_path_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.name.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.name.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "name_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.release.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.release.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "release_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.source_lambda_layer_arn.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.source_lambda_layer_arn.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "source_lambda_layer_arn_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.source_layer_hash.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.source_layer_hash.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "source_layer_hash_2",
					}),

					resource.TestCheckResourceAttr(resourceName, "filter_criteria.0.vulnerable_packages.0.version.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filter_criteria.0.vulnerable_packages.0.version.*", map[string]string{
						"comparison":    comparison,
						names.AttrValue: "version_2",
					}),

					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector2", regexache.MustCompile(`owner/\d{12}/filter/.+$`)),
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

func TestAccInspector2Filter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	action_1 := string(awstypes.FilterActionNone)
	description_1 := "TestDescription_1"
	comparison_1 := string(awstypes.StringComparisonEquals)
	account_id_1 := "111222333444"
	reason_1 := "TestReason_1"
	var filter awstypes.Filter
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_inspector2_filter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_basic(rName, action_1, description_1, reason_1, comparison_1, account_id_1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &filter),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfinspector2.ResourceFilter, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFilterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Inspector2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_filter" {
				continue
			}

			_, err := tfinspector2.FindFilterByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Inspector2 Filter %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckFilterExists(ctx context.Context, t *testing.T, n string, v *awstypes.Filter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).Inspector2Client(ctx)

		output, err := tfinspector2.FindFilterByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
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

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).Inspector2Client(ctx)

	input := &inspector2.ListFiltersInput{}

	_, err := conn.ListFilters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccFilterConfig_basic(rName, action, description, reason, comparison, value string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name        = %[1]q
  action      = %[2]q
  description = %[3]q
  reason      = %[4]q
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
  name        = %[1]q
  action      = %[2]q
  description = %[3]q
  reason      = %[4]q
  filter_criteria {
    code_repository_project_name {
      comparison = %[5]q
      value      = %[6]q
    }
    code_repository_provider_type {
      comparison = %[5]q
      value      = %[6]q
    }
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
  name        = %[1]q
  action      = %[2]q
  description = %[3]q
  reason      = %[4]q
  filter_criteria {
    ecr_image_in_use_count {
      lower_inclusive = %[5]q
      upper_inclusive = %[6]q
    }
    epss_score {
      lower_inclusive = %[5]q
      upper_inclusive = %[6]q
    }
  }
}
`, rName, action, description, reason, lower_inclusive_value, upper_inclusive_value)
}

func testAccFilterConfig_dateFilters(rName, action, description, reason, start_inclusive_value, end_inclusive_value string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name        = %[1]q
  action      = %[2]q
  description = %[3]q
  reason      = %[4]q
  filter_criteria {
    ecr_image_last_in_use_at {
      start_inclusive = %[5]q
      end_inclusive   = %[6]q
    }
    ecr_image_pushed_at {
      start_inclusive = %[5]q
      end_inclusive   = %[6]q
    }
  }
}
`, rName, action, description, reason, start_inclusive_value, end_inclusive_value)
}

func testAccFilterConfig_mapFilters(rName, action, description, reason, comparison, key, value string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name        = %[1]q
  action      = %[2]q
  description = %[3]q
  reason      = %[4]q
  filter_criteria {
    resource_tags {
      comparison = %[5]q
      key        = %[6]q
      value      = %[7]q
    }
  }
}
`, rName, action, description, reason, comparison, key, value)
}

func testAccFilterConfig_portRangeFilters(rName, action, description, reason string, begin_inclusive_value, end_inclusive_value int) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name        = %[1]q
  action      = %[2]q
  description = %[3]q
  reason      = %[4]q
  filter_criteria {
    port_range {
      begin_inclusive = %[5]d
      end_inclusive   = %[6]d
    }
  }
}
`, rName, action, description, reason, begin_inclusive_value, end_inclusive_value)
}

func testAccFilterConfig_packageFilter_1(rName, action, description, reason string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_filter" "test" {
  name        = %[1]q
  action      = %[2]q
  description = %[3]q
  reason      = %[4]q
  filter_criteria {
    vulnerable_packages {
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
  name        = %[1]q
  action      = %[2]q
  description = %[3]q
  reason      = %[4]q
  filter_criteria {
    vulnerable_packages {
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
