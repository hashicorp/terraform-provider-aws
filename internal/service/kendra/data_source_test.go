// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKendraDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, rName2, rName3, rName4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "kendra", regexache.MustCompile(`index/.+/data-source/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, "en"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName4),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.DataSourceStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeCustom)),
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

func TestAccKendraDataSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, rName2, rName3, rName4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkendra.ResourceDataSource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKendraDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, rName2, rName3, rName4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName4),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_basic(rName, rName2, rName3, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName5),
				),
			},
		},
	})
}

func TestAccKendraDataSource_description(t *testing.T) {
	ctx := acctest.Context(t)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_description(rName, rName2, rName3, rName4, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "kendra", regexache.MustCompile(`index/.+/data-source/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, "en"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName4),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.DataSourceStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeCustom)),
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
					testAccCheckDataSourceExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "kendra", regexache.MustCompile(`index/.+/data-source/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, "en"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName4),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.DataSourceStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeCustom)),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccKendraDataSource_languageCode(t *testing.T) {
	ctx := acctest.Context(t)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_languageCode(rName, rName2, rName3, rName4, originalLanguageCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, originalLanguageCode),
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
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, updatedLanguageCode),
				),
			},
		},
	})
}

func TestAccKendraDataSource_roleARN(t *testing.T) {
	ctx := acctest.Context(t)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_roleARN(rName, rName2, rName3, rName4, rName5, rName6, rName7, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
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
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_schedule(t *testing.T) {
	ctx := acctest.Context(t)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_schedule(rName, rName2, rName3, rName4, rName5, rName6, originalSchedule),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, originalSchedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
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
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, updatedSchedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_tags1(rName, rName2, rName3, rName4, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
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
				Config: testAccDataSourceConfig_tags2(rName, rName2, rName3, rName4, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDataSourceConfig_tags1(rName, rName2, rName3, rName4, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKendraDataSource_typeCustomCustomizeDiff(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceConfig_typeCustomConflictRoleARN(rName, rName2, rName3, rName4, rName5),
				ExpectError: regexache.MustCompile(`role_arn must not be set when type is CUSTOM`),
			},
			{
				Config:      testAccDataSourceConfig_typeCustomConflictConfiguration(rName, rName2, rName3, rName4, rName5),
				ExpectError: regexache.MustCompile(`configuration must not be set when type is CUSTOM`),
			},
			{
				Config:      testAccDataSourceConfig_typeCustomConflictSchedule(rName, rName2, rName3, rName4, rName5),
				ExpectError: regexache.MustCompile(`schedule must not be set when type is CUSTOM`),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_S3_Bucket(t *testing.T) {
	ctx := acctest.Context(t)
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
	rName8 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationS3Bucket(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationS3Bucket(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test2", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_S3_AccessControlList(t *testing.T) {
	ctx := acctest.Context(t)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationS3AccessControlList(rName, rName2, rName3, rName4, rName5, rName6, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.access_control_list_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.access_control_list_configuration.0.key_path", fmt.Sprintf("s3://%s/path-1", rName4)),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
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
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.access_control_list_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.access_control_list_configuration.0.key_path", fmt.Sprintf("s3://%s/path-2", rName4)),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_S3_DocumentsMetadataConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationS3DocumentsMetadataConfiguration(rName, rName2, rName3, rName4, rName5, rName6, originalS3Prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.documents_metadata_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.documents_metadata_configuration.0.s3_prefix", originalS3Prefix),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
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
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.documents_metadata_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.documents_metadata_configuration.0.s3_prefix", updatedS3Prefix),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_S3_ExclusionInclusionPatternsPrefixes(t *testing.T) {
	ctx := acctest.Context(t)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationS3ExclusionInclusionPatternsPrefixes1(rName, rName2, rName3, rName4, rName5, rName6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.exclusion_patterns.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.exclusion_patterns.*", "example"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_patterns.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_patterns.*", "hello"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_prefixes.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_prefixes.*", "world"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
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
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.exclusion_patterns.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.exclusion_patterns.*", "example2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.exclusion_patterns.*", "foo"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_patterns.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_patterns.*", "hello2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_patterns.*", "bar"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_prefixes.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_prefixes.*", "world2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.s3_configuration.0.inclusion_prefixes.*", "baz"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_URLsSeedURLs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsSeedURLs(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.crawl_depth", acctest.Ct2),
					// resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.max_content_size_per_page_in_mega_bytes", "50"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.max_links_per_page", "100"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.max_urls_per_minute_crawl_rate", "300"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.0.seed_urls.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.0.seed_urls.*", "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsSeedURLs2(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.crawl_depth", acctest.Ct2),
					// resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.max_content_size_per_page_in_mega_bytes", "50"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.max_links_per_page", "100"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.max_urls_per_minute_crawl_rate", "300"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.0.seed_urls.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.0.seed_urls.*", "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.0.seed_urls.*", "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_faq"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_URLsWebCrawlerMode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	originalWebCrawlerMode := string(types.WebCrawlerModeHostOnly)
	updatedWebCrawlerMode := string(types.WebCrawlerModeSubdomains)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsWebCrawlerMode(rName, rName2, rName3, rName4, rName5, originalWebCrawlerMode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.0.web_crawler_mode", originalWebCrawlerMode),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsWebCrawlerMode(rName, rName2, rName3, rName4, rName5, updatedWebCrawlerMode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.seed_url_configuration.0.web_crawler_mode", updatedWebCrawlerMode),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_AuthenticationConfigurationBasicHostPort(t *testing.T) {
	ctx := acctest.Context(t)
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

	originalHost1 := "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
	originalPort1 := 123
	updatedHost1 := "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_faq"
	updatedPort1 := 234

	host2 := "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_experience"
	port2 := 456

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationAuthenticationConfigurationBasicHostPort(rName, rName2, rName3, rName4, rName5, rName6, rName7, originalHost1, originalPort1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.0.credentials", "aws_secretsmanager_secret.test", names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.*", map[string]string{
						"host":         originalHost1,
						names.AttrPort: strconv.Itoa(originalPort1),
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationAuthenticationConfigurationBasicHostPort2(rName, rName2, rName3, rName4, rName5, rName6, rName7, updatedHost1, updatedPort1, host2, port2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.0.credentials", "aws_secretsmanager_secret.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.1.credentials", "aws_secretsmanager_secret.test", names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.*", map[string]string{
						"host":         updatedHost1,
						names.AttrPort: strconv.Itoa(updatedPort1),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.*", map[string]string{
						"host":         host2,
						names.AttrPort: strconv.Itoa(port2),
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_AuthenticationConfigurationBasicCredentials(t *testing.T) {
	ctx := acctest.Context(t)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationAuthenticationConfigurationBasicCredentials(rName, rName2, rName3, rName4, rName5, rName6, rName7, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.0.credentials", "aws_secretsmanager_secret.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationAuthenticationConfigurationBasicCredentials(rName, rName2, rName3, rName4, rName5, rName6, rName7, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.web_crawler_configuration.0.authentication_configuration.0.basic_authentication.0.credentials", "aws_secretsmanager_secret.test2", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_CrawlDepth(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"
	originalCrawlDepth := 5
	updatedCrawlDepth := 4

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationCrawlDepth(rName, rName2, rName3, rName4, rName5, originalCrawlDepth),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.crawl_depth", strconv.Itoa(originalCrawlDepth)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationCrawlDepth(rName, rName2, rName3, rName4, rName5, updatedCrawlDepth),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.crawl_depth", strconv.Itoa(updatedCrawlDepth)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_MaxLinksPerPage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"
	originalMaxLinksPerPage := 100
	updatedMaxLinksPerPage := 110

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationMaxLinksPerPage(rName, rName2, rName3, rName4, rName5, originalMaxLinksPerPage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.max_links_per_page", strconv.Itoa(originalMaxLinksPerPage)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationMaxLinksPerPage(rName, rName2, rName3, rName4, rName5, updatedMaxLinksPerPage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.max_links_per_page", strconv.Itoa(updatedMaxLinksPerPage)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_MaxURLsPerMinuteCrawlRate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"
	originalMaxUrlsPerMinuteCrawlRate := 200
	updatedMaxUrlsPerMinuteCrawlRate := 250

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationMaxURLsPerMinuteCrawlRate(rName, rName2, rName3, rName4, rName5, originalMaxUrlsPerMinuteCrawlRate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.max_urls_per_minute_crawl_rate", strconv.Itoa(originalMaxUrlsPerMinuteCrawlRate)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationMaxURLsPerMinuteCrawlRate(rName, rName2, rName3, rName4, rName5, updatedMaxUrlsPerMinuteCrawlRate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.max_urls_per_minute_crawl_rate", strconv.Itoa(updatedMaxUrlsPerMinuteCrawlRate)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_ProxyConfigurationCredentials(t *testing.T) {
	ctx := acctest.Context(t)
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

	originalHost1 := "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
	originalPort1 := 123

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationProxyConfigurationHostPort(rName, rName2, rName3, rName4, rName5, rName6, rName7, originalHost1, originalPort1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.proxy_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.web_crawler_configuration.0.proxy_configuration.*", map[string]string{
						"host":         originalHost1,
						names.AttrPort: strconv.Itoa(originalPort1),
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationProxyConfigurationCredentials(rName, rName2, rName3, rName4, rName5, rName6, rName7, originalHost1, originalPort1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.proxy_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.web_crawler_configuration.0.proxy_configuration.0.credentials", "aws_secretsmanager_secret.test", names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.web_crawler_configuration.0.proxy_configuration.*", map[string]string{
						"host":         originalHost1,
						names.AttrPort: strconv.Itoa(originalPort1),
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_ProxyConfigurationHostPort(t *testing.T) {
	ctx := acctest.Context(t)
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

	originalHost1 := "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
	originalPort1 := 123
	updatedHost1 := "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_faq"
	updatedPort1 := 234

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationProxyConfigurationHostPort(rName, rName2, rName3, rName4, rName5, rName6, rName7, originalHost1, originalPort1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.proxy_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.web_crawler_configuration.0.proxy_configuration.*", map[string]string{
						"host":         originalHost1,
						names.AttrPort: strconv.Itoa(originalPort1),
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationProxyConfigurationHostPort(rName, rName2, rName3, rName4, rName5, rName6, rName7, updatedHost1, updatedPort1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.proxy_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.web_crawler_configuration.0.proxy_configuration.*", map[string]string{
						"host":         updatedHost1,
						names.AttrPort: strconv.Itoa(updatedPort1),
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_URLExclusionInclusionPatterns(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationURLExclusionInclusionPatterns(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.url_exclusion_patterns.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.url_exclusion_patterns.*", "example"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.url_inclusion_patterns.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.url_inclusion_patterns.*", "hello"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationURLExclusionInclusionPatterns2(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.url_exclusion_patterns.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.url_exclusion_patterns.*", "example2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.url_exclusion_patterns.*", "foo"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.url_inclusion_patterns.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.url_inclusion_patterns.*", "hello2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.url_inclusion_patterns.*", "bar"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_Configuration_WebCrawler_URLsSiteMaps(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsSiteMapsConfiguration(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.site_maps_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.site_maps_configuration.0.site_maps.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.site_maps_configuration.0.site_maps.*", "https://registry.terraform.io/sitemap.xml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsSiteMapsConfiguration2(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.site_maps_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.site_maps_configuration.0.site_maps.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.site_maps_configuration.0.site_maps.*", "https://registry.terraform.io/sitemap.xml"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.web_crawler_configuration.0.urls.0.site_maps_configuration.0.site_maps.*", "https://www.terraform.io/sitemap.xml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeWebcrawler)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_CustomDocumentEnrichmentConfiguration_InlineConfigurations(t *testing.T) {
	ctx := acctest.Context(t)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationInlineConfigurations1(rName, rName2, rName3, rName4, rName5, rName6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.*", map[string]string{
						"document_content_deletion": acctest.CtFalse,

						"condition.#": acctest.Ct1,
						"condition.0.condition_document_attribute_key":  "_document_title",
						"condition.0.operator":                          string(types.ConditionOperatorEquals),
						"condition.0.condition_on_value.#":              acctest.Ct1,
						"condition.0.condition_on_value.0.string_value": "foo",

						"target.#":                               acctest.Ct1,
						"target.0.target_document_attribute_key": "_category",
						"target.0.target_document_attribute_value_deletion":       acctest.CtFalse,
						"target.0.target_document_attribute_value.#":              acctest.Ct1,
						"target.0.target_document_attribute_value.0.string_value": "bar",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
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
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.*", map[string]string{
						"document_content_deletion": acctest.CtFalse,

						"condition.#": acctest.Ct1,
						"condition.0.condition_document_attribute_key":  "_document_title",
						"condition.0.operator":                          string(types.ConditionOperatorEquals),
						"condition.0.condition_on_value.#":              acctest.Ct1,
						"condition.0.condition_on_value.0.string_value": "foo2",

						"target.#":                               acctest.Ct1,
						"target.0.target_document_attribute_key": "_category",
						"target.0.target_document_attribute_value_deletion":       acctest.CtTrue,
						"target.0.target_document_attribute_value.#":              acctest.Ct1,
						"target.0.target_document_attribute_value.0.string_value": "bar2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.*", map[string]string{
						"document_content_deletion": acctest.CtFalse,

						"condition.#": acctest.Ct1,
						"condition.0.condition_document_attribute_key": "_created_at",
						"condition.0.operator":                         string(types.ConditionOperatorEquals),
						"condition.0.condition_on_value.#":             acctest.Ct1,
						"condition.0.condition_on_value.0.date_value":  "2006-01-02T15:04:05+00:00",

						"target.#":                               acctest.Ct1,
						"target.0.target_document_attribute_key": "_authors",
						"target.0.target_document_attribute_value_deletion":              acctest.CtFalse,
						"target.0.target_document_attribute_value.#":                     acctest.Ct1,
						"target.0.target_document_attribute_value.0.string_list_value.#": acctest.Ct2,
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationInlineConfigurations3(rName, rName2, rName3, rName4, rName5, rName6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.*", map[string]string{
						"document_content_deletion": acctest.CtFalse,

						"condition.#": acctest.Ct1,
						"condition.0.condition_document_attribute_key":  "_document_title",
						"condition.0.operator":                          string(types.ConditionOperatorEquals),
						"condition.0.condition_on_value.#":              acctest.Ct1,
						"condition.0.condition_on_value.0.string_value": "foo2",

						"target.#":                               acctest.Ct1,
						"target.0.target_document_attribute_key": "_category",
						"target.0.target_document_attribute_value_deletion":       acctest.CtTrue,
						"target.0.target_document_attribute_value.#":              acctest.Ct1,
						"target.0.target_document_attribute_value.0.string_value": "bar2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.*", map[string]string{
						"document_content_deletion": acctest.CtFalse,

						"condition.#": acctest.Ct1,
						"condition.0.condition_document_attribute_key": "_created_at",
						"condition.0.operator":                         string(types.ConditionOperatorEquals),
						"condition.0.condition_on_value.#":             acctest.Ct1,
						"condition.0.condition_on_value.0.date_value":  "2006-01-02T15:04:05+00:00",

						"target.#":                               acctest.Ct1,
						"target.0.target_document_attribute_key": "_authors",
						"target.0.target_document_attribute_value_deletion":              acctest.CtFalse,
						"target.0.target_document_attribute_value.#":                     acctest.Ct1,
						"target.0.target_document_attribute_value.0.string_list_value.#": acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_document_enrichment_configuration.0.inline_configurations.*", map[string]string{
						"document_content_deletion": acctest.CtTrue,

						"condition.#": acctest.Ct1,
						"condition.0.condition_document_attribute_key": "_excerpt_page_number",
						"condition.0.operator":                         string(types.ConditionOperatorGreaterThan),
						"condition.0.condition_on_value.#":             acctest.Ct1,
						"condition.0.condition_on_value.0.long_value":  acctest.Ct3,

						"target.#":                               acctest.Ct1,
						"target.0.target_document_attribute_key": "_document_title",
						"target.0.target_document_attribute_value_deletion":       acctest.CtTrue,
						"target.0.target_document_attribute_value.#":              acctest.Ct1,
						"target.0.target_document_attribute_value.0.string_value": "baz",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_CustomDocumentEnrichmentConfiguration_ExtractionHookConfiguration_InvocationCondition(t *testing.T) {
	ctx := acctest.Context(t)
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
	rName8 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName9 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName10 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	originalOperator := string(types.ConditionOperatorEquals)
	updatedOperator := string(types.ConditionOperatorGreaterThan)

	originalStringValue := "original"
	updatedStringValue := "updated"

	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationInvocationCondition(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, originalOperator, originalStringValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.role_arn", "aws_iam_role.test_extraction_hook", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.lambda_arn", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.s3_bucket", "aws_s3_bucket.test_extraction_hook", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_document_attribute_key", "_excerpt_page_number"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.operator", originalOperator),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.0.long_value", acctest.Ct3),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.lambda_arn", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.s3_bucket", "aws_s3_bucket.test_extraction_hook", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.condition_document_attribute_key", "_document_title"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.operator", string(types.ConditionOperatorEquals)),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.0.string_value", originalStringValue),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationInvocationCondition(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, updatedOperator, updatedStringValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.role_arn", "aws_iam_role.test_extraction_hook", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.lambda_arn", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.s3_bucket", "aws_s3_bucket.test_extraction_hook", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_document_attribute_key", "_excerpt_page_number"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.operator", updatedOperator),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.0.long_value", acctest.Ct3),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.lambda_arn", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.s3_bucket", "aws_s3_bucket.test_extraction_hook", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.condition_document_attribute_key", "_document_title"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.operator", string(types.ConditionOperatorEquals)),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.0.string_value", updatedStringValue),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_CustomDocumentEnrichmentConfiguration_ExtractionHookConfiguration_RoleARN(t *testing.T) {
	ctx := acctest.Context(t)
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
	rName8 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName9 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName10 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName11 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationRoleARN(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, rName11, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.role_arn", "aws_iam_role.test_extraction_hook", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.lambda_arn", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.s3_bucket", "aws_s3_bucket.test_extraction_hook", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_document_attribute_key", "_excerpt_page_number"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.operator", string(types.ConditionOperatorEquals)),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.0.long_value", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationRoleARN(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, rName11, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.role_arn", "aws_iam_role.test_extraction_hook2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.lambda_arn", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.s3_bucket", "aws_s3_bucket.test_extraction_hook", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_document_attribute_key", "_excerpt_page_number"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.operator", string(types.ConditionOperatorEquals)),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.0.long_value", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_CustomDocumentEnrichmentConfiguration_ExtractionHookConfiguration_S3Bucket(t *testing.T) {
	ctx := acctest.Context(t)
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
	rName8 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName9 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName10 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName11 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationS3Bucket(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, rName11, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.role_arn", "aws_iam_role.test_extraction_hook", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.lambda_arn", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.s3_bucket", "aws_s3_bucket.test_extraction_hook", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_document_attribute_key", "_excerpt_page_number"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.operator", string(types.ConditionOperatorEquals)),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.0.long_value", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationS3Bucket(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, rName11, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.role_arn", "aws_iam_role.test_extraction_hook", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.lambda_arn", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.s3_bucket", "aws_s3_bucket.test_extraction_hook2", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_document_attribute_key", "_excerpt_page_number"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.operator", string(types.ConditionOperatorEquals)),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.0.long_value", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func TestAccKendraDataSource_CustomDocumentEnrichmentConfiguration_ExtractionHookConfiguration_LambdaARN(t *testing.T) {
	ctx := acctest.Context(t)
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
	rName8 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName9 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName10 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName11 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_kendra_data_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationLambdaARN(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, rName11, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.role_arn", "aws_iam_role.test_extraction_hook", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.lambda_arn", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.s3_bucket", "aws_s3_bucket.test_extraction_hook", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_document_attribute_key", "_excerpt_page_number"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.operator", string(types.ConditionOperatorEquals)),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.0.long_value", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationLambdaARN(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, rName11, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_data_source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.role_arn", "aws_iam_role.test_extraction_hook", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.lambda_arn", "aws_lambda_function.test2", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.s3_bucket", "aws_s3_bucket.test_extraction_hook", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_document_attribute_key", "_excerpt_page_number"),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.operator", string(types.ConditionOperatorEquals)),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.condition_on_value.0.long_value", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.DataSourceTypeS3)),
				),
			},
		},
	})
}

func testAccCheckDataSourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kendra_data_source" {
				continue
			}

			id, indexId, err := tfkendra.DataSourceParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}
			_, err = tfkendra.FindDataSourceByID(ctx, conn, id, indexId)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckDataSourceExists(ctx context.Context, name string) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraClient(ctx)

		_, err = tfkendra.FindDataSourceByID(ctx, conn, id, indexId)

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

func testAccDataSourceConfigS3Base(rName, rName2 string) string {
	// Kendra IAM policies: https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source" {
  name               = %[2]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "data_source_policy"

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
`, rName, rName2)
}

func testAccDataSourceConfigWebCrawlerBase(rName string) string {
	// Kendra IAM policies: https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html
	return fmt.Sprintf(`
resource "aws_iam_role" "test_data_source" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "data_source_policy"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
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
`, rName)
}

func testAccDataSourceConfigWebCrawlerSecretsBase(rName, rName2, rName3 string) string {
	// Kendra IAM policies: https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name                    = %[1]q
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "{\"hello\":\"world\"}"
}

resource "aws_secretsmanager_secret" "test2" {
  name                    = %[2]q
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test2.id
  secret_string = "{\"foo\":\"bar\"}"
}

resource "aws_iam_role" "test_data_source" {
  name               = %[3]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "data_source_policy"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "secretsmanager:GetSecretValue"
          ]
          Effect = "Allow"
          Resource = [
            aws_secretsmanager_secret.test.id,
            aws_secretsmanager_secret.test2.id,
          ]
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
`, rName, rName2, rName3)
}

func testAccDataSourceConfigExtractionHookBase(rName, rName2, rName3, rName4 string) string {
	// Kendra IAM policies: https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test_lambda" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_s3_bucket" "test_extraction_hook" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "test_lambda" {
  name = %[2]q

  assume_role_policy = data.aws_iam_policy_document.test_lambda.json
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[3]q
  role          = aws_iam_role.test_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}

resource "aws_iam_role" "test_extraction_hook" {
  name               = %[4]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "data_source_policy"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "s3:GetObject",
            "s3:PutObject",
          ]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test_extraction_hook.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test_extraction_hook.arn
        },
        {
          Action   = ["lambda:InvokeFunction"]
          Effect   = "Allow"
          Resource = aws_lambda_function.test.arn
        },
      ]
    })
  }
}
`, rName, rName2, rName3, rName4)
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
		testAccDataSourceConfigS3Base(rName4, rName5),
		fmt.Sprintf(`
locals {
  select_arn = %[3]q
}

resource "aws_iam_role" "test_data_source2" {
  name               = %[2]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "data_source_policy"

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
`, rName6, rName7, selectARN))
}

func testAccDataSourceConfig_schedule(rName, rName2, rName3, rName4, rName5, rName6, schedule string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn
  schedule = %[2]q

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`, rName6, schedule))
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

func testAccDataSourceConfig_configurationS3Bucket(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, selectBucket string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		fmt.Sprintf(`
locals {
  select_bucket = %[4]q
}

resource "aws_s3_bucket" "test2" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_iam_role" "test_data_source2" {
  name               = %[3]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "data_source_policy"

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
  role_arn = aws_iam_role.test_data_source2.arn

  configuration {
    s3_configuration {
      bucket_name = local.select_bucket == "first" ? aws_s3_bucket.test.id : aws_s3_bucket.test2.id
    }
  }
}
`, rName6, rName7, rName8, selectBucket))
}

func testAccDataSourceConfig_configurationS3AccessControlList(rName, rName2, rName3, rName4, rName5, rName6, selectKeyPath string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		fmt.Sprintf(`
locals {
  select_key_path = %[2]q
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
`, rName6, selectKeyPath))
}

func testAccDataSourceConfig_configurationS3DocumentsMetadataConfiguration(rName, rName2, rName3, rName4, rName5, rName6, s3Prefix string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id

      documents_metadata_configuration {
        s3_prefix = %[2]q
      }
    }
  }
}
`, rName6, s3Prefix))
}

func testAccDataSourceConfig_configurationS3ExclusionInclusionPatternsPrefixes1(rName, rName2, rName3, rName4, rName5, rName6 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		fmt.Sprintf(`
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
`, rName6))
}

func testAccDataSourceConfig_configurationS3ExclusionInclusionPatternsPrefixes2(rName, rName2, rName3, rName4, rName5, rName6 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		fmt.Sprintf(`
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
`, rName6))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsSeedURLs(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerBase(rName4),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName5))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsSeedURLs2(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerBase(rName4),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index",
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_faq"
          ]
        }
      }
    }
  }
}
`, rName5))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsWebCrawlerMode(rName, rName2, rName3, rName4, rName5, webCrawlerMode string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerBase(rName4),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      urls {
        seed_url_configuration {
          web_crawler_mode = %[2]q

          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName5, webCrawlerMode))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationAuthenticationConfigurationBasicHostPort(rName, rName2, rName3, rName4, rName5, rName6, rName7, host1 string, port1 int) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerSecretsBase(rName4, rName5, rName6),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test
  ]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      authentication_configuration {
        basic_authentication {
          credentials = aws_secretsmanager_secret.test.arn
          host        = %[2]q
          port        = %[3]d
        }
      }

      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName7, host1, port1))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationAuthenticationConfigurationBasicHostPort2(rName, rName2, rName3, rName4, rName5, rName6, rName7, host1 string, port1 int, host2 string, port2 int) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerSecretsBase(rName4, rName5, rName6),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test
  ]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      authentication_configuration {
        basic_authentication {
          credentials = aws_secretsmanager_secret.test.arn
          host        = %[2]q
          port        = %[3]d
        }

        basic_authentication {
          credentials = aws_secretsmanager_secret.test.arn
          host        = %[4]q
          port        = %[5]d
        }
      }

      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName7, host1, port1, host2, port2))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationAuthenticationConfigurationBasicCredentials(rName, rName2, rName3, rName4, rName5, rName6, rName7, selectCredentials string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerSecretsBase(rName4, rName5, rName6),
		fmt.Sprintf(`
locals {
  select_credentials = %[2]q
}

resource "aws_kendra_data_source" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_secretsmanager_secret_version.test2,
  ]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      authentication_configuration {
        basic_authentication {
          credentials = local.select_credentials == "first" ? aws_secretsmanager_secret.test.arn : aws_secretsmanager_secret.test2.arn
          host        = "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          port        = "123"
        }
      }

      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName7, selectCredentials))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationCrawlDepth(rName, rName2, rName3, rName4, rName5 string, crawlDepth int) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerBase(rName4),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      crawl_depth = %[2]d

      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName5, crawlDepth))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationMaxLinksPerPage(rName, rName2, rName3, rName4, rName5 string, maxLinksPerPage int) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerBase(rName4),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      max_links_per_page = %[2]d

      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName5, maxLinksPerPage))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationMaxURLsPerMinuteCrawlRate(rName, rName2, rName3, rName4, rName5 string, maxUrlsPerMinuteCrawlRate int) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerBase(rName4),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      max_urls_per_minute_crawl_rate = %[2]d

      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName5, maxUrlsPerMinuteCrawlRate))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationProxyConfigurationCredentials(rName, rName2, rName3, rName4, rName5, rName6, rName7, host1 string, port1 int) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerSecretsBase(rName4, rName5, rName6),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test
  ]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      proxy_configuration {
        credentials = aws_secretsmanager_secret.test.arn
        host        = %[2]q
        port        = %[3]d
      }

      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName7, host1, port1))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationProxyConfigurationHostPort(rName, rName2, rName3, rName4, rName5, rName6, rName7, host1 string, port1 int) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerSecretsBase(rName4, rName5, rName6),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      proxy_configuration {
        host = %[2]q
        port = %[3]d
      }

      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName7, host1, port1))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationURLExclusionInclusionPatterns(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerBase(rName4),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      url_exclusion_patterns = ["example"]
      url_inclusion_patterns = ["hello"]

      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName5))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationURLExclusionInclusionPatterns2(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerBase(rName4),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      url_exclusion_patterns = ["example2", "foo"]
      url_inclusion_patterns = ["hello2", "bar"]

      urls {
        seed_url_configuration {
          seed_urls = [
            "https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kendra_index"
          ]
        }
      }
    }
  }
}
`, rName5))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsSiteMapsConfiguration(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerBase(rName4),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      urls {
        site_maps_configuration {
          site_maps = [
            "https://registry.terraform.io/sitemap.xml"
          ]
        }
      }
    }
  }
}
`, rName5))
}

func testAccDataSourceConfig_configurationWebCrawlerConfigurationURLsSiteMapsConfiguration2(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigWebCrawlerBase(rName4),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "WEBCRAWLER"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    web_crawler_configuration {
      urls {
        site_maps_configuration {
          site_maps = [
            "https://registry.terraform.io/sitemap.xml",
            "https://www.terraform.io/sitemap.xml"
          ]
        }
      }
    }
  }
}
`, rName5))
}

func testAccDataSourceConfig_customDocumentEnrichmentConfigurationInlineConfigurations1(rName, rName2, rName3, rName4, rName5, rName6 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		fmt.Sprintf(`
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
`, rName6))
}

func testAccDataSourceConfig_customDocumentEnrichmentConfigurationInlineConfigurations2(rName, rName2, rName3, rName4, rName5, rName6 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		fmt.Sprintf(`
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
          date_value = "2006-01-02T15:04:05+00:00"
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
`, rName6))
}

func testAccDataSourceConfig_customDocumentEnrichmentConfigurationInlineConfigurations3(rName, rName2, rName3, rName4, rName5, rName6 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		fmt.Sprintf(`
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
          date_value = "2006-01-02T15:04:05+00:00"
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

    inline_configurations {
      document_content_deletion = true

      condition {
        condition_document_attribute_key = "_excerpt_page_number"
        operator                         = "GreaterThan"

        condition_on_value {
          long_value = 3
        }
      }

      target {
        target_document_attribute_key            = "_document_title"
        target_document_attribute_value_deletion = true

        target_document_attribute_value {
          string_value = "baz"
        }
      }
    }
  }
}
`, rName6))
}

func testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationInvocationCondition(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, preExtractionOperator, postExtractionStringValue string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		testAccDataSourceConfigExtractionHookBase(rName6, rName7, rName8, rName9),
		fmt.Sprintf(`
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
    role_arn = aws_iam_role.test_extraction_hook.arn

    post_extraction_hook_configuration {
      lambda_arn = aws_lambda_function.test.arn
      s3_bucket  = aws_s3_bucket.test_extraction_hook.id

      invocation_condition {
        condition_document_attribute_key = "_excerpt_page_number"
        operator                         = %[2]q

        condition_on_value {
          long_value = 3
        }
      }
    }

    pre_extraction_hook_configuration {
      lambda_arn = aws_lambda_function.test.arn
      s3_bucket  = aws_s3_bucket.test_extraction_hook.id

      invocation_condition {
        condition_document_attribute_key = "_document_title"
        operator                         = "Equals"

        condition_on_value {
          string_value = %[3]q
        }
      }
    }
  }
}
`, rName10, preExtractionOperator, postExtractionStringValue))
}

func testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationRoleARN(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, rName11, selectARN string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		testAccDataSourceConfigExtractionHookBase(rName6, rName7, rName8, rName9),
		fmt.Sprintf(`
locals {
  select_arn = %[3]q
}

resource "aws_iam_role" "test_extraction_hook2" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "data_source_policy"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "s3:GetObject",
            "s3:PutObject",
          ]
          Effect   = "Allow"
          Resource = "${aws_s3_bucket.test_extraction_hook.arn}/*"
        },
        {
          Action   = ["s3:ListBucket"]
          Effect   = "Allow"
          Resource = aws_s3_bucket.test_extraction_hook.arn
        },
        {
          Action   = ["lambda:InvokeFunction"]
          Effect   = "Allow"
          Resource = aws_lambda_function.test.arn
        },
      ]
    })
  }
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[2]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id
    }
  }

  custom_document_enrichment_configuration {
    role_arn = local.select_arn == "first" ? aws_iam_role.test_extraction_hook.arn : aws_iam_role.test_extraction_hook2.arn

    post_extraction_hook_configuration {
      lambda_arn = aws_lambda_function.test.arn
      s3_bucket  = aws_s3_bucket.test_extraction_hook.id

      invocation_condition {
        condition_document_attribute_key = "_excerpt_page_number"
        operator                         = "Equals"

        condition_on_value {
          long_value = 3
        }
      }
    }
  }
}
`, rName10, rName11, selectARN))
}

func testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationS3Bucket(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, rName11, selectBucket string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		testAccDataSourceConfigExtractionHookBase(rName6, rName7, rName8, rName9),
		fmt.Sprintf(`
locals {
  select_bucket = %[3]q
}

resource "aws_s3_bucket" "test_extraction_hook2" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[2]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id
    }
  }

  custom_document_enrichment_configuration {
    role_arn = aws_iam_role.test_extraction_hook.arn

    post_extraction_hook_configuration {
      lambda_arn = aws_lambda_function.test.arn
      s3_bucket  = local.select_bucket == "first" ? aws_s3_bucket.test_extraction_hook.id : aws_s3_bucket.test_extraction_hook2.id

      invocation_condition {
        condition_document_attribute_key = "_excerpt_page_number"
        operator                         = "Equals"

        condition_on_value {
          long_value = 3
        }
      }
    }
  }
}
`, rName10, rName11, selectBucket))
}

func testAccDataSourceConfig_customDocumentEnrichmentConfigurationExtractionHookConfigurationLambdaARN(rName, rName2, rName3, rName4, rName5, rName6, rName7, rName8, rName9, rName10, rName11, selectLambdaARN string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		testAccDataSourceConfigS3Base(rName4, rName5),
		testAccDataSourceConfigExtractionHookBase(rName6, rName7, rName8, rName9),
		fmt.Sprintf(`
locals {
  select_lambda_arn = %[3]q
}

resource "aws_lambda_function" "test2" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[2]q
  type     = "S3"
  role_arn = aws_iam_role.test_data_source.arn

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id
    }
  }

  custom_document_enrichment_configuration {
    role_arn = aws_iam_role.test_extraction_hook.arn

    post_extraction_hook_configuration {
      lambda_arn = local.select_lambda_arn == "first" ? aws_lambda_function.test.arn : aws_lambda_function.test2.arn
      s3_bucket  = aws_s3_bucket.test_extraction_hook.id

      invocation_condition {
        condition_document_attribute_key = "_excerpt_page_number"
        operator                         = "Equals"

        condition_on_value {
          long_value = 3
        }
      }
    }
  }
}
`, rName10, rName11, selectLambdaARN))
}
