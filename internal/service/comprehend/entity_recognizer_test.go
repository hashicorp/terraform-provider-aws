package comprehend_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	"github.com/aws/aws-sdk-go-v2/service/comprehend/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcomprehend "github.com/hashicorp/terraform-provider-aws/internal/service/comprehend"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccComprehendEntityRecognizer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "data_access_role_arn", "aws_iam_role.test", "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "comprehend", regexp.MustCompile(fmt.Sprintf(`entity-recognizer/%s/version/%s$`, rName, uniqueIDPattern()))),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.entity_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.annotations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.augmented_manifests.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_format", string(types.EntityRecognizerDataFormatComprehendCsv)),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.documents.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.documents.0.input_format", string(types.InputFormatOneDocPerLine)),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.documents.0.test_s3_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.entity_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					acctest.CheckResourceAttrNameGenerated(resourceName, "version_name"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", resource.UniqueIdPrefix),
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

func TestAccComprehendEntityRecognizer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcomprehend.ResourceEntityRecognizer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccComprehendEntityRecognizer_versionName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_versionName(rName, vName1, "key", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version_name", vName1),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "comprehend", regexp.MustCompile(fmt.Sprintf(`entity-recognizer/%s/version/%s$`, rName, vName1))),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEntityRecognizerConfig_versionName(rName, vName2, "key", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version_name", vName2),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "comprehend", regexp.MustCompile(fmt.Sprintf(`entity-recognizer/%s/version/%s$`, rName, vName2))),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value2"),
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

func TestAccComprehendEntityRecognizer_versionNameEmpty(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_versionNameEmpty(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version_name", ""),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "comprehend", regexp.MustCompile(fmt.Sprintf(`entity-recognizer/%s$`, rName))),
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

func TestAccComprehendEntityRecognizer_versionNameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_versionNameNotSet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					acctest.CheckResourceAttrNameGenerated(resourceName, "version_name"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", resource.UniqueIdPrefix),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "comprehend", regexp.MustCompile(fmt.Sprintf(`entity-recognizer/%s/version/%s$`, rName, uniqueIDPattern()))),
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

func TestAccComprehendEntityRecognizer_versionNamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_versioNamePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "version_name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", "tf-acc-test-prefix-"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "comprehend", regexp.MustCompile(fmt.Sprintf(`entity-recognizer/%s/version/%s$`, rName, prefixedUniqueIDPattern("tf-acc-test-prefix-")))),
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

func TestAccComprehendEntityRecognizer_documents_testDocuments(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_testDocuments(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "comprehend", regexp.MustCompile(fmt.Sprintf(`entity-recognizer/%s/version/%s$`, rName, uniqueIDPattern()))),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.entity_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.annotations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.augmented_manifests.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_format", string(types.EntityRecognizerDataFormatComprehendCsv)),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.documents.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "input_data_config.0.documents.0.test_s3_uri"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.entity_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					acctest.CheckResourceAttrNameGenerated(resourceName, "version_name"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", resource.UniqueIdPrefix),
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

func TestAccComprehendEntityRecognizer_annotations_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_annotations_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "data_access_role_arn", "aws_iam_role.test", "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "comprehend", regexp.MustCompile(fmt.Sprintf(`entity-recognizer/%s/version/%s$`, rName, uniqueIDPattern()))),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.entity_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.annotations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.augmented_manifests.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_format", string(types.EntityRecognizerDataFormatComprehendCsv)),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.documents.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.documents.0.input_format", string(types.InputFormatOneDocPerLine)),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.documents.0.test_s3_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.entity_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					acctest.CheckResourceAttrNameGenerated(resourceName, "version_name"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", resource.UniqueIdPrefix),
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

func TestAccComprehendEntityRecognizer_annotations_testDocuments(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_annotations_testDocuments(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "data_access_role_arn", "aws_iam_role.test", "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "comprehend", regexp.MustCompile(fmt.Sprintf(`entity-recognizer/%s/version/%s$`, rName, uniqueIDPattern()))),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.entity_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.annotations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.augmented_manifests.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_format", string(types.EntityRecognizerDataFormatComprehendCsv)),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.documents.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.documents.0.input_format", string(types.InputFormatOneDocPerLine)),
					resource.TestCheckResourceAttrSet(resourceName, "input_data_config.0.documents.0.test_s3_uri"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.entity_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					acctest.CheckResourceAttrNameGenerated(resourceName, "version_name"),
					resource.TestCheckResourceAttr(resourceName, "version_name_prefix", resource.UniqueIdPrefix),
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

func TestAccComprehendEntityRecognizer_annotations_validateNoTestDocuments(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccEntityRecognizerConfig_annotations_noTestDocuments(rName),
				ExpectError: regexp.MustCompile("input_data_config.documents.test_s3_uri must be set when input_data_config.annotations.test_s3_uri is set"),
			},
		},
	})
}

func TestAccComprehendEntityRecognizer_annotations_validateNoTestAnnotations(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccEntityRecognizerConfig_annotations_noTestAnnotations(rName),
				ExpectError: regexp.MustCompile("input_data_config.annotations.test_s3_uri must be set when input_data_config.documents.test_s3_uri is set"),
			},
		},
	})
}

func TestAccComprehendEntityRecognizer_KMSKeys_CreateIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_kmsKeyIds(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttrPair(resourceName, "model_kms_key_id", "aws_kms_key.model", "key_id"),
					resource.TestCheckResourceAttrPair(resourceName, "volume_kms_key_id", "aws_kms_key.volume", "key_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccEntityRecognizerConfig_kmsKeyARNs(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccComprehendEntityRecognizer_KMSKeys_CreateARNs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_kmsKeyARNs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &entityrecognizer),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttrPair(resourceName, "model_kms_key_id", "aws_kms_key.model", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "volume_kms_key_id", "aws_kms_key.volume", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccEntityRecognizerConfig_kmsKeyIds(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccComprehendEntityRecognizer_KMSKeys_Update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3, v4 types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_kmsKeys_None(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &v1),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "volume_kms_key_id", ""),
				),
			},
			{
				Config: testAccEntityRecognizerConfig_kmsKeys_Set(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &v2),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 2),
					resource.TestCheckResourceAttrPair(resourceName, "model_kms_key_id", "aws_kms_key.model", "key_id"),
					resource.TestCheckResourceAttrPair(resourceName, "volume_kms_key_id", "aws_kms_key.volume", "key_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEntityRecognizerConfig_kmsKeys_Update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &v3),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 3),
					resource.TestCheckResourceAttrPair(resourceName, "model_kms_key_id", "aws_kms_key.model2", "key_id"),
					resource.TestCheckResourceAttrPair(resourceName, "volume_kms_key_id", "aws_kms_key.volume2", "key_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEntityRecognizerConfig_kmsKeys_None(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &v4),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 4),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "volume_kms_key_id", ""),
				),
			},
		},
	})
}

func TestAccComprehendEntityRecognizer_VPCConfig_Create(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var er1, er2 types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_vpcConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &er1),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.security_group_ids.*", "aws_security_group.test.0", "id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.1", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEntityRecognizerConfig_vpcConfig_Update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &er2),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.security_group_ids.*", "aws_security_group.test.1", "id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.2", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.3", "id"),
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

func TestAccComprehendEntityRecognizer_VPCConfig_Update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var er1, er2, er3 types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_vpcConfig_None(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &er1),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
			{
				Config: testAccEntityRecognizerConfig_vpcConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &er2),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.security_group_ids.*", "aws_security_group.test.0", "id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_config.0.subnets.*", "aws_subnet.test.1", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEntityRecognizerConfig_vpcConfig_None(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &er3),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 3),
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

func TestAccComprehendEntityRecognizer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3 types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &v1),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEntityRecognizerConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &v2),
					testAccCheckEntityRecognizerNotRecreated(&v1, &v2),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEntityRecognizerConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &v3),
					testAccCheckEntityRecognizerNotRecreated(&v2, &v3),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccComprehendEntityRecognizer_DefaultTags_providerOnly(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3 types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccEntityRecognizerConfig_tags0(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &v1),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2("providerkey1", "providervalue1", "providerkey2", "providervalue2"),
					testAccEntityRecognizerConfig_tags0(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &v2),
					testAccCheckEntityRecognizerNotRecreated(&v1, &v2),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "value1"),
					testAccEntityRecognizerConfig_tags0(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(ctx, resourceName, &v3),
					testAccCheckEntityRecognizerNotRecreated(&v2, &v3),
					testAccCheckEntityRecognizerPublishedVersions(ctx, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "value1"),
				),
			},
		},
	})
}

func testAccCheckEntityRecognizerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ComprehendClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_comprehend_entity_recognizer" {
				continue
			}

			name, err := tfcomprehend.EntityRecognizerParseARN(rs.Primary.ID)
			if err != nil {
				return err
			}

			input := &comprehend.ListEntityRecognizersInput{
				Filter: &types.EntityRecognizerFilter{
					RecognizerName: aws.String(name),
				},
			}
			total := 0
			paginator := comprehend.NewListEntityRecognizersPaginator(conn, input)
			for paginator.HasMorePages() {
				output, err := paginator.NextPage(ctx)
				if err != nil {
					return err
				}
				total += len(output.EntityRecognizerPropertiesList)
			}

			if total != 0 {
				return fmt.Errorf("Expected Comprehend Entity Recognizer (%s) to be destroyed, found %d versions", rs.Primary.ID, total)
			}
			return nil
		}

		return nil
	}
}

func testAccCheckEntityRecognizerExists(ctx context.Context, name string, entityrecognizer *types.EntityRecognizerProperties) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Comprehend Entity Recognizer is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ComprehendClient()

		resp, err := tfcomprehend.FindEntityRecognizerByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error describing Comprehend Entity Recognizer: %w", err)
		}

		*entityrecognizer = *resp

		return nil
	}
}

// func testAccCheckEntityRecognizerRecreated(before, after *types.EntityRecognizerProperties) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if entityRecognizerIdentity(before, after) {
// 			return fmt.Errorf("Comprehend Entity Recognizer not recreated")
// 		}

// 		return nil
// 	}
// }

func testAccCheckEntityRecognizerNotRecreated(before, after *types.EntityRecognizerProperties) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !entityRecognizerIdentity(before, after) {
			return fmt.Errorf("Comprehend Entity Recognizer recreated")
		}

		return nil
	}
}

func entityRecognizerIdentity(before, after *types.EntityRecognizerProperties) bool {
	return aws.ToTime(before.SubmitTime).Equal(aws.ToTime(after.SubmitTime))
}

func testAccCheckEntityRecognizerPublishedVersions(ctx context.Context, name string, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Comprehend Entity Recognizer is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ComprehendClient()

		name, err := tfcomprehend.EntityRecognizerParseARN(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &comprehend.ListEntityRecognizersInput{
			Filter: &types.EntityRecognizerFilter{
				RecognizerName: aws.String(name),
			},
		}
		count := 0
		paginator := comprehend.NewListEntityRecognizersPaginator(conn, input)
		for paginator.HasMorePages() {
			output, err := paginator.NextPage(ctx)
			if err != nil {
				return err
			}
			count += len(output.EntityRecognizerPropertiesList)
		}

		if count != expected {
			return fmt.Errorf("expected %d published versions, found %d", expected, count)
		}

		return nil
	}
}

func testAccEntityRecognizerConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccEntityRecognizerConfig_versionName(rName, vName, key, value string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name         = %[1]q
  version_name = %[2]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {
    %[3]q = %[4]q
  }

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, vName, key, value))
}

func testAccEntityRecognizerConfig_versionNameEmpty(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name         = %[1]q
  version_name = ""

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccEntityRecognizerConfig_versionNameNotSet(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccEntityRecognizerConfig_versioNamePrefix(rName, versionNamePrefix string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name                = %[1]q
  version_name_prefix = %[2]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, versionNamePrefix))
}

func testAccEntityRecognizerConfig_testDocuments(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri      = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
      test_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccEntityRecognizerConfig_annotations_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_annotations,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    annotations {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.annotations.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccEntityRecognizerConfig_annotations_testDocuments(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_annotations,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri      = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
      test_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    annotations {
      s3_uri      = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.annotations.id}"
      test_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.annotations.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccEntityRecognizerConfig_annotations_noTestDocuments(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_annotations,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    annotations {
      s3_uri      = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.annotations.id}"
      test_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.annotations.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccEntityRecognizerConfig_annotations_noTestAnnotations(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_annotations,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri      = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
      test_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    annotations {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.annotations.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccEntityRecognizerConfig_kmsKeyIds(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  model_kms_key_id  = aws_kms_key.model.key_id
  volume_kms_key_id = aws_kms_key.volume.key_id

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
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
}

resource "aws_kms_key" "volume" {
  deletion_window_in_days = 7
}
`, rName))
}

func testAccEntityRecognizerConfig_kmsKeyARNs(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  model_kms_key_id  = aws_kms_key.model.arn
  volume_kms_key_id = aws_kms_key.volume.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
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
}

resource "aws_kms_key" "volume" {
  deletion_window_in_days = 7
}
`, rName))
}

func testAccEntityRecognizerConfig_kmsKeys_None(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccEntityRecognizerConfig_kmsKeys_Set(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  model_kms_key_id  = aws_kms_key.model.key_id
  volume_kms_key_id = aws_kms_key.volume.key_id

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
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
}

resource "aws_kms_key" "volume" {
  deletion_window_in_days = 7
}
`, rName))
}

func testAccEntityRecognizerConfig_kmsKeys_Update(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  model_kms_key_id  = aws_kms_key.model2.key_id
  volume_kms_key_id = aws_kms_key.volume2.key_id

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
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
}

resource "aws_kms_key" "volume2" {
  deletion_window_in_days = 7
}
`, rName))
}

func testAccEntityRecognizerConfig_tags0(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {}

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

func testAccEntityRecognizerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, tagKey1, tagValue1))
}

func testAccEntityRecognizerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEntityRecognizerS3BucketConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
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

func testAccEntityRecognizerBasicRoleConfig(rName string) string {
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

func testAccEntityRecognizerConfig_vpcRole() string {
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

func testAccEntityRecognizerConfig_vpcConfig(rName string) string {
	const subnetCount = 2
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerConfig_vpcRole(),
		testAccEntityRecognizerS3BucketConfig(rName),
		configVPCWithSubnetsAndDNS(rName, subnetCount),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  vpc_config {
    security_group_ids = [aws_security_group.test[0].id]
    subnets            = aws_subnet.test[*].id
  }

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
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
  count = length(aws_subnet.test)

  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.test.id
}

resource "aws_vpc_endpoint_route_table_association" "test" {
  route_table_id  = aws_route_table.test.id
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
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
`, rName))
}

func testAccEntityRecognizerConfig_vpcConfig_Update(rName string) string {
	const subnetCount = 4
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerConfig_vpcRole(),
		testAccEntityRecognizerS3BucketConfig(rName),
		configVPCWithSubnetsAndDNS(rName, subnetCount),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  vpc_config {
    security_group_ids = [aws_security_group.test[1].id]
    subnets            = slice(aws_subnet.test[*].id, 2, 4)
  }

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
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
  count = length(aws_subnet.test)

  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.test.id
}

resource "aws_vpc_endpoint_route_table_association" "test" {
  route_table_id  = aws_route_table.test.id
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
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
`, rName))
}

func testAccEntityRecognizerConfig_vpcConfig_None(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerConfig_vpcRole(),
		testAccEntityRecognizerS3BucketConfig(rName),
		testAccEntityRecognizerConfig_S3_entityList,
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test,
  ]
}
`, rName))
}

const testAccEntityRecognizerConfig_S3_entityList = `
resource "aws_s3_object" "documents" {
  bucket = aws_s3_bucket.test.bucket
  key    = "documents.txt"
  source = "test-fixtures/entity_recognizer/documents.txt"
}

resource "aws_s3_object" "entities" {
  bucket = aws_s3_bucket.test.bucket
  key    = "entitylist.csv"
  source = "test-fixtures/entity_recognizer/entitylist.csv"
}
`

const testAccEntityRecognizerConfig_S3_annotations = `
resource "aws_s3_object" "documents" {
  bucket = aws_s3_bucket.test.bucket
  key    = "documents.txt"
  source = "test-fixtures/entity_recognizer/documents.txt"
}

resource "aws_s3_object" "annotations" {
  bucket = aws_s3_bucket.test.bucket
  key    = "entitylist.csv"
  source = "test-fixtures/entity_recognizer/annotations.csv"
}
`
