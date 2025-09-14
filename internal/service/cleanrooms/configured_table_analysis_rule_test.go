// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cleanrooms_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcleanrooms "github.com/hashicorp/terraform-provider-aws/internal/service/cleanrooms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCleanRoomsConfiguredTableAnalysisRule_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "LIST"
	var joinColumns = "[\"my_column_1\", \"my_column_2\"]"
	var allowedJoinOperators = "[\"AND\"]"
	var listColumns = "[\"my_column_3\", \"my_column_4\"]"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType, joinColumns,
					allowedJoinOperators, listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.1", "my_column_4"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationAnalysisRuleType(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_customAnalysisRuleType(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "CUSTOM"
	var customAnalysis = "[\"ANY_QUERY\"]"
	var analysisProviders = "[\"123456789012\"]"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_custom(analysisRuleType, customAnalysis, analysisProviders),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_customEnableDifferentialPrivacy(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "CUSTOM"
	var customAnalysis = "[\"ANY_QUERY\"]"
	var analysisProviders = "[\"123456789012\"]"
	var columns = "my_column_1"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_customDifferentialPrivacy(analysisRuleType, customAnalysis, analysisProviders, columns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.differential_privacy.0.columns.0.name", "my_column_1"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_listOperatorDisappears(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "LIST"
	var joinColumns = "[\"my_column_1\", \"my_column_2\"]"
	var allowedJoinOperators = "[\"AND\"]"
	var listColumns = "[\"my_column_3\", \"my_column_4\"]"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType, joinColumns,
					allowedJoinOperators, listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider,
						tfcleanrooms.ResourceConfiguredTableAnalysisRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationOperatorDisappears(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider,
						tfcleanrooms.ResourceConfiguredTableAnalysisRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_customOperatorDisappears(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "CUSTOM"
	var customAnalysis = "[\"ANY_QUERY\"]"
	var analysisProviders = "[\"123456789012\"]"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_custom(analysisRuleType, customAnalysis, analysisProviders),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider,
						tfcleanrooms.ResourceConfiguredTableAnalysisRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_updateListToAggregation(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "LIST"
	var joinColumns = "[\"my_column_1\", \"my_column_2\"]"
	var allowedJoinOperators = "[\"AND\"]"
	var listColumns = "[\"my_column_3\", \"my_column_4\"]"

	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"

	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType, joinColumns,
					allowedJoinOperators, listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.1", "my_column_4"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation("AGGREGATION", aggregateColumns,
					aggregateFunction, joinAggregateColumns, "[\"AND\"]", dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, "AGGREGATION",
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_updateListToCustom(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "LIST"
	var joinColumns = "[\"my_column_1\", \"my_column_2\"]"
	var allowedJoinOperators = "[\"AND\"]"
	var listColumns = "[\"my_column_3\", \"my_column_4\"]"

	var customAnalysis = "[\"ANY_QUERY\"]"
	var analysisProviders = "[\"123456789012\"]"

	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType, joinColumns,
					allowedJoinOperators, listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.1", "my_column_4"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_custom("CUSTOM", customAnalysis, analysisProviders),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, "CUSTOM", &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_updateAggregationToList(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"

	var joinColumns = "[\"my_column_1\", \"my_column_2\"]"
	var listColumns = "[\"my_column_3\", \"my_column_4\"]"

	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic("LIST", joinColumns,
					allowedJoinOperators, listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, "LIST", &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.1", "my_column_4"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_updateAggregationToCustom(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"

	var customAnalysis = "[\"ANY_QUERY\"]"
	var analysisProviders = "[\"123456789012\"]"

	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_custom("CUSTOM", customAnalysis, analysisProviders),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, "CUSTOM", &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_updateCustomToList(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "CUSTOM"
	var customAnalysis = "[\"ANY_QUERY\"]"
	var analysisProviders = "[\"123456789012\"]"

	var joinColumns = "[\"my_column_1\", \"my_column_2\"]"
	var allowedJoinOperators = "[\"AND\"]"
	var listColumns = "[\"my_column_3\", \"my_column_4\"]"

	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_custom(analysisRuleType, customAnalysis, analysisProviders),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic("LIST", joinColumns,
					allowedJoinOperators, listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, "LIST", &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.1", "my_column_4"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_updateCustomToAggregation(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "CUSTOM"
	var customAnalysis = "[\"ANY_QUERY\"]"
	var analysisProviders = "[\"123456789012\"]"

	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"

	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_custom(analysisRuleType, customAnalysis, analysisProviders),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation("AGGREGATION", aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, "AGGREGATION",
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_updateJoinColumns(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "LIST"
	var joinColumns = "[\"my_column_1\", \"my_column_2\"]"
	var allowedJoinOperators = "[\"AND\"]"
	var listColumns = "[\"my_column_3\", \"my_column_4\"]"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType, joinColumns,
					allowedJoinOperators, listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.1", "my_column_4"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType, "[\"my_column_1\"]",
					allowedJoinOperators, listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.1", "my_column_4"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_updateJoinOperators(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "LIST"
	var joinColumns = "[\"my_column_1\", \"my_column_2\"]"
	var allowedJoinOperators = "[\"AND\"]"
	var listColumns = "[\"my_column_3\", \"my_column_4\"]"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType, joinColumns,
					allowedJoinOperators, listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.1", "my_column_4"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType, joinColumns,
					"[\"AND\", \"OR\"]", listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.1", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.1", "my_column_4"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_updateListColumns(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "LIST"
	var joinColumns = "[\"my_column_1\", \"my_column_2\"]"
	var allowedJoinOperators = "[\"AND\"]"
	var listColumns = "[\"my_column_3\", \"my_column_4\"]"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType, joinColumns,
					allowedJoinOperators, listColumns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.1", "my_column_4"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType, joinColumns,
					allowedJoinOperators, "[\"my_column_3\"]"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "LIST"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.join_columns.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.allowed_join_operators.0", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.list.0.list_columns.0", "my_column_3"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationAnalysisRuleTypeMultipleAggregateColumns(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateColumnsTwo = "[\"my_column_6\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregationMultipleAggregateColumns(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType, aggregateColumnsTwo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.1.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.1.column_names.0", "my_column_6"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.1.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationAnalysisRuleTypeMultipleOutputConstraints(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var outputConstraintsColumnNameTwo = "my_column_5"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregationMultipleOutputConstraints(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType, outputConstraintsColumnNameTwo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.1.column_name", "my_column_5"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.1.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.1.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationUpdateAggregateColumnsColumnNames(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, "[\"my_column_1\", \"my_column_2\"]",
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.1", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationUpdateAggregateColumnsFunction(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					"COUNT", joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "COUNT"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationUpdateJoinAggregateColumns(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, "[\"my_column_2\", \"my_column_6\"]", allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.1", "my_column_6"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationUpdateAllowedJoinOperators(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, "[\"OR\", \"AND\"]", dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.1", "AND"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationUpdateDimensionColumns(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, "[\"my_column_3\", \"my_column_6\"]", scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.1", "my_column_6"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationUpdateScalarFunctions(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, "[\"TRUNC\", \"ABS\"]",
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.1", "ABS"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationUpdateOutputConstraintsColumnName(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					"my_column_5", outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_5"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_aggregationUpdateOutputConstraintsMinimum(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "AGGREGATION"
	var aggregateColumns = "[\"my_column_1\"]"
	var aggregateFunction = "SUM"
	var joinAggregateColumns = "[\"my_column_2\"]"
	var allowedJoinOperators = "[\"OR\"]"
	var dimensionColumns = "[\"my_column_3\"]"
	var scalarFunctions = "[\"TRUNC\"]"
	var outputConstraintsColumnName = "my_column_4"
	var outputConstraintsMinimum = "2"
	var outputConstraintsType = "COUNT_DISTINCT"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType, aggregateColumns,
					aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns, scalarFunctions,
					outputConstraintsColumnName, "10", outputConstraintsType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType,
						&configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "AGGREGATION"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.aggregate_columns.0.function", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_aggregate_columns.0", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.join_required", "QUERY_RUNNER"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.allowed_join_operators.0", "OR"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.dimension_columns.0", "my_column_3"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.scalar_functions.0", "TRUNC"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.column_name", "my_column_4"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.minimum", "10"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.aggregation.0.output_constraints.0.type", "COUNT_DISTINCT"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_customUpdateAnalysisProvider(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "CUSTOM"
	var customAnalysis = "[\"ANY_QUERY\"]"
	var analysisProviders = "[\"123456789012\"]"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_custom(analysisRuleType, customAnalysis, analysisProviders),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_custom(analysisRuleType, customAnalysis, "[\"123456789012\", \"098765432109\"]"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.1", "098765432109"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_customUpdateDifferentialPrivacy(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "CUSTOM"
	var customAnalysis = "[\"ANY_QUERY\"]"
	var analysisProviders = "[\"123456789012\"]"
	var columns = "my_column_1"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_customDifferentialPrivacy(analysisRuleType, customAnalysis, analysisProviders, columns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.differential_privacy.0.columns.0.name", "my_column_1"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_customDifferentialPrivacy(analysisRuleType, customAnalysis, analysisProviders, "my_column_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.differential_privacy.0.columns.0.name", "my_column_2"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTableAnalysisRule_customDisableDifferentialPrivacy(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTableAnalysisRule cleanrooms.GetConfiguredTableAnalysisRuleOutput
	var analysisRuleType = "CUSTOM"
	var customAnalysis = "[\"ANY_QUERY\"]"
	var analysisProviders = "[\"123456789012\"]"
	var columns = "my_column_1"
	var resourceName = "aws_cleanrooms_configured_table_analysis_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccConfiguredTableAnalysisRuleDestroy(ctx, analysisRuleType),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_customDifferentialPrivacy(analysisRuleType, customAnalysis, analysisProviders, columns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.differential_privacy.0.columns.0.name", "my_column_1"),
				),
			},
			{
				Config: testAccConfiguredTableAnalysisRuleConfig_custom(analysisRuleType, customAnalysis, analysisProviders),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableAnalysisRuleExists(ctx, resourceName, analysisRuleType, &configuredTableAnalysisRule),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_custom_analyses.0", "ANY_QUERY"),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "analysis_rule_policy.0.v1.0.custom.0.allowed_analyses_providers.0", "123456789012"),
				),
			},
		},
	})
}

func testAccCheckConfiguredTableAnalysisRuleExists(ctx context.Context, name string, analysisRuleType string, configuredTableAnalysisRule *cleanrooms.GetConfiguredTableAnalysisRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence,
				tfcleanrooms.ResNameConfiguredTableAnalysisRule, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence,
				tfcleanrooms.ResNameConfiguredTableAnalysisRule, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)
		resp, err := conn.GetConfiguredTableAnalysisRule(ctx, &cleanrooms.GetConfiguredTableAnalysisRuleInput{
			ConfiguredTableIdentifier: aws.String(rs.Primary.ID),
			AnalysisRuleType:          types.ConfiguredTableAnalysisRuleType(analysisRuleType),
		})

		if err != nil {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence,
				tfcleanrooms.ResNameConfiguredTableAnalysisRule, rs.Primary.ID, err)
		}

		*configuredTableAnalysisRule = *resp

		return nil
	}
}

func testAccConfiguredTableAnalysisRuleConfig_basic(analysisRuleType string, joinColumns string,
	allowedJoinOperators string, listColumns string) string {
	return testAccConfiguredTableAnalysisRuleConfig(analysisRuleType, joinColumns, allowedJoinOperators, listColumns)
}

func testAccConfiguredTableAnalysisRuleConfig_aggregation(analysisRuleType string, aggregateColumns string, aggregateFunction string,
	joinAggregateColumns string, allowedJoinOperators string, dimensionColumns string, scalarFunctions string,
	outputConstraintsColumnName string, outputConstraintsMinimum string, outputConstraintsType string) string {
	return testAccConfiguredTableAnalysisRuleAggregation(analysisRuleType, aggregateColumns, aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns,
		scalarFunctions, outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType)
}

func testAccConfiguredTableAnalysisRuleConfig_aggregationMultipleAggregateColumns(analysisRuleType string, aggregateColumns string, aggregateFunction string,
	joinAggregateColumns string, allowedJoinOperators string, dimensionColumns string, scalarFunctions string,
	outputConstraintsColumnName string, outputConstraintsMinimum string, outputConstraintsType string, aggregateColumnsTwo string) string {
	return testAccConfiguredTableAnalysisRuleAggregationMultipleAggregateColumns(analysisRuleType, aggregateColumns, aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns,
		scalarFunctions, outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType, aggregateColumnsTwo)
}

func testAccConfiguredTableAnalysisRuleConfig_aggregationMultipleOutputConstraints(analysisRuleType string, aggregateColumns string, aggregateFunction string,
	joinAggregateColumns string, allowedJoinOperators string, dimensionColumns string, scalarFunctions string,
	outputConstraintsColumnName string, outputConstraintsMinimum string, outputConstraintsType string, outputConstraintsColumnNameTwo string) string {
	return testAccConfiguredTableAnalysisRuleAggregationMultipleOutputConstraints(analysisRuleType, aggregateColumns, aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns,
		scalarFunctions, outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType, outputConstraintsColumnNameTwo)
}

func testAccConfiguredTableAnalysisRuleConfig_custom(analysisRuleType string, customAnalysis string,
	analysisProvider string) string {
	return testAccConfiguredTableAnalysisRuleCustom(analysisRuleType, customAnalysis, analysisProvider)
}

func testAccConfiguredTableAnalysisRuleConfig_customDifferentialPrivacy(analysisRuleType string, customAnalysis string,
	analysisProvider string, columns string) string {
	return testAccConfiguredTableAnalysisRuleCustomDifferentialPrivacy(analysisRuleType, customAnalysis,
		analysisProvider, columns)
}

func testAccConfiguredTableAnalysisRuleDestroy(ctx context.Context, analysisRuleType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != tfcleanrooms.ResNameConfiguredTableAnalysisRule {
				continue
			}

			_, err := conn.GetConfiguredTableAnalysisRule(ctx, &cleanrooms.GetConfiguredTableAnalysisRuleInput{
				ConfiguredTableIdentifier: aws.String(rs.Primary.ID),
				AnalysisRuleType:          types.ConfiguredTableAnalysisRuleType(analysisRuleType),
			})

			if err == nil {
				return create.Error(names.CleanRooms, create.ErrActionCheckingExistence,
					tfcleanrooms.ResNameConfiguredTableAnalysisRule, rs.Primary.ID, errors.New("not destroyed"))
			}
		}

		return nil
	}
}

func testAccConfiguredTableAnalysisRuleCustomDifferentialPrivacy(analysisRuleType string, customAnalysis string,
	analysisProvider string, columns string) string {
	return fmt.Sprintf(`
resource "aws_cleanrooms_configured_table" "test_configured_table" {
  name            = "terraform-example-table"
  description     = "I made this table with terraform!"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = [
    "my_column_1",
    "my_column_2",
    "my_column_3",
    "my_column_4",
    "my_column_5",
    "my_column_6"
  ]

  table_reference {
    database_name = "test-db"
    table_name    = "test-table"
  }
}

resource "aws_cleanrooms_configured_table_analysis_rule" "test" {
  name                        = "test"
  analysis_rule_type          = %[1]q
  configured_table_identifier = aws_cleanrooms_configured_table.test_configured_table.id

  analysis_rule_policy {
    v1 {
      custom {
        allowed_custom_analyses    = jsondecode(%[2]q)
        allowed_analyses_providers = jsondecode(%[3]q)
		differential_privacy {
          columns {
            name = %[4]q
          }
        }
      }
    }
  }
}
	`, analysisRuleType, customAnalysis, analysisProvider, columns)
}

func testAccConfiguredTableAnalysisRuleCustom(analysisRuleType string, customAnalysis string,
	analysisProvider string) string {
	return fmt.Sprintf(`
resource "aws_cleanrooms_configured_table" "test_configured_table" {
  name            = "terraform-example-table"
  description     = "I made this table with terraform!"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = [
    "my_column_1",
    "my_column_2",
    "my_column_3",
    "my_column_4",
    "my_column_5",
    "my_column_6"
  ]

  table_reference {
    database_name = "test-db"
    table_name    = "test-table"
  }
}

resource "aws_cleanrooms_configured_table_analysis_rule" "test" {
  name                        = "test"
  analysis_rule_type          = %[1]q
  configured_table_identifier = aws_cleanrooms_configured_table.test_configured_table.id

  analysis_rule_policy {
    v1 {
      custom {
        allowed_custom_analyses    = jsondecode(%[2]q)
        allowed_analyses_providers = jsondecode(%[3]q)
      }
    }
  }
}
	`, analysisRuleType, customAnalysis, analysisProvider)
}

func testAccConfiguredTableAnalysisRuleAggregationMultipleOutputConstraints(analysisRuleType string,
	aggregateColumns string, aggregateFunction string, joinAggregateColumns string, allowedJoinOperators string,
	dimensionColumns string, scalarFunctions string, outputConstraintsColumnName string, outputConstraintsMinimum string,
	outputConstraintsType string, outputConstraintsColumnNameTwo string) string {
	return fmt.Sprintf(`
resource "aws_cleanrooms_configured_table" "test_configured_table" {
  name            = "terraform-example-table"
  description     = "I made this table with terraform!"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = [
    "my_column_1",
    "my_column_2",
    "my_column_3",
    "my_column_4",
    "my_column_5",
    "my_column_6"
  ]

  table_reference {
    database_name = "test-db"
    table_name    = "test-table"
  }
}

resource "aws_cleanrooms_configured_table_analysis_rule" "test" {
  name                        = "test"
  analysis_rule_type          = %[1]q
  configured_table_identifier = aws_cleanrooms_configured_table.test_configured_table.id

  analysis_rule_policy {
    v1 {
      aggregation {
        aggregate_columns {
          column_names = jsondecode(%[2]q)
          function     = %[3]q
        }
        join_aggregate_columns = jsondecode(%[4]q)
        join_required          = "QUERY_RUNNER"
        allowed_join_operators = jsondecode(%[5]q)
        dimension_columns      = jsondecode(%[6]q)
        scalar_functions       = jsondecode(%[7]q)
        output_constraints {
          column_name = %[8]q
          minimum     = %[9]q
          type        = %[10]q
        }
		output_constraints {
          column_name = %[11]q
          minimum     = %[9]q
          type        = %[10]q
        }
      }
    }
  }
}
	`, analysisRuleType, aggregateColumns, aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns,
		scalarFunctions, outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType, outputConstraintsColumnNameTwo)
}

func testAccConfiguredTableAnalysisRuleAggregationMultipleAggregateColumns(analysisRuleType string,
	aggregateColumns string, aggregateFunction string, joinAggregateColumns string, allowedJoinOperators string,
	dimensionColumns string, scalarFunctions string, outputConstraintsColumnName string, outputConstraintsMinimum string,
	outputConstraintsType string, aggregateColumnsTwo string) string {
	return fmt.Sprintf(`
resource "aws_cleanrooms_configured_table" "test_configured_table" {
  name            = "terraform-example-table"
  description     = "I made this table with terraform!"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = [
    "my_column_1",
    "my_column_2",
    "my_column_3",
    "my_column_4",
    "my_column_5",
    "my_column_6"
  ]

  table_reference {
    database_name = "test-db"
    table_name    = "test-table"
  }
}

resource "aws_cleanrooms_configured_table_analysis_rule" "test" {
  name                        = "test"
  analysis_rule_type          = %[1]q
  configured_table_identifier = aws_cleanrooms_configured_table.test_configured_table.id

  analysis_rule_policy {
    v1 {
      aggregation {
        aggregate_columns {
          column_names = jsondecode(%[2]q)
          function     = %[3]q
        }
		aggregate_columns {
          column_names = jsondecode(%[11]q)
          function     = %[3]q
        }
        join_aggregate_columns = jsondecode(%[4]q)
        join_required          = "QUERY_RUNNER"
        allowed_join_operators = jsondecode(%[5]q)
        dimension_columns      = jsondecode(%[6]q)
        scalar_functions       = jsondecode(%[7]q)
        output_constraints {
          column_name = %[8]q
          minimum     = %[9]q
          type        = %[10]q
        } 
      }
    }
  }
}
	`, analysisRuleType, aggregateColumns, aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns,
		scalarFunctions, outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType, aggregateColumnsTwo)
}

func testAccConfiguredTableAnalysisRuleAggregation(analysisRuleType string, aggregateColumns string, aggregateFunction string,
	joinAggregateColumns string, allowedJoinOperators string, dimensionColumns string, scalarFunctions string,
	outputConstraintsColumnName string, outputConstraintsMinimum string, outputConstraintsType string) string {
	return fmt.Sprintf(`
resource "aws_cleanrooms_configured_table" "test_configured_table" {
  name            = "terraform-example-table"
  description     = "I made this table with terraform!"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = [
    "my_column_1",
    "my_column_2",
    "my_column_3",
    "my_column_4",
    "my_column_5",
    "my_column_6"
  ]

  table_reference {
    database_name = "test-db"
    table_name    = "test-table"
  }
}

resource "aws_cleanrooms_configured_table_analysis_rule" "test" {
  name                        = "test"
  analysis_rule_type          = %[1]q
  configured_table_identifier = aws_cleanrooms_configured_table.test_configured_table.id

  analysis_rule_policy {
    v1 {
      aggregation {
        aggregate_columns {
          column_names = jsondecode(%[2]q)
          function     = %[3]q
        }
        join_aggregate_columns = jsondecode(%[4]q)
        join_required          = "QUERY_RUNNER"
        allowed_join_operators = jsondecode(%[5]q)
        dimension_columns      = jsondecode(%[6]q)
        scalar_functions       = jsondecode(%[7]q)
        output_constraints {
          column_name = %[8]q
          minimum     = %[9]q
          type        = %[10]q
        } 
      }
    }
  }
}
	`, analysisRuleType, aggregateColumns, aggregateFunction, joinAggregateColumns, allowedJoinOperators, dimensionColumns,
		scalarFunctions, outputConstraintsColumnName, outputConstraintsMinimum, outputConstraintsType)
}

func testAccConfiguredTableAnalysisRuleConfig(analysisRuleType string, joinColumns string,
	allowedJoinOperators string, listColumns string) string {
	return fmt.Sprintf(`
resource "aws_cleanrooms_configured_table" "test_configured_table" {
  name            = "terraform-example-table"
  description     = "I made this table with terraform!"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = [
    "my_column_1",
    "my_column_2",
    "my_column_3",
    "my_column_4",
    "my_column_5",
    "my_column_6"
  ]

  table_reference {
    database_name = "test-db"
    table_name    = "test-table"
  }
}

resource "aws_cleanrooms_configured_table_analysis_rule" "test" {
  name                        = "test"
  analysis_rule_type 		  = %[1]q
  configured_table_identifier = aws_cleanrooms_configured_table.test_configured_table.id

  analysis_rule_policy {
    v1 {
	  list {
	    join_columns = jsondecode(%[2]q)
	    allowed_join_operators = jsondecode(%[3]q)
	    list_columns = jsondecode(%[4]q)
	  }
    }
  }
}
	`, analysisRuleType, joinColumns, allowedJoinOperators, listColumns)
}
