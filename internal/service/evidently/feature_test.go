// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/evidently/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchevidently "github.com/hashicorp/terraform-provider-aws/internal/service/evidently"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEvidentlyFeature_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "evidently", fmt.Sprintf("project/%s/feature/%s", rName, rName2)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "default_variation", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "entity_overrides.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "evaluation_rules.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "evaluation_strategy", string(awstypes.FeatureEvaluationStrategyAllRules)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttrPair(resourceName, "project", "aws_evidently_project.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.FeatureStatusAvailable)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "value_type", string(awstypes.VariationValueTypeString)),
					resource.TestCheckResourceAttr(resourceName, "variations.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         "Variation1",
						"value.#":              acctest.Ct1,
						"value.0.string_value": "test",
					}),
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

func TestAccEvidentlyFeature_updateDefaultVariation(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	variationName1 := "Variation1"
	variationName2 := "Variation2"
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_defaultVariation(rName, rName2, variationName1, variationName2, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", variationName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureConfig_defaultVariation(rName, rName2, variationName1, variationName2, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", variationName2),
				),
			},
		},
	})
}

func TestAccEvidentlyFeature_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_description(rName, rName2, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureConfig_description(rName, rName2, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
				),
			},
		},
	})
}

func TestAccEvidentlyFeature_updateEntityOverrides(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	variationName1 := "Variation1"
	variationName2 := "Variation2"
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_entityOverrides1(rName, rName2, variationName1, variationName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "entity_overrides.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "entity_overrides.test1", variationName1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         variationName1,
						"value.#":              acctest.Ct1,
						"value.0.string_value": "testval1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         variationName2,
						"value.#":              acctest.Ct1,
						"value.0.string_value": "testval2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureConfig_entityOverrides2(rName, rName2, variationName1, variationName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "entity_overrides.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "entity_overrides.test1", variationName2),
					resource.TestCheckResourceAttr(resourceName, "entity_overrides.test2", variationName1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         variationName1,
						"value.#":              acctest.Ct1,
						"value.0.string_value": "testval1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         variationName2,
						"value.#":              acctest.Ct1,
						"value.0.string_value": "testval2",
					}),
				),
			},
		},
	})
}

func TestAccEvidentlyFeature_updateEvaluationStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalEvaluationStategy := string(awstypes.FeatureEvaluationStrategyAllRules)
	updatedEvaluationStategy := string(awstypes.FeatureEvaluationStrategyDefaultVariation)
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_evaluationStrategy(rName, rName2, originalEvaluationStategy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "evaluation_strategy", originalEvaluationStategy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureConfig_evaluationStrategy(rName, rName2, updatedEvaluationStategy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "evaluation_strategy", updatedEvaluationStategy),
				),
			},
		},
	})
}

func TestAccEvidentlyFeature_updateVariationsBoolValue(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalVariationName1 := "Variation1Original"
	updatedVariationName1 := "Variation1Updated"
	originalVariationBoolVal1 := true
	updatedVariationBoolVal1 := false
	variationName2 := "Variation2"
	variationBoolVal2 := true
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_variationsBoolValue1(rName, rName2, originalVariationName1, originalVariationBoolVal1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", originalVariationName1),
					resource.TestCheckResourceAttr(resourceName, "value_type", string(awstypes.VariationValueTypeBoolean)),
					resource.TestCheckResourceAttr(resourceName, "variations.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:       originalVariationName1,
						"value.#":            acctest.Ct1,
						"value.0.bool_value": strconv.FormatBool(originalVariationBoolVal1),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureConfig_variationsBoolValue2(rName, rName2, updatedVariationName1, updatedVariationBoolVal1, variationName2, variationBoolVal2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", variationName2), // update default_variation since the first variation is deleted
					resource.TestCheckResourceAttr(resourceName, "value_type", string(awstypes.VariationValueTypeBoolean)),
					resource.TestCheckResourceAttr(resourceName, "variations.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:       updatedVariationName1,
						"value.#":            acctest.Ct1,
						"value.0.bool_value": strconv.FormatBool(updatedVariationBoolVal1),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:       variationName2,
						"value.#":            acctest.Ct1,
						"value.0.bool_value": strconv.FormatBool(variationBoolVal2),
					}),
				),
			},
		},
	})
}

func TestAccEvidentlyFeature_updateVariationsDoubleValue(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalVariationName1 := "Variation1Original"
	updatedVariationName1 := "Variation1Updated"
	originalVariationDoubleVal1 := 0.0
	updatedVariationDoubleVal1 := 2.2
	variationName2 := "Variation2"
	variationDoubleVal2 := 3
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_variationsDoubleValue1(rName, rName2, originalVariationName1, originalVariationDoubleVal1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", originalVariationName1),
					resource.TestCheckResourceAttr(resourceName, "value_type", string(awstypes.VariationValueTypeDouble)),
					resource.TestCheckResourceAttr(resourceName, "variations.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         originalVariationName1,
						"value.#":              acctest.Ct1,
						"value.0.double_value": strconv.FormatFloat(originalVariationDoubleVal1, 'f', -1, 64),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureConfig_variationsDoubleValue2(rName, rName2, updatedVariationName1, updatedVariationDoubleVal1, variationName2, float64(variationDoubleVal2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", variationName2), // update default_variation since the first variation is deleted
					resource.TestCheckResourceAttr(resourceName, "value_type", string(awstypes.VariationValueTypeDouble)),
					resource.TestCheckResourceAttr(resourceName, "variations.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         updatedVariationName1,
						"value.#":              acctest.Ct1,
						"value.0.double_value": strconv.FormatFloat(updatedVariationDoubleVal1, 'f', -1, 64),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         variationName2,
						"value.#":              acctest.Ct1,
						"value.0.double_value": strconv.FormatFloat(float64(variationDoubleVal2), 'f', -1, 64),
					}),
				),
			},
		},
	})
}

func TestAccEvidentlyFeature_updateVariationsLongValue(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalVariationName1 := "Variation1Original"
	updatedVariationName1 := "Variation1Updated"
	originalVariationLongVal1 := 0
	updatedVariationLongVal1 := 2
	variationName2 := "Variation2"
	variationLongVal2 := 3
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_variationsLongValue1(rName, rName2, originalVariationName1, originalVariationLongVal1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", originalVariationName1),
					resource.TestCheckResourceAttr(resourceName, "value_type", string(awstypes.VariationValueTypeLong)),
					resource.TestCheckResourceAttr(resourceName, "variations.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:       originalVariationName1,
						"value.#":            acctest.Ct1,
						"value.0.long_value": strconv.Itoa(originalVariationLongVal1),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureConfig_variationsLongValue2(rName, rName2, updatedVariationName1, updatedVariationLongVal1, variationName2, variationLongVal2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", variationName2), // update default_variation since the first variation is deleted
					resource.TestCheckResourceAttr(resourceName, "value_type", string(awstypes.VariationValueTypeLong)),
					resource.TestCheckResourceAttr(resourceName, "variations.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:       updatedVariationName1,
						"value.#":            acctest.Ct1,
						"value.0.long_value": strconv.Itoa(updatedVariationLongVal1),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:       variationName2,
						"value.#":            acctest.Ct1,
						"value.0.long_value": strconv.Itoa(variationLongVal2),
					}),
				),
			},
		},
	})
}

func TestAccEvidentlyFeature_updateVariationsStringValue(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalVariationName1 := "Variation1Original"
	updatedVariationName1 := "Variation1Updated"
	originalVariationStringVal1 := "Variation1StringValOriginal"
	updatedVariationStringVal1 := "Variation1StringValUpdated"
	variationName2 := "Variation2"
	variationStringVal2 := "Variation2StringVal"
	updatedVariationStringVal2 := ""
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_variationsStringValue1(rName, rName2, originalVariationName1, originalVariationStringVal1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", originalVariationName1),
					resource.TestCheckResourceAttr(resourceName, "value_type", string(awstypes.VariationValueTypeString)),
					resource.TestCheckResourceAttr(resourceName, "variations.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         originalVariationName1,
						"value.#":              acctest.Ct1,
						"value.0.string_value": originalVariationStringVal1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureConfig_variationsStringValue2(rName, rName2, updatedVariationName1, updatedVariationStringVal1, variationName2, variationStringVal2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", variationName2), // update default_variation since the first variation is deleted
					resource.TestCheckResourceAttr(resourceName, "value_type", string(awstypes.VariationValueTypeString)),
					resource.TestCheckResourceAttr(resourceName, "variations.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         updatedVariationName1,
						"value.#":              acctest.Ct1,
						"value.0.string_value": updatedVariationStringVal1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         variationName2,
						"value.#":              acctest.Ct1,
						"value.0.string_value": variationStringVal2,
					}),
				),
			},
			{
				Config: testAccFeatureConfig_variationsStringValue2(rName, rName2, updatedVariationName1, updatedVariationStringVal1, variationName2, updatedVariationStringVal2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, "default_variation", variationName2), // update default_variation since the first variation is deleted
					resource.TestCheckResourceAttr(resourceName, "value_type", string(awstypes.VariationValueTypeString)),
					resource.TestCheckResourceAttr(resourceName, "variations.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         updatedVariationName1,
						"value.#":              acctest.Ct1,
						"value.0.string_value": updatedVariationStringVal1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						names.AttrName:         variationName2,
						"value.#":              acctest.Ct1,
						"value.0.string_value": updatedVariationStringVal2, // test empty string
					}),
				),
			},
		},
	})
}

func TestAccEvidentlyFeature_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_tags1(rName, rName2, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureConfig_tags2(rName, rName2, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFeatureConfig_tags1(rName, rName2, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEvidentlyFeature_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(ctx, resourceName, &feature),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudwatchevidently.ResourceFeature(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFeatureDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_evidently_feature" {
				continue
			}

			featureName, projectNameOrARN, err := tfcloudwatchevidently.FeatureParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfcloudwatchevidently.FindFeatureWithProjectNameorARN(ctx, conn, featureName, projectNameOrARN)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Evidently Feature %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFeatureExists(ctx context.Context, n string, v *awstypes.Feature) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Evidently Feature ID is set")
		}

		featureName, projectNameOrARN, err := tfcloudwatchevidently.FeatureParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyClient(ctx)

		output, err := tfcloudwatchevidently.FindFeatureWithProjectNameorARN(ctx, conn, featureName, projectNameOrARN)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFeatureConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_evidently_project" "test" {
  name = %[1]q
}
`, rName)
}

func testAccFeatureConfig_basic(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  variations {
    name = "Variation1"
    value {
      string_value = "test"
    }
  }
}
`, rName2))
}

func testAccFeatureConfig_defaultVariation(rName, rName2, variationName1, variationName2, selectDefaultVariation string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
locals {
  select_default_variation = %[4]q
  variation_name1          = %[2]q
  variation_name2          = %[3]q
}

resource "aws_evidently_feature" "test" {
  name              = %[1]q
  project           = aws_evidently_project.test.name
  default_variation = local.select_default_variation == "first" ? local.variation_name1 : local.variation_name2

  variations {
    name = %[2]q
    value {
      string_value = "testval1"
    }
  }

  variations {
    name = %[3]q
    value {
      string_value = "testval2"
    }
  }
}
`, rName2, variationName1, variationName2, selectDefaultVariation))
}

func testAccFeatureConfig_description(rName, rName2, description string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name        = %[1]q
  description = %[2]q
  project     = aws_evidently_project.test.name

  variations {
    name = "Variation1"
    value {
      string_value = "test"
    }
  }
}
`, rName2, description))
}

func testAccFeatureConfig_entityOverrides1(rName, rName2, variationName1, variationName2 string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  entity_overrides = {
    test1 = %[2]q
  }

  variations {
    name = %[2]q
    value {
      string_value = "testval1"
    }
  }

  variations {
    name = %[3]q
    value {
      string_value = "testval2"
    }
  }
}
`, rName2, variationName1, variationName2))
}

func testAccFeatureConfig_entityOverrides2(rName, rName2, variationName1, variationName2 string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  entity_overrides = {
    test1 = %[3]q
    test2 = %[2]q
  }

  variations {
    name = %[2]q
    value {
      string_value = "testval1"
    }
  }

  variations {
    name = %[3]q
    value {
      string_value = "testval2"
    }
  }
}
`, rName2, variationName1, variationName2))
}

func testAccFeatureConfig_evaluationStrategy(rName, rName2, evaluationStrategy string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name                = %[1]q
  evaluation_strategy = %[2]q
  project             = aws_evidently_project.test.name

  variations {
    name = "Variation1"
    value {
      string_value = "test"
    }
  }
}
`, rName2, evaluationStrategy))
}

func testAccFeatureConfig_variationsBoolValue1(rName, rName2, variationName1 string, boolVal1 bool) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  variations {
    name = %[2]q
    value {
      bool_value = %[3]t
    }
  }
}
`, rName2, variationName1, boolVal1))
}

func testAccFeatureConfig_variationsBoolValue2(rName, rName2, variationName1 string, boolVal1 bool, variationName2 string, boolVal2 bool) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name              = %[1]q
  project           = aws_evidently_project.test.name
  default_variation = %[4]q

  variations {
    name = %[2]q
    value {
      bool_value = %[3]t
    }
  }

  variations {
    name = %[4]q
    value {
      bool_value = %[5]t
    }
  }
}
`, rName2, variationName1, boolVal1, variationName2, boolVal2))
}

func testAccFeatureConfig_variationsDoubleValue1(rName, rName2, variationName1 string, doubleVal1 float64) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  variations {
    name = %[2]q
    value {
      double_value = %[3]f
    }
  }
}
`, rName2, variationName1, doubleVal1))
}

func testAccFeatureConfig_variationsDoubleValue2(rName, rName2, variationName1 string, doubleVal1 float64, variationName2 string, doubleVal2 float64) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name              = %[1]q
  project           = aws_evidently_project.test.name
  default_variation = %[4]q

  variations {
    name = %[2]q
    value {
      double_value = %[3]f
    }
  }

  variations {
    name = %[4]q
    value {
      double_value = %[5]f
    }
  }
}
`, rName2, variationName1, doubleVal1, variationName2, doubleVal2))
}

func testAccFeatureConfig_variationsLongValue1(rName, rName2, variationName1 string, longVal1 int) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  variations {
    name = %[2]q
    value {
      long_value = %[3]d
    }
  }
}
`, rName2, variationName1, longVal1))
}

func testAccFeatureConfig_variationsLongValue2(rName, rName2, variationName1 string, longVal1 int, variationName2 string, longVal2 int) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name              = %[1]q
  project           = aws_evidently_project.test.name
  default_variation = %[4]q

  variations {
    name = %[2]q
    value {
      long_value = %[3]d
    }
  }

  variations {
    name = %[4]q
    value {
      long_value = %[5]d
    }
  }
}
`, rName2, variationName1, longVal1, variationName2, longVal2))
}

func testAccFeatureConfig_variationsStringValue1(rName, rName2, variationName1, stringVal1 string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  variations {
    name = %[2]q
    value {
      string_value = %[3]q
    }
  }
}
`, rName2, variationName1, stringVal1))
}

func testAccFeatureConfig_variationsStringValue2(rName, rName2, variationName1, stringVal1, variationName2, stringVal2 string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name              = %[1]q
  project           = aws_evidently_project.test.name
  default_variation = %[4]q

  variations {
    name = %[2]q
    value {
      string_value = %[3]q
    }
  }

  variations {
    name = %[4]q
    value {
      string_value = %[5]q
    }
  }
}
`, rName2, variationName1, stringVal1, variationName2, stringVal2))
}

func testAccFeatureConfig_tags1(rName, rName2, tag, value string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  variations {
    name = "Variation1"
    value {
      string_value = "test"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName2, tag, value))
}

func testAccFeatureConfig_tags2(rName, rName2, tag1, value1, tag2, value2 string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  variations {
    name = "Variation1"
    value {
      string_value = "test"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName2, tag1, value1, tag2, value2))
}
