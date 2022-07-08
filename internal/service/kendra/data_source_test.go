package kendra_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, rName2, rName3, rName4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`index/.+/data-source/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.DataSourceStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeCustom)),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
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

func testAccDataSource_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, rName2, rName3, rName4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfkendra.ResourceDataSource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataSource_description(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "Original Description"
	updatedDescription := "Updated Description"
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_description(rName, rName2, rName3, rName4, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`index/.+/data-source/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.DataSourceStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeCustom)),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_description(rName, rName2, rName3, rName4, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`index/.+/data-source/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.DataSourceStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeCustom)),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccDataSource_languageCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalLanguageCode := "en"
	updatedLanguageCode := "zh"
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_languageCode(rName, rName2, rName3, rName4, originalLanguageCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "language_code", originalLanguageCode),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_languageCode(rName, rName2, rName3, rName4, updatedLanguageCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "language_code", updatedLanguageCode),
				),
			},
		},
	})
}

func testAccDataSource_roleARN(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName6 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName7 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_roleARN(rName, rName2, rName3, rName4, rName5, rName6, rName7, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_roleARN(rName, rName2, rName3, rName4, rName5, rName6, rName7, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source2", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func testAccDataSource_schedule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName6 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalSchedule := "cron(9 10 1 * ? *)"
	updatedSchedule := "cron(9 10 2 * ? *)"
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_schedule(rName, rName2, rName3, rName4, rName5, rName6, originalSchedule),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule", originalSchedule),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_schedule(rName, rName2, rName3, rName4, rName5, rName6, updatedSchedule),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule", updatedSchedule),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func testAccDataSource_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_tags1(rName, rName2, rName3, rName4, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
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
				Config: testAccDataSourceConfig_tags2(rName, rName2, rName3, rName4, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDataSourceConfig_tags1(rName, rName2, rName3, rName4, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccDataSource_typeCustomCustomizeDiff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceConfig_typeCustomConflictRoleARN(rName, rName2, rName3, rName4, rName5),
				ExpectError: regexp.MustCompile(`role_arn must not be set when type is CUSTOM`),
			},
			{
				Config:      testAccDataSourceConfig_typeCustomConflictConfiguration(rName, rName2, rName3, rName4, rName5),
				ExpectError: regexp.MustCompile(`configuration must not be set when type is CUSTOM`),
			},
			{
				Config:      testAccDataSourceConfig_typeCustomConflictSchedule(rName, rName2, rName3, rName4, rName5),
				ExpectError: regexp.MustCompile(`schedule must not be set when type is CUSTOM`),
			},
		},
	})
}

func testAccDataSource_Configuration_S3_Bucket(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName6 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName7 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationS3Bucket(rName, rName2, rName3, rName4, rName5, rName6, rName7, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationS3Bucket(rName, rName2, rName3, rName4, rName5, rName6, rName7, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test2", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func testAccDataSource_Configuration_S3_AccessControlList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName6 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationS3AccessControlList(rName, rName2, rName3, rName4, rName5, rName6, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.access_control_list_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.access_control_list_configuration.0.key_path", fmt.Sprintf("s3://%s/path-1", rName5)),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationS3AccessControlList(rName, rName2, rName3, rName4, rName5, rName6, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.access_control_list_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.access_control_list_configuration.0.key_path", fmt.Sprintf("s3://%s/path-2", rName5)),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func testAccDataSource_Configuration_S3_DocumentsMetadataConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName6 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalS3Prefix := "original"
	updatedS3Prefix := "updated"
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationS3DocumentsMetadataConfiguration(rName, rName2, rName3, rName4, rName5, rName6, originalS3Prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.documents_metadata_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.documents_metadata_configuration.0.s3_prefix", originalS3Prefix),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationS3DocumentsMetadataConfiguration(rName, rName2, rName3, rName4, rName5, rName6, updatedS3Prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.documents_metadata_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.documents_metadata_configuration.0.s3_prefix", updatedS3Prefix),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func testAccDataSource_Configuration_S3_ExclusionInclusionPatternsPrefixes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName6 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationS3ExclusionInclusionPatternsPrefixes1(rName, rName2, rName3, rName4, rName5, rName6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.exclusion_patterns.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.exclusion_patterns.*", "example"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_patterns.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_patterns.*", "hello"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_prefixes.*", "world"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationS3ExclusionInclusionPatternsPrefixes2(rName, rName2, rName3, rName4, rName5, rName6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.exclusion_patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.exclusion_patterns.*", "example2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.exclusion_patterns.*", "foo"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_patterns.*", "hello2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_patterns.*", "bar"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_prefixes.*", "world2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_prefixes.*", "baz"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func testAccDataSource_CustomDocumentEnrichmentConfiguration_InlineConfigurations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName6 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationInlineConfigurations1(rName, rName2, rName3, rName4, rName5, rName6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.*", map[string]string{
						"document_content_deletion": "false",

						"condition.#": "1",
						"condition.0.condition_document_attribute_key":  "_document_title",
						"condition.0.operator":                          "Equals",
						"condition.0.condition_on_value.#":              "1",
						"condition.0.condition_on_value.0.string_value": "foo",

						"target.#":                               "1",
						"target.0.target_document_attribute_key": "_category",
						"target.0.target_document_attribute_value_deletion":       "false",
						"target.0.target_document_attribute_value.#":              "1",
						"target.0.target_document_attribute_value.0.string_value": "bar",
					}),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationInlineConfigurations2(rName, rName2, rName3, rName4, rName5, rName6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_data_source", "arn"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.*", map[string]string{
						"document_content_deletion": "false",

						"condition.#": "1",
						"condition.0.condition_document_attribute_key":  "_document_title",
						"condition.0.operator":                          "Equals",
						"condition.0.condition_on_value.#":              "1",
						"condition.0.condition_on_value.0.string_value": "foo2",

						"target.#":                               "1",
						"target.0.target_document_attribute_key": "_category",
						"target.0.target_document_attribute_value_deletion":       "true",
						"target.0.target_document_attribute_value.#":              "1",
						"target.0.target_document_attribute_value.0.string_value": "bar2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.*", map[string]string{
						"document_content_deletion": "false",

						"condition.#": "1",
						"condition.0.condition_document_attribute_key": "_created_at",
						"condition.0.operator":                         "Equals",
						"condition.0.condition_on_value.#":             "1",
						"condition.0.condition_on_value.0.date_value":  "2006-01-02T15:04:05+02:00",

						"target.#":                               "1",
						"target.0.target_document_attribute_key": "_authors",
						"target.0.target_document_attribute_value_deletion":              "false",
						"target.0.target_document_attribute_value.#":                     "1",
						"target.0.target_document_attribute_value.0.string_list_value.#": "2",
					}),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func testAccCheckDataSourceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kendra_data_source" {
			continue
		}

		id, indexId, err := tfkendra.DataSourceParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = tfkendra.FindDataSourceByID(context.TODO(), conn, id, indexId)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckDataSourceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kendra Data Source is set")
		}

		id, indexId, err := tfkendra.DataSourceParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

		_, err = tfkendra.FindDataSourceByID(context.TODO(), conn, id, indexId)

		if err != nil {
			return fmt.Errorf("Error describing Kendra Data Source: %s", err.Error())
		}

		return nil
	}
}

func testAccDataSourceConfigBase(rName, rName2, rName3 string) string {
	// Kendra IAM policies: https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_kms_key" "this" {
  key_id = "alias/aws/kendra"
}
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["kendra.amazonaws.com"]
    }
  }
}
data "aws_iam_policy_document" "test_index" {
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData"
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "cloudwatch:namespace"

      values = [
        "Kendra"
      ]
    }
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogGroups"
    ]
    resources = ["*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*"
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogStreams",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*:log-stream:*"
    ]
  }
}

resource "aws_iam_policy" "test_index" {
  name        = %[1]q
  description = "Kendra Index IAM permissions"
  policy      = data.aws_iam_policy_document.test_index.json
}

resource "aws_iam_role_policy_attachment" "test_index" {
  role       = aws_iam_role.test_index.name
  policy_arn = aws_iam_policy.test_index.arn
}

resource "aws_iam_role" "test_index" {
  name               = %[2]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_kendra_index" "test" {
  depends_on = [aws_iam_role_policy_attachment.test_index]
  name       = %[3]q
  role_arn   = aws_iam_role.test_index.arn
}
`, rName, rName2, rName3)
}

func testAccDataSourceConfig_basic(rName, rName2, rName3, rName4 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"
}
`, rName4))
}

func testAccDataSourceConfig_description(rName, rName2, rName3, rName4, description string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id    = aws_kendra_index.test.id
  name        = %[1]q
  description = %[2]q
  type        = "CUSTOM"
}
`, rName4, description))
}

func testAccDataSourceConfig_languageCode(rName, rName2, rName3, rName4, languageCode string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id      = aws_kendra_index.test.id
  name          = %[1]q
  language_code = %[2]q
  type          = "CUSTOM"
}
`, rName4, languageCode))
}

func testAccDataSourceConfig_roleARN(rName, rName2, rName3, rName4, rName5, rName6, rName7, selectARN string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
locals {
  select_arn = %[5]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source" {
  name               = %[3]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test.arn
        },
        {
          Action = [
            "kendra:BatchPutDocument",
            "kendra:BatchDeleteDocument",
          ]
          Effect   = "Allow"
          Resource = aws_kendra_index.test.arn
        },
      ]
    })
  }
}

resource "aws_iam_role" "test_data_source2" {
  name               = %[4]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test.arn
        },
        {
          Action = [
            "kendra:BatchPutDocument",
            "kendra:BatchDeleteDocument",
          ]
          Effect   = "Allow"
          Resource = aws_kendra_index.test.arn
        },
      ]
    })
  }
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = local.select_arn == "first" ? aws_iam_role.test_data_source.arn : aws_iam_role.test_data_source2.arn

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`, rName4, rName5, rName6, rName7, selectARN))
}

func testAccDataSourceConfig_schedule(rName, rName2, rName3, rName4, rName5, rName6, schedule string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source" {
  name               = %[3]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test.arn
        },
        {
          Action = [
            "kendra:BatchPutDocument",
            "kendra:BatchDeleteDocument",
          ]
          Effect   = "Allow"
          Resource = aws_kendra_index.test.arn
        },
      ]
    })
  }
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn
  schedule = %[4]q

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`, rName4, rName5, rName6, schedule))
}

func testAccDataSourceConfig_tags1(rName, rName2, rName3, rName4, tag, value string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName4, tag, value))
}

func testAccDataSourceConfig_tags2(rName, rName2, rName3, rName4, tag1, value1, tag2, value2 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName4, tag1, value1, tag2, value2))
}

func testAccDataSourceConfig_typeCustomConflictRoleARN(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_iam_role" "test_data_source" {
  name               = %[2]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"
  role_arn = aws_iam_role.test_data_source.arn
}
`, rName4, rName5))
}

func testAccDataSourceConfig_typeCustomConflictSchedule(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"
  schedule = "cron(9 10 1 * ? *)"
}
`, rName4, rName5))
}

func testAccDataSourceConfig_typeCustomConflictConfiguration(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`, rName4, rName5))
}

func testAccDataSourceConfig_configurationS3Bucket(rName, rName2, rName3, rName4, rName5, rName6, rName7, selectBucket string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
locals {
  select_bucket = %[5]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  bucket        = %[3]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source" {
  name               = %[4]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test.arn}/*"
        },
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test2.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test.arn
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test2.arn
        },
        {
          Action = [
            "kendra:BatchPutDocument",
            "kendra:BatchDeleteDocument",
          ]
          Effect   = "Allow"
          Resource = aws_kendra_index.test.arn
        },
      ]
    })
  }
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      bucket_name = local.select_bucket == "first" ? aws_s3_bucket.test.id : aws_s3_bucket.test2.id
    }
  }
}
`, rName4, rName5, rName6, rName7, selectBucket))
}

func testAccDataSourceConfig_configurationS3AccessControlList(rName, rName2, rName3, rName4, rName5, rName6, selectKeyPath string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
locals {
  select_key_path = %[4]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source" {
  name               = %[3]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test.arn
        },
        {
          Action = [
            "kendra:BatchPutDocument",
            "kendra:BatchDeleteDocument",
          ]
          Effect   = "Allow"
          Resource = aws_kendra_index.test.arn
        },
      ]
    })
  }
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      access_control_list_configuration {
        key_path = local.select_key_path == "first" ? "s3://${aws_s3_bucket.test.id}/path-1" : "s3://${aws_s3_bucket.test.id}/path-2"
      }
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`, rName4, rName5, rName6, selectKeyPath))
}

func testAccDataSourceConfig_configurationS3DocumentsMetadataConfiguration(rName, rName2, rName3, rName4, rName5, rName6, s3Prefix string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source" {
  name               = %[3]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test.arn
        },
        {
          Action = [
            "kendra:BatchPutDocument",
            "kendra:BatchDeleteDocument",
          ]
          Effect   = "Allow"
          Resource = aws_kendra_index.test.arn
        },
      ]
    })
  }
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id

      documents_metadata_configuration {
        s3_prefix = %[4]q
      }
    }
  }
}
`, rName4, rName5, rName6, s3Prefix))
}

func testAccDataSourceConfig_configurationS3ExclusionInclusionPatternsPrefixes1(rName, rName2, rName3, rName4, rName5, rName6 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source" {
  name               = %[3]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test.arn
        },
        {
          Action = [
            "kendra:BatchPutDocument",
            "kendra:BatchDeleteDocument",
          ]
          Effect   = "Allow"
          Resource = aws_kendra_index.test.arn
        },
      ]
    })
  }
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id

      exclusion_patterns = ["example"]
      inclusion_patterns = ["hello"]
      inclusion_prefixes = ["world"]
    }
  }
}
`, rName4, rName5, rName6))
}

func testAccDataSourceConfig_configurationS3ExclusionInclusionPatternsPrefixes2(rName, rName2, rName3, rName4, rName5, rName6 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source" {
  name               = %[3]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test.arn
        },
        {
          Action = [
            "kendra:BatchPutDocument",
            "kendra:BatchDeleteDocument",
          ]
          Effect   = "Allow"
          Resource = aws_kendra_index.test.arn
        },
      ]
    })
  }
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id

      exclusion_patterns = ["example2", "foo"]
      inclusion_patterns = ["hello2", "bar"]
      inclusion_prefixes = ["world2", "baz"]
    }
  }
}
`, rName4, rName5, rName6))
}

func testAccDataSourceConfig_customDocumentEnrichmentConfigurationInlineConfigurations1(rName, rName2, rName3, rName4, rName5, rName6 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source" {
  name               = %[3]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test.arn
        },
        {
          Action = [
            "kendra:BatchPutDocument",
            "kendra:BatchDeleteDocument",
          ]
          Effect   = "Allow"
          Resource = aws_kendra_index.test.arn
        },
      ]
    })
  }
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id
    }
  }

  custom_document_enrichment_configuration {
    inline_configurations {
      document_content_deletion = false

      condition {
        condition_document_attribute_key = "_document_title"
        operator                         = "Equals"

        condition_on_value {
          string_value = "foo"
        }
      }

      target {
        target_document_attribute_key            = "_category"
        target_document_attribute_value_deletion = false

        target_document_attribute_value {
          string_value = "bar"
        }
      }
    }
  }
}
`, rName4, rName5, rName6))
}

func testAccDataSourceConfig_customDocumentEnrichmentConfigurationInlineConfigurations2(rName, rName2, rName3, rName4, rName5, rName6 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source" {
  name               = %[3]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:GetObject"]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test.arn
        },
        {
          Action = [
            "kendra:BatchPutDocument",
            "kendra:BatchDeleteDocument",
          ]
          Effect   = "Allow"
          Resource = aws_kendra_index.test.arn
        },
      ]
    })
  }
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id
    }
  }

  custom_document_enrichment_configuration {
    inline_configurations {
      document_content_deletion = false

      condition {
        condition_document_attribute_key = "_document_title"
        operator                         = "Equals"

        condition_on_value {
          string_value = "foo2"
        }
      }

      target {
        target_document_attribute_key            = "_category"
        target_document_attribute_value_deletion = true

        target_document_attribute_value {
          string_value = "bar2"
        }
      }
    }

    inline_configurations {
      document_content_deletion = false

      condition {
        condition_document_attribute_key = "_created_at"
        operator                         = "Equals"

        condition_on_value {
          date_value = "2006-01-02T15:04:05+02:00"
        }
      }

      target {
        target_document_attribute_key            = "_authors"
        target_document_attribute_value_deletion = false

        target_document_attribute_value {
          string_list_value = ["foo", "baz"]
        }
      }
    }
  }
}
`, rName4, rName5, rName6))
}
