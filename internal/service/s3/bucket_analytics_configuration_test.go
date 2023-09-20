// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/go-cmp/cmp"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestExpandAnalyticsFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := map[string]struct {
		Input    []interface{}
		Expected types.AnalyticsFilter
	}{
		"prefix only": {
			Input: []interface{}{
				map[string]interface{}{
					"prefix": "prefix/",
				},
			},
			Expected: &types.AnalyticsFilterMemberPrefix{
				Value: "prefix/",
			},
		},
		"prefix and single tag": {
			Input: []interface{}{
				map[string]interface{}{
					"prefix": "prefix/",
					"tags": map[string]interface{}{
						"tag1key": "tag1value",
					},
				},
			},
			Expected: &types.AnalyticsFilterMemberAnd{
				Value: types.AnalyticsAndOperator{
					Prefix: aws.String("prefix/"),
					Tags: []types.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
					},
				},
			},
		},
		"prefix and multiple tags": {
			Input: []interface{}{map[string]interface{}{
				"prefix": "prefix/",
				"tags": map[string]interface{}{
					"tag1key": "tag1value",
					"tag2key": "tag2value",
				},
			},
			},
			Expected: &types.AnalyticsFilterMemberAnd{
				Value: types.AnalyticsAndOperator{
					Prefix: aws.String("prefix/"),
					Tags: []types.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
						{
							Key:   aws.String("tag2key"),
							Value: aws.String("tag2value"),
						},
					},
				},
			},
		},
		"single tag only": {
			Input: []interface{}{
				map[string]interface{}{
					"tags": map[string]interface{}{
						"tag1key": "tag1value",
					},
				},
			},
			Expected: &types.AnalyticsFilterMemberTag{
				Value: types.Tag{
					Key:   aws.String("tag1key"),
					Value: aws.String("tag1value"),
				},
			},
		},
		"multiple tags only": {
			Input: []interface{}{
				map[string]interface{}{
					"tags": map[string]interface{}{
						"tag1key": "tag1value",
						"tag2key": "tag2value",
					},
				},
			},
			Expected: &types.AnalyticsFilterMemberAnd{
				Value: types.AnalyticsAndOperator{
					Tags: []types.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
						{
							Key:   aws.String("tag2key"),
							Value: aws.String("tag2value"),
						},
					},
				},
			},
		},
	}

	for k, tc := range testCases {
		value := tfs3.ExpandAnalyticsFilter(ctx, tc.Input[0].(map[string]interface{}))

		if value == nil {
			if tc.Expected == nil {
				continue
			}

			t.Errorf("Case %q: Got nil\nExpected:\n%v", k, tc.Expected)
		}

		if tc.Expected == nil {
			t.Errorf("Case %q: Got: %v\nExpected: nil", k, value)
		}

		// Sort tags by key for consistency
		// if value.And != nil && value.And.Tags != nil {
		// 	sort.Slice(value.And.Tags, func(i, j int) bool {
		// 		return *value.And.Tags[i].Key < *value.And.Tags[j].Key
		// 	})
		// }

		if diff := cmp.Diff(value, tc.Expected); diff != "" {
			t.Errorf("unexpected AnalyticsFilter diff (+wanted, -got): %s", diff)
		}
	}
}

func TestExpandStorageClassAnalysis(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		Input    []interface{}
		Expected *types.StorageClassAnalysis
	}{
		"nil input": {
			Input:    nil,
			Expected: &types.StorageClassAnalysis{},
		},
		"empty input": {
			Input:    []interface{}{},
			Expected: &types.StorageClassAnalysis{},
		},
		"nil array": {
			Input: []interface{}{
				nil,
			},
			Expected: &types.StorageClassAnalysis{},
		},
		"empty data_export": {
			Input: []interface{}{
				map[string]interface{}{
					"data_export": []interface{}{},
				},
			},
			Expected: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{},
			},
		},
		"data_export complete": {
			Input: []interface{}{
				map[string]interface{}{
					"data_export": []interface{}{
						map[string]interface{}{
							"output_schema_version": types.StorageClassAnalysisSchemaVersionV1,
							"destination":           []interface{}{},
						},
					},
				},
			},
			Expected: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{
					OutputSchemaVersion: types.StorageClassAnalysisSchemaVersionV1,
					Destination:         &types.AnalyticsExportDestination{},
				},
			},
		},
		"empty s3_bucket_destination": {
			Input: []interface{}{
				map[string]interface{}{
					"data_export": []interface{}{
						map[string]interface{}{
							"destination": []interface{}{
								map[string]interface{}{
									"s3_bucket_destination": []interface{}{},
								},
							},
						},
					},
				},
			},
			Expected: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{
					Destination: &types.AnalyticsExportDestination{
						S3BucketDestination: &types.AnalyticsS3BucketDestination{},
					},
				},
			},
		},
		"s3_bucket_destination complete": {
			Input: []interface{}{
				map[string]interface{}{
					"data_export": []interface{}{
						map[string]interface{}{
							"destination": []interface{}{
								map[string]interface{}{
									"s3_bucket_destination": []interface{}{
										map[string]interface{}{
											"bucket_arn":        "arn:aws:s3", //lintignore:AWSAT005
											"bucket_account_id": "1234567890",
											"format":            types.AnalyticsS3ExportFileFormatCsv,
											"prefix":            "prefix/",
										},
									},
								},
							},
						},
					},
				},
			},
			Expected: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{
					Destination: &types.AnalyticsExportDestination{
						S3BucketDestination: &types.AnalyticsS3BucketDestination{
							Bucket:          aws.String("arn:aws:s3"), //lintignore:AWSAT005
							BucketAccountId: aws.String("1234567890"),
							Format:          types.AnalyticsS3ExportFileFormatCsv,
							Prefix:          aws.String("prefix/"),
						},
					},
				},
			},
		},
		"s3_bucket_destination required": {
			Input: []interface{}{
				map[string]interface{}{
					"data_export": []interface{}{
						map[string]interface{}{
							"destination": []interface{}{
								map[string]interface{}{
									"s3_bucket_destination": []interface{}{
										map[string]interface{}{
											"bucket_arn": "arn:aws:s3", //lintignore:AWSAT005
											"format":     types.AnalyticsS3ExportFileFormatCsv,
										},
									},
								},
							},
						},
					},
				},
			},
			Expected: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{
					Destination: &types.AnalyticsExportDestination{
						S3BucketDestination: &types.AnalyticsS3BucketDestination{
							Bucket:          aws.String("arn:aws:s3"), //lintignore:AWSAT005
							BucketAccountId: nil,
							Format:          types.AnalyticsS3ExportFileFormatCsv,
							Prefix:          nil,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		value := tfs3.ExpandStorageClassAnalysis(tc.Input)

		if diff := cmp.Diff(value, tc.Expected); diff != "" {
			t.Errorf("unexpected StorageClassAnalysis diff (+wanted, -got): %s", diff)
		}
	}
}

func TestFlattenAnalyticsFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := map[string]struct {
		Input    types.AnalyticsFilter
		Expected []map[string]interface{}
	}{
		"nil input": {
			Input:    nil,
			Expected: nil,
		},
		"empty input": {
			Input:    &types.AnalyticsFilterMemberAnd{},
			Expected: nil,
		},
		"prefix only": {
			Input: &types.AnalyticsFilterMemberPrefix{
				Value: "prefix/",
			},
			Expected: []map[string]interface{}{
				{
					"prefix": "prefix/",
				},
			},
		},
		"prefix and single tag": {
			Input: &types.AnalyticsFilterMemberAnd{
				Value: types.AnalyticsAndOperator{
					Prefix: aws.String("prefix/"),
					Tags: []types.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
					},
				},
			},
			Expected: []map[string]interface{}{
				{
					"prefix": "prefix/",
					"tags": map[string]string{
						"tag1key": "tag1value",
					},
				},
			},
		},
		"prefix and multiple tags": {
			Input: &types.AnalyticsFilterMemberAnd{
				Value: types.AnalyticsAndOperator{
					Prefix: aws.String("prefix/"),
					Tags: []types.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
						{
							Key:   aws.String("tag2key"),
							Value: aws.String("tag2value"),
						},
					},
				},
			},
			Expected: []map[string]interface{}{
				{
					"prefix": "prefix/",
					"tags": map[string]string{
						"tag1key": "tag1value",
						"tag2key": "tag2value",
					},
				},
			},
		},
		"single tag only": {
			Input: &types.AnalyticsFilterMemberTag{
				Value: types.Tag{
					Key:   aws.String("tag1key"),
					Value: aws.String("tag1value"),
				},
			},
			Expected: []map[string]interface{}{
				{
					"tags": map[string]string{
						"tag1key": "tag1value",
					},
				},
			},
		},
		"multiple tags only": {
			Input: &types.AnalyticsFilterMemberAnd{
				Value: types.AnalyticsAndOperator{
					Tags: []types.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
						{
							Key:   aws.String("tag2key"),
							Value: aws.String("tag2value"),
						},
					},
				},
			},
			Expected: []map[string]interface{}{
				{
					"tags": map[string]string{
						"tag1key": "tag1value",
						"tag2key": "tag2value",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		value := tfs3.FlattenAnalyticsFilter(ctx, tc.Input)

		if diff := cmp.Diff(value, tc.Expected); diff != "" {
			t.Errorf("unexpected AnalyticsFilter diff (+wanted, -got): %s", diff)
		}
	}
}

func TestFlattenStorageClassAnalysis(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		Input    *types.StorageClassAnalysis
		Expected []map[string]interface{}
	}{
		"nil value": {
			Input:    nil,
			Expected: []map[string]interface{}{},
		},
		"empty root": {
			Input:    &types.StorageClassAnalysis{},
			Expected: []map[string]interface{}{},
		},
		"empty data_export": {
			Input: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{},
			},
			Expected: []map[string]interface{}{
				{
					"data_export": []interface{}{
						map[string]interface{}{},
					},
				},
			},
		},
		"data_export complete": {
			Input: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{
					OutputSchemaVersion: types.StorageClassAnalysisSchemaVersionV1,
					Destination:         &types.AnalyticsExportDestination{},
				},
			},
			Expected: []map[string]interface{}{
				{
					"data_export": []interface{}{
						map[string]interface{}{
							"output_schema_version": types.StorageClassAnalysisSchemaVersionV1,
							"destination":           []interface{}{},
						},
					},
				},
			},
		},
		"s3_bucket_destination required": {
			Input: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{
					Destination: &types.AnalyticsExportDestination{
						S3BucketDestination: &types.AnalyticsS3BucketDestination{
							Bucket: aws.String("arn:aws:s3"), //lintignore:AWSAT005
							Format: types.AnalyticsS3ExportFileFormatCsv,
						},
					},
				},
			},
			Expected: []map[string]interface{}{
				{
					"data_export": []interface{}{
						map[string]interface{}{
							"destination": []interface{}{
								map[string]interface{}{
									"s3_bucket_destination": []interface{}{
										map[string]interface{}{
											"bucket_arn": "arn:aws:s3", //lintignore:AWSAT005
											"format":     types.AnalyticsS3ExportFileFormatCsv,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"s3_bucket_destination complete": {
			Input: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{
					Destination: &types.AnalyticsExportDestination{
						S3BucketDestination: &types.AnalyticsS3BucketDestination{
							Bucket:          aws.String("arn:aws:s3"), //lintignore:AWSAT005
							BucketAccountId: aws.String("1234567890"),
							Format:          types.AnalyticsS3ExportFileFormatCsv,
							Prefix:          aws.String("prefix/"),
						},
					},
				},
			},
			Expected: []map[string]interface{}{
				{
					"data_export": []interface{}{
						map[string]interface{}{
							"destination": []interface{}{
								map[string]interface{}{
									"s3_bucket_destination": []interface{}{
										map[string]interface{}{
											"bucket_arn":        "arn:aws:s3", //lintignore:AWSAT005
											"bucket_account_id": "1234567890",
											"format":            types.AnalyticsS3ExportFileFormatCsv,
											"prefix":            "prefix/",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		value := tfs3.FlattenStorageClassAnalysis(tc.Input)

		if diff := cmp.Diff(value, tc.Expected); diff != "" {
			t.Errorf("unexpected StorageClassAnalysis diff (+wanted, -got): %s", diff)
		}
	}
}

func TestAccS3BucketAnalyticsConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
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

func TestAccS3BucketAnalyticsConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketAnalyticsConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketAnalyticsConfiguration_updateBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	originalACName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedACName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_basic(originalACName, originalBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "name", originalACName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_basic(updatedACName, originalBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "name", updatedACName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_update(updatedACName, originalBucketName, updatedBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "name", updatedACName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test_2", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketAnalyticsConfigurationConfig_emptyFilter(rName, rName),
				ExpectError: regexache.MustCompile(`one of .* must be specified`),
			},
		},
	})
}

func TestAccS3BucketAnalyticsConfiguration_WithFilter_prefix(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterPrefix(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "0"),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterPrefix(rName, rName, prefixUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "0"),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_singleTag(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterSingleTag(rName, rName, tag1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterSingleTag(rName, rName, tag1Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1Update),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_multipleTags(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterMultipleTags(rName, rName, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterMultipleTags(rName, rName, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1Update),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2Update),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_prefixAndTags(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterPrefixAndTags(rName, rName, prefix, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterPrefixAndTags(rName, rName, prefixUpdate, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1Update),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2Update),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_remove(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterPrefix(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
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

func TestAccS3BucketAnalyticsConfiguration_WithStorageClassAnalysis_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketAnalyticsConfigurationConfig_emptyStorageClassAnalysis(rName, rName),
				ExpectError: regexache.MustCompile(`running pre-apply refresh`),
			},
		},
	})
}

func TestAccS3BucketAnalyticsConfiguration_WithStorageClassAnalysis_default(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_defaultStorageClassAnalysis(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.output_schema_version", "V_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.format", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.bucket_arn", "aws_s3_bucket.destination", "arn"),
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

func TestAccS3BucketAnalyticsConfiguration_WithStorageClassAnalysis_full(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rInt := sdkacctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_fullStorageClassAnalysis(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.output_schema_version", "V_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.format", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.bucket_arn", "aws_s3_bucket.destination", "arn"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.prefix", prefix),
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

func testAccCheckBucketAnalyticsConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_analytics_configuration" {
				continue
			}

			bucket, _, err := tfs3.BucketAnalyticsConfigurationParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindAnalyticsConfiguration(ctx, conn, bucket)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Analytics Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketAnalyticsConfigurationExists(ctx context.Context, n string, v *types.AnalyticsConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		bucket, _, err := tfs3.BucketAnalyticsConfigurationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		output, err := tfs3.FindAnalyticsConfiguration(ctx, conn, bucket)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccBucketAnalyticsConfigurationConfig_basic(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_update(name, originalBucket, updatedBucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test_2.bucket
  name   = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}

resource "aws_s3_bucket" "test_2" {
  bucket = %[3]q
}
`, name, originalBucket, updatedBucket)
}

func testAccBucketAnalyticsConfigurationConfig_emptyFilter(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_filterPrefix(name, bucket, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    prefix = %[2]q
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[3]q
}
`, name, prefix, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_filterSingleTag(name, bucket, tag string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    tags = {
      "tag1" = %[2]q
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[3]q
}
`, name, tag, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_filterMultipleTags(name, bucket, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    tags = {
      "tag1" = %[2]q
      "tag2" = %[3]q
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[4]q
}
`, name, tag1, tag2, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_filterPrefixAndTags(name, bucket, prefix, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    prefix = %[2]q

    tags = {
      "tag1" = %[3]q
      "tag2" = %[4]q
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[5]q
}
`, name, prefix, tag1, tag2, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_emptyStorageClassAnalysis(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  storage_class_analysis {
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_defaultStorageClassAnalysis(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  storage_class_analysis {
    data_export {
      destination {
        s3_bucket_destination {
          bucket_arn = aws_s3_bucket.destination.arn
        }
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}

resource "aws_s3_bucket" "destination" {
  bucket = "%[2]s-destination"
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_fullStorageClassAnalysis(name, bucket, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  storage_class_analysis {
    data_export {
      output_schema_version = "V_1"

      destination {
        s3_bucket_destination {
          format     = "CSV"
          bucket_arn = aws_s3_bucket.destination.arn
          prefix     = %[2]q
        }
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[3]q
}

resource "aws_s3_bucket" "destination" {
  bucket = "%[3]s-destination"
}
`, name, prefix, bucket)
}
