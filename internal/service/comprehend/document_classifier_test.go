// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package comprehend_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	"github.com/aws/aws-sdk-go-v2/service/comprehend/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfcomprehend "github.com/hashicorp/terraform-provider-aws/internal/service/comprehend"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccComprehendDocumentClassifier_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "data_access_role_arn", "aws_iam_role.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "comprehend", regexache.MustCompile(fmt.Sprintf(`document-classifier/%s/version/%s$`, rName, uniqueIDPattern()))),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.augmented_manifests.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_format", string(types.DocumentClassifierDataFormatComprehendCsv)),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.label_delimiter", ""),
					resource.TestCheckResourceAttrSet(resourceName, "input_data_config.0.s3_uri"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.test_s3_uri", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, "en"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMode, string(types.DocumentClassifierModeMultiClass)),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					acctest.CheckResourceAttrNameGenerated(resourceName, "version_name"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", sdkid.UniqueIdPrefix),
					resource.TestCheckResourceAttr(resourceName, "volume_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_Mode_singleLabel(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcomprehend.ResourceDocumentClassifier(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_versionName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	vName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	vName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_versionName(rName, vName1, names.AttrKey, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "version_name", vName1),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", ""),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "comprehend", regexache.MustCompile(fmt.Sprintf(`document-classifier/%s/version/%s$`, rName, vName1))),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_versionName(rName, vName2, names.AttrKey, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "version_name", vName2),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", ""),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "comprehend", regexache.MustCompile(fmt.Sprintf(`document-classifier/%s/version/%s$`, rName, vName2))),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", acctest.CtValue2),
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

func TestAccComprehendDocumentClassifier_versionNameEmpty(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_versionNameEmpty(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "version_name", ""),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", ""),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "comprehend", regexache.MustCompile(fmt.Sprintf(`document-classifier/%s$`, rName))),
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

func TestAccComprehendDocumentClassifier_versionNameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_versionNameNotSet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					acctest.CheckResourceAttrNameGenerated(resourceName, "version_name"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", sdkid.UniqueIdPrefix),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "comprehend", regexache.MustCompile(fmt.Sprintf(`document-classifier/%s/version/%s$`, rName, uniqueIDPattern()))),
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

func TestAccComprehendDocumentClassifier_versionNamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_versioNamePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "version_name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", "tf-acc-test-prefix-"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "comprehend", regexache.MustCompile(fmt.Sprintf(`document-classifier/%s/version/%s$`, rName, prefixedUniqueIDPattern("tf-acc-test-prefix-")))),
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

func TestAccComprehendDocumentClassifier_testDocuments(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_testDocuments(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "comprehend", regexache.MustCompile(fmt.Sprintf(`document-classifier/%s/version/%s$`, rName, uniqueIDPattern()))),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.augmented_manifests.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_format", string(types.DocumentClassifierDataFormatComprehendCsv)),
					resource.TestCheckResourceAttrSet(resourceName, "input_data_config.0.test_s3_uri"),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, "en"),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					acctest.CheckResourceAttrNameGenerated(resourceName, "version_name"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", sdkid.UniqueIdPrefix),
					resource.TestCheckResourceAttr(resourceName, "volume_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
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

func TestAccComprehendDocumentClassifier_SingleLabel_ValidateNoDelimiterSet(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccDocumentClassifierConfig_modeDefault_ValidateNoDelimiterSet(rName, tfcomprehend.DocumentClassifierLabelSeparatorDefault),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`input_data_config.label_delimiter must not be set when mode is %s`, types.DocumentClassifierModeMultiClass)),
			},
			{
				Config:      testAccDocumentClassifierConfig_modeDefault_ValidateNoDelimiterSet(rName, ">"),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`input_data_config.label_delimiter must not be set when mode is %s`, types.DocumentClassifierModeMultiClass)),
			},
			{
				Config:      testAccDocumentClassifierConfig_modeSingleLabel_ValidateNoDelimiterSet(rName, tfcomprehend.DocumentClassifierLabelSeparatorDefault),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`input_data_config.label_delimiter must not be set when mode is %s`, types.DocumentClassifierModeMultiClass)),
			},
			{
				Config:      testAccDocumentClassifierConfig_modeSingleLabel_ValidateNoDelimiterSet(rName, ">"),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`input_data_config.label_delimiter must not be set when mode is %s`, types.DocumentClassifierModeMultiClass)),
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_multiLabel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_multiLabel_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "data_access_role_arn", "aws_iam_role.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "comprehend", regexache.MustCompile(fmt.Sprintf(`document-classifier/%s/version/%s$`, rName, uniqueIDPattern()))),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.augmented_manifests.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_format", string(types.DocumentClassifierDataFormatComprehendCsv)),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.label_delimiter", tfcomprehend.DocumentClassifierLabelSeparatorDefault),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, "en"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMode, string(types.DocumentClassifierModeMultiLabel)),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					acctest.CheckResourceAttrNameGenerated(resourceName, "version_name"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", sdkid.UniqueIdPrefix),
					resource.TestCheckResourceAttr(resourceName, "volume_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_multiLabel_defaultDelimiter(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_outputDataConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_basic(rName, "outputs"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "output_data_config.0.s3_uri", regexache.MustCompile(`s3://.+/outputs`)),
					resource.TestMatchResourceAttr(resourceName, "output_data_config.0.output_s3_uri", regexache.MustCompile(`s3://.+/outputs/[0-9A-Za-z-]+/output/output.tar.gz`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_basic(rName, "outputs/"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_outputDataConfig_kmsKeyCreateID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyId(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "output_data_config.0.kms_key_id", "aws_kms_key.output", names.AttrKeyID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyARN(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_outputDataConfig_kmsKeyCreateARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "output_data_config.0.kms_key_id", "aws_kms_key.output", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyId(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_outputDataConfig_kmsKeyCreateAliasName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyAliasName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "output_data_config.0.kms_key_id", "aws_kms_alias.output", names.AttrName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyAliasARN(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_outputDataConfig_kmsKeyCreateAliasARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyAliasARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "output_data_config.0.kms_key_id", "aws_kms_alias.output", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyAliasName(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_outputDataConfig_kmsKeyAdd(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyNone(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.0.kms_key_id", ""),
				),
			},
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeySet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "output_data_config.0.kms_key_id", "aws_kms_key.output", names.AttrKeyID),
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

func TestAccComprehendDocumentClassifier_outputDataConfig_kmsKeyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeySet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "output_data_config.0.kms_key_id", "aws_kms_key.output", names.AttrKeyID),
				),
			},
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "output_data_config.0.kms_key_id", "aws_kms_key.output2", names.AttrKeyID),
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

func TestAccComprehendDocumentClassifier_outputDataConfig_kmsKeyRemove(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeySet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "output_data_config.0.kms_key_id", "aws_kms_key.output", names.AttrKeyID),
				),
			},
			{
				Config: testAccDocumentClassifierConfig_outputDataConfig_kmsKeyNone(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.0.kms_key_id", ""),
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

func TestAccComprehendDocumentClassifier_multiLabel_labelDelimiter(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"
	const delimiter = "~"
	const delimiterUpdated = "/"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_multiLabel_delimiter(rName, delimiter),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "data_access_role_arn", "aws_iam_role.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "comprehend", regexache.MustCompile(fmt.Sprintf(`document-classifier/%s/version/%s$`, rName, uniqueIDPattern()))),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.augmented_manifests.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_format", string(types.DocumentClassifierDataFormatComprehendCsv)),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.label_delimiter", delimiter),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, "en"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMode, string(types.DocumentClassifierModeMultiLabel)),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					acctest.CheckResourceAttrNameGenerated(resourceName, "version_name"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", sdkid.UniqueIdPrefix),
					resource.TestCheckResourceAttr(resourceName, "volume_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_multiLabel_delimiter(rName, delimiterUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.label_delimiter", delimiterUpdated),
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

func TestAccComprehendDocumentClassifier_KMSKeys_CreateIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_kmsKeyIds(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttrPair(resourceName, "model_kms_key_id", "aws_kms_key.model", names.AttrKeyID),
					resource.TestCheckResourceAttrPair(resourceName, "volume_kms_key_id", "aws_kms_key.volume", names.AttrKeyID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_kmsKeyARNs(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_KMSKeys_CreateARNs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var documentclassifier types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_kmsKeyARNs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &documentclassifier),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttrPair(resourceName, "model_kms_key_id", "aws_kms_key.model", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "volume_kms_key_id", "aws_kms_key.volume", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_kmsKeyIds(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_KMSKeys_Add(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_kmsKeys_None(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "volume_kms_key_id", ""),
				),
			},
			{
				Config: testAccDocumentClassifierConfig_kmsKeys_Set(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttrPair(resourceName, "model_kms_key_id", "aws_kms_key.model", names.AttrKeyID),
					resource.TestCheckResourceAttrPair(resourceName, "volume_kms_key_id", "aws_kms_key.volume", names.AttrKeyID),
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

func TestAccComprehendDocumentClassifier_KMSKeys_Update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_kmsKeys_Set(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttrPair(resourceName, "model_kms_key_id", "aws_kms_key.model", names.AttrKeyID),
					resource.TestCheckResourceAttrPair(resourceName, "volume_kms_key_id", "aws_kms_key.volume", names.AttrKeyID),
				),
			},
			{
				Config: testAccDocumentClassifierConfig_kmsKeys_Update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttrPair(resourceName, "model_kms_key_id", "aws_kms_key.model2", names.AttrKeyID),
					resource.TestCheckResourceAttrPair(resourceName, "volume_kms_key_id", "aws_kms_key.volume2", names.AttrKeyID),
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

func TestAccComprehendDocumentClassifier_KMSKeys_Remove(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_kmsKeys_Set(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttrPair(resourceName, "model_kms_key_id", "aws_kms_key.model", names.AttrKeyID),
					resource.TestCheckResourceAttrPair(resourceName, "volume_kms_key_id", "aws_kms_key.volume", names.AttrKeyID),
				),
			},
			{
				Config: testAccDocumentClassifierConfig_kmsKeys_None(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "volume_kms_key_id", ""),
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

func TestAccComprehendDocumentClassifier_VPCConfig_Create(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dc1, dc2 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_vpcConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &dc1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.security_group_ids.*", "aws_security_group.test.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.1", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_vpcConfig_Update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &dc2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.security_group_ids.*", "aws_security_group.test.1", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.2", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.3", names.AttrID),
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

func TestAccComprehendDocumentClassifier_VPCConfig_Add(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dc1, dc2 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_vpcConfig_None(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &dc1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
			{
				Config: testAccDocumentClassifierConfig_vpcConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &dc2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.security_group_ids.*", "aws_security_group.test.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.1", names.AttrID),
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

func TestAccComprehendDocumentClassifier_VPCConfig_Remove(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dc1, dc2 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_vpcConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &dc1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.security_group_ids.*", "aws_security_group.test.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.1", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_vpcConfig_None(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &dc2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
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

func TestAccComprehendDocumentClassifier_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentClassifierConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentClassifierConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v2),
					testAccCheckDocumentClassifierNotRecreated(&v1, &v2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDocumentClassifierConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v3),
					testAccCheckDocumentClassifierNotRecreated(&v2, &v3),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccComprehendDocumentClassifier_DefaultTags_providerOnly(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3 types.DocumentClassifierProperties
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_comprehend_document_classifier.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComprehendEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentClassifierDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtProviderKey1, acctest.CtProviderValue1),
					testAccDocumentClassifierConfig_tags0(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v1),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtProviderValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2(acctest.CtProviderKey1, acctest.CtProviderValue1, "providerkey2", "providervalue2"),
					testAccDocumentClassifierConfig_tags0(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v2),
					testAccCheckDocumentClassifierNotRecreated(&v1, &v2),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtProviderValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtProviderKey1, acctest.CtValue1),
					testAccDocumentClassifierConfig_tags0(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentClassifierExists(ctx, t, resourceName, &v3),
					testAccCheckDocumentClassifierNotRecreated(&v2, &v3),
					testAccCheckDocumentClassifierPublishedVersions(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtValue1),
				),
			},
		},
	})
}

func testAccCheckDocumentClassifierDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ComprehendClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_comprehend_document_classifier" {
				continue
			}

			name, err := tfcomprehend.DocumentClassifierParseARN(rs.Primary.ID)
			if err != nil {
				return err
			}

			input := &comprehend.ListDocumentClassifiersInput{
				Filter: &types.DocumentClassifierFilter{
					DocumentClassifierName: aws.String(name),
				},
			}
			total := 0
			paginator := comprehend.NewListDocumentClassifiersPaginator(conn, input)
			for paginator.HasMorePages() {
				output, err := paginator.NextPage(ctx)
				if err != nil {
					return err
				}
				total += len(output.DocumentClassifierPropertiesList)
			}

			if total != 0 {
				return fmt.Errorf("Expected Comprehend Document Classifier (%s) to be destroyed, found %d versions", rs.Primary.ID, total)
			}
			return nil
		}

		return nil
	}
}

func testAccCheckDocumentClassifierExists(ctx context.Context, t *testing.T, name string, documentclassifier *types.DocumentClassifierProperties) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Comprehend Document Classifier is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ComprehendClient(ctx)

		resp, err := tfcomprehend.FindDocumentClassifierByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error describing Comprehend Document Classifier: %w", err)
		}

		*documentclassifier = *resp

		return nil
	}
}

func testAccCheckDocumentClassifierNotRecreated(before, after *types.DocumentClassifierProperties) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !documentClassifierIdentity(before, after) {
			return fmt.Errorf("Comprehend Document Classifier recreated")
		}

		return nil
	}
}

func documentClassifierIdentity(before, after *types.DocumentClassifierProperties) bool {
	return aws.ToTime(before.SubmitTime).Equal(aws.ToTime(after.SubmitTime))
}

func testAccCheckDocumentClassifierPublishedVersions(ctx context.Context, t *testing.T, name string, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Comprehend Document Classifier is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ComprehendClient(ctx)

		name, err := tfcomprehend.DocumentClassifierParseARN(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &comprehend.ListDocumentClassifiersInput{
			Filter: &types.DocumentClassifierFilter{
				DocumentClassifierName: aws.String(name),
			},
		}
		count := 0
		paginator := comprehend.NewListDocumentClassifiersPaginator(conn, input)
		for paginator.HasMorePages() {
			output, err := paginator.NextPage(ctx)
			if err != nil {
				return err
			}
			count += len(output.DocumentClassifierPropertiesList)
		}

		if count != expected {
			return fmt.Errorf("expected %d published versions, found %d", expected, count)
		}

		return nil
	}
}

func testAccDocumentClassifierConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccDocumentClassifierConfig_Mode_singleLabel(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  mode          = "MULTI_CLASS"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccDocumentClassifierConfig_versionName(rName, vName, key, value string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name         = %[1]q
  version_name = %[2]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {
    %[3]q = %[4]q
  }

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, vName, key, value))
}

func testAccDocumentClassifierConfig_versionNameEmpty(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name         = %[1]q
  version_name = ""

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccDocumentClassifierConfig_versionNameNotSet(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccDocumentClassifierConfig_versioNamePrefix(rName, versionNamePrefix string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name                = %[1]q
  version_name_prefix = %[2]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, versionNamePrefix))
}

func testAccDocumentClassifierConfig_testDocuments(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri      = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
    test_s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccDocumentClassifierConfig_modeDefault_ValidateNoDelimiterSet(rName, delimiter string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri          = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
    label_delimiter = %q
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, delimiter))
}

func testAccDocumentClassifierConfig_modeSingleLabel_ValidateNoDelimiterSet(rName, delimiter string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  mode          = "MULTI_CLASS"
  input_data_config {
    s3_uri          = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
    label_delimiter = %q
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, delimiter))
}

func testAccDocumentClassifierConfig_multiLabel_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_multilabel,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  mode          = "MULTI_LABEL"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.multilabel.bucket}/${aws_s3_object.multilabel.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccDocumentClassifierConfig_multiLabel_defaultDelimiter(rName string) string {
	return testAccDocumentClassifierConfig_multiLabel_delimiter(rName, tfcomprehend.DocumentClassifierLabelSeparatorDefault)
}

func testAccDocumentClassifierConfig_multiLabel_delimiter(rName, delimiter string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_multilabel,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  mode          = "MULTI_LABEL"
  input_data_config {
    s3_uri          = "s3://${aws_s3_object.multilabel.bucket}/${aws_s3_object.multilabel.key}"
    label_delimiter = %[2]q
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, delimiter))
}

func testAccDocumentClassifierConfig_kmsKeyIds(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  model_kms_key_id  = aws_kms_key.model.key_id
  volume_kms_key_id = aws_kms_key.volume.key_id

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.kms_keys,
  ]
}

resource "aws_kms_key" "model" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_key" "volume" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_iam_role_policy" "kms_keys" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.kms_keys.json
}

data "aws_iam_policy_document" "kms_keys" {
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.model.arn,
    ]
  }
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.volume.arn,
    ]
  }
}
`, rName))
}

func testAccDocumentClassifierConfig_kmsKeyARNs(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  model_kms_key_id  = aws_kms_key.model.arn
  volume_kms_key_id = aws_kms_key.volume.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}

resource "aws_iam_role_policy" "kms_keys" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.kms_keys.json
}

data "aws_iam_policy_document" "kms_keys" {
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.model.arn,
    ]
  }
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.volume.arn,
    ]
  }
}

resource "aws_kms_key" "model" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_key" "volume" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`, rName))
}

func testAccDocumentClassifierConfig_kmsKeys_None(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccDocumentClassifierConfig_kmsKeys_Set(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  model_kms_key_id  = aws_kms_key.model.key_id
  volume_kms_key_id = aws_kms_key.volume.key_id

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}

resource "aws_iam_role_policy" "kms_keys" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.kms_keys.json
}

data "aws_iam_policy_document" "kms_keys" {
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.model.arn,
    ]
  }
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.volume.arn,
    ]
  }
}

resource "aws_kms_key" "model" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_key" "volume" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`, rName))
}

func testAccDocumentClassifierConfig_kmsKeys_Update(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  model_kms_key_id  = aws_kms_key.model2.key_id
  volume_kms_key_id = aws_kms_key.volume2.key_id

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}

resource "aws_iam_role_policy" "kms_keys" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.kms_keys.json
}

data "aws_iam_policy_document" "kms_keys" {
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.model2.arn,
    ]
  }
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.volume2.arn,
    ]
  }
}

resource "aws_kms_key" "model2" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_key" "volume2" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`, rName))
}

func testAccDocumentClassifierConfig_tags0(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {}

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccDocumentClassifierConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, tagKey1, tagValue1))
}

func testAccDocumentClassifierConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDocumentClassifierConfig_outputDataConfig_basic(rName, outputPath string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_s3OutputRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.test.bucket}/%[2]s"
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.s3_output,
  ]
}
`, rName, outputPath))
}

func testAccDocumentClassifierConfig_outputDataConfig_kmsKeyId(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_s3OutputRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  output_data_config {
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/outputs"
    kms_key_id = aws_kms_key.output.key_id
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.s3_output,
    aws_iam_role_policy.kms_keys,
  ]
}

resource "aws_kms_key" "output" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_iam_role_policy" "kms_keys" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.kms_keys.json
}

data "aws_iam_policy_document" "kms_keys" {
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.output.arn,
    ]
  }
}
`, rName))
}

func testAccDocumentClassifierConfig_outputDataConfig_kmsKeyARN(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_s3OutputRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  output_data_config {
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/outputs"
    kms_key_id = aws_kms_key.output.arn
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.s3_output,
    aws_iam_role_policy.kms_keys,
  ]
}

resource "aws_kms_key" "output" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_iam_role_policy" "kms_keys" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.kms_keys.json
}

data "aws_iam_policy_document" "kms_keys" {
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.output.arn,
    ]
  }
}
`, rName))
}

func testAccDocumentClassifierConfig_outputDataConfig_kmsKeyAliasName(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_s3OutputRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  output_data_config {
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/outputs"
    kms_key_id = aws_kms_alias.output.name
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.s3_output,
    aws_iam_role_policy.kms_keys,
  ]
}

resource "aws_kms_alias" "output" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.output.key_id
}

resource "aws_kms_key" "output" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_iam_role_policy" "kms_keys" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.kms_keys.json
}

data "aws_iam_policy_document" "kms_keys" {
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.output.arn,
    ]
  }
}
`, rName))
}

func testAccDocumentClassifierConfig_outputDataConfig_kmsKeyAliasARN(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_s3OutputRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  output_data_config {
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/outputs"
    kms_key_id = aws_kms_alias.output.arn
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.s3_output,
    aws_iam_role_policy.kms_keys,
  ]
}

resource "aws_kms_alias" "output" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.output.key_id
}

resource "aws_kms_key" "output" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_iam_role_policy" "kms_keys" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.kms_keys.json
}

data "aws_iam_policy_document" "kms_keys" {
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.output.arn,
    ]
  }
}
`, rName))
}

func testAccDocumentClassifierConfig_outputDataConfig_kmsKeySet(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_s3OutputRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  output_data_config {
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/outputs"
    kms_key_id = aws_kms_key.output.key_id
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.s3_output,
    aws_iam_role_policy.kms_keys,
  ]
}

resource "aws_kms_key" "output" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_iam_role_policy" "kms_keys" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.kms_keys.json
}

data "aws_iam_policy_document" "kms_keys" {
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.output.arn,
    ]
  }
}
`, rName))
}

func testAccDocumentClassifierConfig_outputDataConfig_kmsKeyNone(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_s3OutputRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.test.bucket}/outputs"
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.s3_output,
  ]
}
`, rName))
}

func testAccDocumentClassifierConfig_outputDataConfig_kmsKeyUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_s3OutputRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  output_data_config {
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/outputs"
    kms_key_id = aws_kms_key.output2.key_id
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.s3_output,
    aws_iam_role_policy.kms_keys,
  ]
}

resource "aws_kms_key" "output2" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_iam_role_policy" "kms_keys" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.kms_keys.json
}

data "aws_iam_policy_document" "kms_keys" {
  statement {
    actions = [
      "*",
    ]

    resources = [
      aws_kms_key.output2.arn,
    ]
  }
}
`, rName))
}

func testAccDocumentClassifierS3BucketConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.bucket

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}
`, rName)
}

func testAccDocumentClassifierBasicRoleConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "comprehend.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.role.json
}

data "aws_iam_policy_document" "role" {
  statement {
    actions = [
      "s3:GetObject",
    ]

    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
  }
  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      aws_s3_bucket.test.arn,
    ]
  }
}
`, rName)
}

func testAccDocumentClassifierConfig_vpcRole() string {
	return `
resource "aws_iam_role_policy" "vpc_access" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.vpc_access.json
}

data "aws_iam_policy_document" "vpc_access" {
  statement {
    actions = [
      "ec2:CreateNetworkInterface",
      "ec2:CreateNetworkInterfacePermission",
      "ec2:DeleteNetworkInterface",
      "ec2:DeleteNetworkInterfacePermission",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DescribeVpcs",
      "ec2:DescribeDhcpOptions",
      "ec2:DescribeSubnets",
      "ec2:DescribeSecurityGroups",
    ]

    resources = [
      "*",
    ]
  }
}
`
}

func testAccDocumentClassifierConfig_s3OutputRole() string {
	return `
resource "aws_iam_role_policy" "s3_output" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.s3_output.json
}

data "aws_iam_policy_document" "s3_output" {
  statement {
    actions = [
      "s3:PutObject",
    ]

    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
  }
}
`
}

func testAccDocumentClassifierConfig_vpcConfig(rName string) string {
	const subnetCount = 2
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_vpcRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		configVPCWithSubnetsAndDNS(rName, subnetCount),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  vpc_config {
    security_group_ids = [aws_security_group.test[0].id]
    subnets            = aws_subnet.test[*].id
  }

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.vpc_access,
    aws_vpc_endpoint_route_table_association.test,
  ]
}

resource "aws_security_group" "test" {
  count = 1

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    prefix_list_ids = [aws_vpc_endpoint.s3.prefix_list_id]
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table_association" "test" {
  count = %[2]d

  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.test.id
}

resource "aws_vpc_endpoint_route_table_association" "test" {
  route_table_id  = aws_route_table.test.id
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"
}

resource "aws_vpc_endpoint_policy" "s3" {
  vpc_endpoint_id = aws_vpc_endpoint.s3.id

  policy = data.aws_iam_policy_document.s3_endpoint.json
}

data "aws_iam_policy_document" "s3_endpoint" {
  statement {
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:ListBucket",
      "s3:GetBucketLocation",
      "s3:DeleteObject",
      "s3:ListMultipartUploadParts",
      "s3:AbortMultipartUpload",
    ]

    resources = [
      "*",
    ]
  }
}
`, rName, subnetCount))
}

func testAccDocumentClassifierConfig_vpcConfig_Update(rName string) string {
	const subnetCount = 4
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_vpcRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		configVPCWithSubnetsAndDNS(rName, subnetCount),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  vpc_config {
    security_group_ids = [aws_security_group.test[1].id]
    subnets            = slice(aws_subnet.test[*].id, 2, 4)
  }

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
    aws_iam_role_policy.vpc_access,
    aws_vpc_endpoint_route_table_association.test,
  ]
}

resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    prefix_list_ids = [aws_vpc_endpoint.s3.prefix_list_id]
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table_association" "test" {
  count = %[2]d

  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.test.id
}

resource "aws_vpc_endpoint_route_table_association" "test" {
  route_table_id  = aws_route_table.test.id
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"
}

resource "aws_vpc_endpoint_policy" "s3" {
  vpc_endpoint_id = aws_vpc_endpoint.s3.id

  policy = data.aws_iam_policy_document.s3_endpoint.json
}

data "aws_iam_policy_document" "s3_endpoint" {
  statement {
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:ListBucket",
      "s3:GetBucketLocation",
      "s3:DeleteObject",
      "s3:ListMultipartUploadParts",
      "s3:AbortMultipartUpload",
    ]

    resources = [
      "*",
    ]
  }
}
`, rName, subnetCount))
}

func testAccDocumentClassifierConfig_vpcConfig_None(rName string) string {
	return acctest.ConfigCompose(
		testAccDocumentClassifierBasicRoleConfig(rName),
		testAccDocumentClassifierConfig_vpcRole(),
		testAccDocumentClassifierS3BucketConfig(rName),
		testAccDocumentClassifierConfig_S3_documents,
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_comprehend_document_classifier" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_object.documents.bucket}/${aws_s3_object.documents.key}"
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

const testAccDocumentClassifierConfig_S3_documents = `
resource "aws_s3_object" "documents" {
  bucket = aws_s3_bucket.test.bucket
  key    = "documents.csv"
  source = "test-fixtures/document_classifier/documents.csv"
}
`

const testAccDocumentClassifierConfig_S3_multilabel = `
resource "aws_s3_object" "multilabel" {
  bucket = aws_s3_bucket.test.bucket
  key    = "documents.csv"
  source = "test-fixtures/document_classifier_multilabel/documents.csv"
}
`
