// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketMetric_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_metric.test"
	metricName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_noFilter(rName, metricName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, metricName),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11813
// Disallow Empty filter block
func TestAccS3BucketMetric_withEmptyFilter(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_metric.test"
	metricName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_emptyFilter(rName, metricName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
				),
				ExpectError: regexache.MustCompile(`(?is)one of.*must be specified`),
			},
		},
	})
}

func TestAccS3BucketMetric_withFilterAccessPoint(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	baseAccessPointArn := generateARN("aws:s3", names.USEast1RegionID, "accesspoint")
	accessPoint := fmt.Sprintf("%s/ap-%d", baseAccessPointArn, rInt)
	accessPointUpdate := fmt.Sprintf("%s/ap-update-%d", baseAccessPointArn, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterAccessPoint(bucketName, metricName, accessPoint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPoint),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct0),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterAccessPoint(bucketName, metricName, accessPointUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPointUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct0),
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

func TestAccS3BucketMetric_withFilterAccessPointAndPrefixAndMultiTags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	baseAccessPointArn := generateARN("aws:s3", names.USEast1RegionID, "accesspoint")
	accessPoint := fmt.Sprintf("%s/ap-%d", baseAccessPointArn, rInt)
	accessPointUpdate := fmt.Sprintf("%s/ap-update-%d", baseAccessPointArn, rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterAccessPointAndPrefixAndMultipleTags(bucketName, metricName, accessPoint, prefix, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPoint),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterAccessPointAndPrefixAndMultipleTags(bucketName, metricName, accessPointUpdate, prefixUpdate, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPointUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
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

func TestAccS3BucketMetric_withFilterAccessPointAndPrefixAndSingleTag(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	baseAccessPointArn := generateARN("aws:s3", names.USEast1RegionID, "accesspoint")
	accessPoint := fmt.Sprintf("%s/ap-%d", baseAccessPointArn, rInt)
	accessPointUpdate := fmt.Sprintf("%s/ap-update-%d", baseAccessPointArn, rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterAccessPointAndPrefixAndSingleTag(bucketName, metricName, accessPoint, prefix, tag1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPoint),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterAccessPointAndPrefixAndSingleTag(bucketName, metricName, accessPointUpdate, prefixUpdate, tag1Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPointUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
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

func TestAccS3BucketMetric_withFilterAccessPointAndPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	baseAccessPointArn := generateARN("aws:s3", names.USEast1RegionID, "accesspoint")
	accessPoint := fmt.Sprintf("%s/ap-%d", baseAccessPointArn, rInt)
	accessPointUpdate := fmt.Sprintf("%s/ap-update-%d", baseAccessPointArn, rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterAccessPointAndPrefix(bucketName, metricName, accessPoint, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPoint),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct0),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterAccessPointAndPrefix(bucketName, metricName, accessPointUpdate, prefixUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPointUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct0),
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

func TestAccS3BucketMetric_withFilterAccessPointAndMultipleTags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	baseAccessPointArn := generateARN("aws:s3", names.USEast1RegionID, "accesspoint")
	accessPoint := fmt.Sprintf("%s/ap-%d", baseAccessPointArn, rInt)
	accessPointUpdate := fmt.Sprintf("%s/ap-update-%d", baseAccessPointArn, rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterAccessPointAndMultipleTags(bucketName, metricName, accessPoint, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPoint),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterAccessPointAndMultipleTags(bucketName, metricName, accessPointUpdate, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPointUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
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

func TestAccS3BucketMetric_withFilterAccessPointAndSingleTag(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	baseAccessPointArn := generateARN("aws:s3", names.USEast1RegionID, "accesspoint")
	accessPoint := fmt.Sprintf("%s/ap-%d", baseAccessPointArn, rInt)
	accessPointUpdate := fmt.Sprintf("%s/ap-update-%d", baseAccessPointArn, rInt)
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterAccessPointAndSingleTag(bucketName, metricName, accessPoint, tag1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPoint),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterAccessPointAndSingleTag(bucketName, metricName, accessPointUpdate, tag1Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.access_point", accessPointUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
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

func TestAccS3BucketMetric_withFilterPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterPrefix(bucketName, metricName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct0),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterPrefix(bucketName, metricName, prefixUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct0),
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

func TestAccS3BucketMetric_withFilterPrefixAndMultipleTags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterPrefixAndMultipleTags(bucketName, metricName, prefix, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterPrefixAndMultipleTags(bucketName, metricName, prefixUpdate, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
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

func TestAccS3BucketMetric_withFilterPrefixAndSingleTag(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterPrefixAndSingleTag(bucketName, metricName, prefix, tag1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterPrefixAndSingleTag(bucketName, metricName, prefixUpdate, tag1Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
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

func TestAccS3BucketMetric_withFilterMultipleTags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterMultipleTags(bucketName, metricName, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterMultipleTags(bucketName, metricName, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
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

func TestAccS3BucketMetric_withFilterSingleTag(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"
	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricConfig_filterSingleTag(bucketName, metricName, tag1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccBucketMetricConfig_filterSingleTag(bucketName, metricName, tag1Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
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

func TestAccS3BucketMetric_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	metricName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetricDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketMetricConfig_directoryBucket(rName, metricName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketMetricDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_metric" {
				continue
			}

			bucket, name, err := tfs3.BucketMetricParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindMetricsConfiguration(ctx, conn, bucket, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Metric %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketMetricsExistsConfig(ctx context.Context, n string, v *types.MetricsConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		bucket, name, err := tfs3.BucketMetricParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		output, err := tfs3.FindMetricsConfiguration(ctx, conn, bucket, name)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func generateARN(partition, region, resource string) string {
	randAccount := fmt.Sprintf("%012d", rand.Intn(1e12))
	return "arn:" + partition + ":" + region + ":" + randAccount + ":" + resource
}

func testAccBucketMetricConfig_base(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}
`, bucketName)
}

func testAccBucketMetricConfig_emptyFilter(bucketName, metricName string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {}
}
`, metricName))
}

func testAccBucketMetricConfig_filterAccessPoint(bucketName, metricName, accessPoint string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    access_point = %[2]q
  }
}
`, metricName, accessPoint))
}

func testAccBucketMetricConfig_filterAccessPointAndPrefixAndMultipleTags(bucketName, metricName, accessPoint, prefix, tag1, tag2 string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    access_point = %[2]q
    prefix       = %[3]q

    tags = {
      "tag1" = %[4]q
      "tag2" = %[5]q
    }
  }
}
`, metricName, accessPoint, prefix, tag1, tag2))
}

func testAccBucketMetricConfig_filterAccessPointAndPrefixAndSingleTag(bucketName, metricName, accessPoint, prefix, tag string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    access_point = %[2]q
    prefix       = %[3]q

    tags = {
      "tag1" = %[4]q
    }
  }
}
`, metricName, accessPoint, prefix, tag))
}

func testAccBucketMetricConfig_filterAccessPointAndPrefix(bucketName, metricName, accessPoint, prefix string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    access_point = %[2]q
    prefix       = %[3]q
  }
}
`, metricName, accessPoint, prefix))
}

func testAccBucketMetricConfig_filterAccessPointAndMultipleTags(bucketName, metricName, accessPoint, tag1, tag2 string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    access_point = %[2]q

    tags = {
      "tag1" = %[3]q
      "tag2" = %[4]q
    }
  }
}
`, metricName, accessPoint, tag1, tag2))
}

func testAccBucketMetricConfig_filterAccessPointAndSingleTag(bucketName, metricName, accessPoint, tag string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    access_point = %[2]q

    tags = {
      "tag1" = %[3]q
    }
  }
}
`, metricName, accessPoint, tag))
}

func testAccBucketMetricConfig_filterPrefix(bucketName, metricName, prefix string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    prefix = %[2]q
  }
}
`, metricName, prefix))
}

func testAccBucketMetricConfig_filterPrefixAndMultipleTags(bucketName, metricName, prefix, tag1, tag2 string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    prefix = %[2]q

    tags = {
      "tag1" = %[3]q
      "tag2" = %[4]q
    }
  }
}
`, metricName, prefix, tag1, tag2))
}

func testAccBucketMetricConfig_filterPrefixAndSingleTag(bucketName, metricName, prefix, tag string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    prefix = %[2]q

    tags = {
      "tag1" = %[3]q
    }
  }
}
`, metricName, prefix, tag))
}

func testAccBucketMetricConfig_filterMultipleTags(bucketName, metricName, tag1, tag2 string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    tags = {
      "tag1" = %[2]q
      "tag2" = %[3]q
    }
  }
}
`, metricName, tag1, tag2))
}

func testAccBucketMetricConfig_filterSingleTag(bucketName, metricName, tag string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q

  filter {
    tags = {
      "tag1" = %[2]q
    }
  }
}
`, metricName, tag))
}

func testAccBucketMetricConfig_noFilter(bucketName, metricName string) string {
	return acctest.ConfigCompose(testAccBucketMetricConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = %[1]q
}
`, metricName))
}

func testAccBucketMetricConfig_directoryBucket(bucketName, metricName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q
}
`, metricName))
}
