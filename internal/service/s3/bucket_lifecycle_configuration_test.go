// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func TestAccS3BucketLifecycleConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.days": "365",
						"filter.#":          acctest.Ct1,
						"filter.0.prefix":   "",
						names.AttrID:        rName,
						names.AttrStatus:    tfs3.LifecycleRuleStatusEnabled,
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

func TestAccS3BucketLifecycleConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketLifecycleConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_filterWithPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	dateUpdated := time.Date(currTime.Year()+1, currTime.Month(), currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_basicUpdate(rName, date, "logs/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.date": date,
						"filter.#":          acctest.Ct1,
						"filter.0.prefix":   "logs/",
						names.AttrID:        rName,
						names.AttrStatus:    tfs3.LifecycleRuleStatusEnabled,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_basicUpdate(rName, dateUpdated, "tmp/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.date": dateUpdated,
						"filter.#":          acctest.Ct1,
						"filter.0.prefix":   "tmp/",
						names.AttrID:        rName,
						names.AttrStatus:    tfs3.LifecycleRuleStatusEnabled,
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

func TestAccS3BucketLifecycleConfiguration_Filter_ObjectSizeGreaterThan(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThan(rName, date, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":                      acctest.Ct1,
						"expiration.0.date":                 date,
						"filter.#":                          acctest.Ct1,
						"filter.0.object_size_greater_than": "100",
						names.AttrID:                        rName,
						names.AttrStatus:                    tfs3.LifecycleRuleStatusEnabled,
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

func TestAccS3BucketLifecycleConfiguration_Filter_ObjectSizeGreaterThanZero(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThan(rName, date, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":                      acctest.Ct1,
						"expiration.0.date":                 date,
						"filter.#":                          acctest.Ct1,
						"filter.0.object_size_greater_than": acctest.Ct0,
						names.AttrID:                        rName,
						names.AttrStatus:                    tfs3.LifecycleRuleStatusEnabled,
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

func TestAccS3BucketLifecycleConfiguration_Filter_ObjectSizeLessThan(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeLessThan(rName, date, 500),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":                   acctest.Ct1,
						"expiration.0.date":              date,
						"filter.#":                       acctest.Ct1,
						"filter.0.object_size_less_than": "500",
						names.AttrID:                     rName,
						names.AttrStatus:                 tfs3.LifecycleRuleStatusEnabled,
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

func TestAccS3BucketLifecycleConfiguration_Filter_ObjectSizeRange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeRange(rName, date, 500, 64000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.date": date,
						"filter.#":          acctest.Ct1,
						"filter.0.and.#":    acctest.Ct1,
						"filter.0.and.0.object_size_greater_than": "500",
						"filter.0.and.0.object_size_less_than":    "64000",
						names.AttrID:                              rName,
						names.AttrStatus:                          tfs3.LifecycleRuleStatusEnabled,
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

func TestAccS3BucketLifecycleConfiguration_Filter_ObjectSizeRangeAndPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeRangeAndPrefix(rName, date, 500, 64000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.date": date,
						"filter.#":          acctest.Ct1,
						"filter.0.and.#":    acctest.Ct1,
						"filter.0.and.0.object_size_greater_than": "500",
						"filter.0.and.0.object_size_less_than":    "64000",
						"filter.0.and.0.prefix":                   rName,
						names.AttrID:                              rName,
						names.AttrStatus:                          tfs3.LifecycleRuleStatusEnabled,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.transition.0.days",
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_disableRule(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_basicStatus(rName, tfs3.LifecycleRuleStatusEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_basicStatus(rName, tfs3.LifecycleRuleStatusDisabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrStatus: tfs3.LifecycleRuleStatusDisabled,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_basicStatus(rName, tfs3.LifecycleRuleStatusEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrStatus: tfs3.LifecycleRuleStatusEnabled,
					}),
				),
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_multipleRules(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	date := time.Now()
	expirationDate := time.Date(date.Year(), date.Month(), date.Day()+14, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_multipleRules(rName, expirationDate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:            "log",
						"expiration.#":          acctest.Ct1,
						"expiration.0.days":     "90",
						"filter.#":              acctest.Ct1,
						"filter.0.and.#":        acctest.Ct1,
						"filter.0.and.0.prefix": "log/",
						"filter.0.and.0.tags.%": acctest.Ct2,
						names.AttrStatus:        tfs3.LifecycleRuleStatusEnabled,
						"transition.#":          acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.transition.*", map[string]string{
						"days":                 "30",
						names.AttrStorageClass: string(types.StorageClassStandardIa),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.transition.*", map[string]string{
						"days":                 "60",
						names.AttrStorageClass: string(types.StorageClassGlacier),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:        "tmp",
						"expiration.#":      acctest.Ct1,
						"expiration.0.date": expirationDate,
						"filter.#":          acctest.Ct1,
						"filter.0.prefix":   "tmp/",
						names.AttrStatus:    tfs3.LifecycleRuleStatusEnabled,
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23730
func TestAccS3BucketLifecycleConfiguration_multipleRules_noFilterOrPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_multipleRulesNoFilterOrPrefix(rName, tfs3.LifecycleRuleStatusEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.1.filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.1.filter.0.prefix", ""),
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

func TestAccS3BucketLifecycleConfiguration_nonCurrentVersionExpiration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_nonCurrentVersionExpiration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"noncurrent_version_expiration.#":                 acctest.Ct1,
						"noncurrent_version_expiration.0.noncurrent_days": "90",
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

func TestAccS3BucketLifecycleConfiguration_nonCurrentVersionTransition(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_nonCurrentVersionTransition(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"noncurrent_version_transition.#": acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.noncurrent_version_transition.*", map[string]string{
						"noncurrent_days":      "30",
						names.AttrStorageClass: string(types.StorageClassStandardIa),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.noncurrent_version_transition.*", map[string]string{
						"noncurrent_days":      "60",
						names.AttrStorageClass: string(types.StorageClassGlacier),
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

// Ensure backwards compatible with now-deprecated "prefix" configuration
func TestAccS3BucketLifecycleConfiguration_prefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_basicPrefix(rName, "path1/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.days": "365",
						names.AttrID:        rName,
						names.AttrPrefix:    "path1/",
						names.AttrStatus:    tfs3.LifecycleRuleStatusEnabled,
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23239
func TestAccS3BucketLifecycleConfiguration_Filter_Tag(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterTag(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":         acctest.Ct1,
						"expiration.0.days":    "365",
						names.AttrID:           rName,
						"filter.#":             acctest.Ct1,
						"filter.0.tag.#":       acctest.Ct1,
						"filter.0.tag.0.key":   acctest.CtKey1,
						"filter.0.tag.0.value": acctest.CtValue1,
						names.AttrStatus:       tfs3.LifecycleRuleStatusEnabled,
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

func TestAccS3BucketLifecycleConfiguration_RuleExpiration_expireMarkerOnly(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationExpiredDeleteMarker(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#": acctest.Ct1,
						"expiration.0.expired_object_delete_marker": acctest.CtTrue,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationExpiredDeleteMarker(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#": acctest.Ct1,
						"expiration.0.expired_object_delete_marker": acctest.CtFalse,
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11420
func TestAccS3BucketLifecycleConfiguration_RuleExpiration_emptyBlock(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationEmptyBlock(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#": acctest.Ct1,
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15138
func TestAccS3BucketLifecycleConfiguration_ruleAbortIncompleteMultipartUpload(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleAbortIncompleteMultipartUpload(rName, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"abort_incomplete_multipart_upload.#":                       acctest.Ct1,
						"abort_incomplete_multipart_upload.0.days_after_initiation": "7",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleAbortIncompleteMultipartUpload(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"abort_incomplete_multipart_upload.#":                       acctest.Ct1,
						"abort_incomplete_multipart_upload.0.days_after_initiation": "5",
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

// TestAccS3BucketLifecycleConfiguration_TransitionDate_standardIa validates the change to address
// https://github.com/hashicorp/terraform-provider-aws/issues/23117
// does not introduce a regression.
func TestAccS3BucketLifecycleConfiguration_TransitionDate_standardIa(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_dateTransition(rName, date, string(types.TransitionStorageClassStandardIa)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.transition.0.days",
				},
			},
		},
	})
}

// TestAccS3BucketLifecycleConfiguration_TransitionDate_intelligentTiering validates the change to address
// https://github.com/hashicorp/terraform-provider-aws/issues/23117
// does not introduce a regression.
func TestAccS3BucketLifecycleConfiguration_TransitionDate_intelligentTiering(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_dateTransition(rName, date, string(types.StorageClassIntelligentTiering)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.transition.0.days",
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23117
func TestAccS3BucketLifecycleConfiguration_TransitionStorageClassOnly_intelligentTiering(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_transitionStorageClassOnly(rName, string(types.StorageClassIntelligentTiering)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.0.transition.*", map[string]string{
						"days":                 acctest.Ct0,
						"date":                 "",
						names.AttrStorageClass: string(types.StorageClassIntelligentTiering),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.transition.0.days",
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23117
func TestAccS3BucketLifecycleConfiguration_TransitionZeroDays_intelligentTiering(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_zeroDaysTransition(rName, string(types.StorageClassIntelligentTiering)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
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

func TestAccS3BucketLifecycleConfiguration_TransitionUpdateBetweenDaysAndDate_intelligentTiering(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_zeroDaysTransition(rName, string(types.StorageClassIntelligentTiering)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_dateTransition(rName, date, string(types.StorageClassIntelligentTiering)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.transition.0.days",
				},
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_zeroDaysTransition(rName, string(types.StorageClassIntelligentTiering)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23228
func TestAccS3BucketLifecycleConfiguration_EmptyFilter_NonCurrentVersions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
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

func TestAccS3BucketLifecycleConfiguration_migrate_noChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_lifecycleExpireMarker(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.days", acctest.Ct0),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", acctest.CtTrue),
				),
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_migrateNoChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.status", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.expiration.0.days", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.0.expiration.0.expired_object_delete_marker", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_migrate_withChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_lifecycleExpireMarker(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.days", acctest.Ct0),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", acctest.CtTrue),
				),
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_migrateChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.status", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.expiration.0.days", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.0.expiration.0.expired_object_delete_marker", acctest.CtFalse),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23884
func TestAccS3BucketLifecycleConfiguration_Update_filterWithAndToFilterWithPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThanAndPrefix(rName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.and.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.and.0.object_size_greater_than", "300"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.and.0.prefix", "prefix1"),
				),
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterPrefix(rName, "prefix2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.and.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.prefix", "prefix2"),
				),
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketLifecycleConfigurationConfig_directoryBucket(rName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketLifecycleConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_lifecycle_configuration" {
				continue
			}

			bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindLifecycleRules(ctx, conn, bucket, expectedBucketOwner)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Lifecycle Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketLifecycleConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tfs3.FindLifecycleRules(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketLifecycleConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id     = %[1]q
    status = "Enabled"

    expiration {
      days = 365
    }
  }
}
`, rName)
}

func testAccBucketLifecycleConfigurationConfig_multipleRulesNoFilterOrPrefix(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id     = "%[1]s-1"
    status = %[2]q

    expiration {
      days = 365
    }
  }

  rule {
    id     = "%[1]s-2"
    status = %[2]q

    expiration {
      expired_object_delete_marker = true
    }
  }
}
`, rName, status)
}

func testAccBucketLifecycleConfigurationConfig_basicStatus(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id     = %[1]q
    status = %[2]q

    expiration {
      days = 365
    }
  }
}
`, rName, status)
}

func testAccBucketLifecycleConfigurationConfig_basicUpdate(rName, date, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id     = %[1]q
    status = "Enabled"

    expiration {
      date = %[2]q
    }

    # One of prefix or filter required to ensure XML is well-formed
    filter {
      prefix = %[3]q
    }
  }
}
`, rName, date, prefix)
}

func testAccBucketLifecycleConfigurationConfig_basicPrefix(rName, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id = %[1]q

    # One of prefix or filter required to ensure XML is well-formed
    prefix = %[2]q
    status = "Enabled"

    expiration {
      days = 365
    }
  }
}
`, rName, prefix)
}

func testAccBucketLifecycleConfigurationConfig_filterTag(rName, key, value string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id     = %[1]q
    status = "Enabled"

    filter {
      tag {
        key   = %[2]q
        value = %[3]q
      }
    }

    expiration {
      days = 365
    }
  }
}
`, rName, key, value)
}

func testAccBucketLifecycleConfigurationConfig_ruleExpirationExpiredDeleteMarker(rName string, expired bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id     = %[1]q
    status = "Enabled"

    expiration {
      expired_object_delete_marker = %[2]t
    }
  }
}
`, rName, expired)
}

func testAccBucketLifecycleConfigurationConfig_ruleExpirationEmptyBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id     = %[1]q
    status = "Enabled"

    expiration {}
  }
}
`, rName)
}

func testAccBucketLifecycleConfigurationConfig_ruleAbortIncompleteMultipartUpload(rName string, days int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    abort_incomplete_multipart_upload {
      days_after_initiation = %[2]d
    }

    id     = %[1]q
    status = "Enabled"
  }
}
`, rName, days)
}

func testAccBucketLifecycleConfigurationConfig_multipleRules(rName, date string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = "log"

    expiration {
      days = 90
    }

    filter {
      and {
        prefix = "log/"

        tags = {
          key1 = "value1"
          key2 = "value2"
        }
      }
    }

    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "GLACIER"
    }
  }

  rule {
    id = "tmp"

    filter {
      prefix = "tmp/"
    }

    expiration {
      date = %[2]q
    }

    status = "Enabled"
  }
}
`, rName, date)
}

func testAccBucketLifecycleConfigurationConfig_nonCurrentVersionExpiration(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    # One of prefix or filter required to ensure XML is well-formed
    filter {
      prefix = "config/"
    }

    noncurrent_version_expiration {
      noncurrent_days = 90
    }

    status = "Enabled"
  }
}
`, rName)
}

func testAccBucketLifecycleConfigurationConfig_nonCurrentVersionTransition(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    # One of prefix or filter required to ensure XML is well-formed
    filter {
      prefix = "config/"
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }

    noncurrent_version_transition {
      noncurrent_days = 60
      storage_class   = "GLACIER"
    }

    status = "Enabled"
  }
}
`, rName)
}

func testAccBucketLifecycleConfigurationConfig_transitionStorageClassOnly(rName, storageClass string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }

    noncurrent_version_transition {
      noncurrent_days = 0
      storage_class   = "INTELLIGENT_TIERING"
    }

    transition {
      storage_class = %[2]q
    }

    status = "Enabled"
  }
}
`, rName, storageClass)
}

func testAccBucketLifecycleConfigurationConfig_zeroDaysTransition(rName, storageClass string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }

    noncurrent_version_transition {
      noncurrent_days = 0
      storage_class   = "INTELLIGENT_TIERING"
    }

    transition {
      days          = 0
      storage_class = %[2]q
    }

    status = "Enabled"
  }
}
`, rName, storageClass)
}

func testAccBucketLifecycleConfigurationConfig_dateTransition(rName, transitionDate, storageClass string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }

    noncurrent_version_transition {
      noncurrent_days = 0
      storage_class   = "INTELLIGENT_TIERING"
    }

    transition {
      date          = %[2]q
      storage_class = %[3]q
    }

    status = "Enabled"
  }
}
`, rName, transitionDate, storageClass)
}

func testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    filter {}

    noncurrent_version_expiration {
      newer_noncurrent_versions = 2
      noncurrent_days           = 30
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }

    status = "Enabled"
  }
}
`, rName)
}

func testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThan(rName, date string, sizeGreaterThan int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    expiration {
      date = %[2]q
    }

    filter {
      object_size_greater_than = %[3]d
    }

    status = "Enabled"
  }
}
`, rName, date, sizeGreaterThan)
}

func testAccBucketLifecycleConfigurationConfig_filterObjectSizeLessThan(rName, date string, sizeLessThan int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    expiration {
      date = %[2]q
    }

    filter {
      object_size_less_than = %[3]d
    }

    status = "Enabled"
  }
}
`, rName, date, sizeLessThan)
}

func testAccBucketLifecycleConfigurationConfig_filterObjectSizeRange(rName, date string, sizeGreaterThan, sizeLessThan int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    expiration {
      date = %[2]q
    }

    filter {
      and {
        object_size_greater_than = %[3]d
        object_size_less_than    = %[4]d
      }
    }

    status = "Enabled"
  }
}
`, rName, date, sizeGreaterThan, sizeLessThan)
}

func testAccBucketLifecycleConfigurationConfig_filterObjectSizeRangeAndPrefix(rName, date string, sizeGreaterThan, sizeLessThan int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    expiration {
      date = %[2]q
    }

    filter {
      and {
        object_size_greater_than = %[3]d
        object_size_less_than    = %[4]d
        prefix                   = %[1]q
      }
    }

    status = "Enabled"
  }
}
`, rName, date, sizeGreaterThan, sizeLessThan)
}

func testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThanAndPrefix(rName, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    id = %[1]q

    expiration {
      days = 90
    }

    filter {
      and {
        object_size_greater_than = 300
        prefix                   = %[2]q
      }
    }

    status = "Enabled"
  }
}`, rName, prefix)
}

func testAccBucketLifecycleConfigurationConfig_filterPrefix(rName, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    id = %[1]q

    expiration {
      days = 90
    }

    filter {
      prefix = %[2]q
    }

    status = "Enabled"
  }
}`, rName, prefix)
}

func testAccBucketLifecycleConfigurationConfig_migrateNoChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id     = "id1"
    prefix = "path1/"
    status = "Enabled"

    expiration {
      expired_object_delete_marker = true
    }
  }
}
`, rName)
}

func testAccBucketLifecycleConfigurationConfig_migrateChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id     = "id1"
    prefix = "path1/"
    status = "Disabled"

    expiration {
      expired_object_delete_marker = false
    }
  }
}
`, rName)
}

func testAccBucketLifecycleConfigurationConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  rule {
    id     = %[1]q
    status = "Enabled"

    expiration {
      days = 365
    }
  }
}
`, rName))
}
