// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"maps"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// This file contains tests for validating updates from resources deployed using v5.86.0 of the provider,
// where the migration to the Terraform Plugin Framework caused "inconsisten result" errors.

const (
	providerVersion_5_86_0 = "5.86.0"
)

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_rule_NoFilterOrPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_rule_NoFilterOrPrefix(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(365),
							names.AttrFilter:                    checkFilter_None(),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_rule_NoFilterOrPrefix(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(365),
							names.AttrFilter:                    checkFilter_None(),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_FilterWithPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_filterWithPrefix(rName, date, "prefix/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(date),
							names.AttrFilter:                    checkProvider5_86_0Filter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_filterWithPrefix(rName, date, "prefix/"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_Filter_ObjectSizeGreaterThan(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThan(rName, date, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(date),
							names.AttrFilter:                    checkProvider5_86_0Filter_GreaterThan(100),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThan(rName, date, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(date),
							names.AttrFilter:                    checkFilter_GreaterThan(100),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_Filter_ObjectSizeLessThan(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeLessThan(rName, date, 500),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(date),
							names.AttrFilter:                    checkProvider5_86_0Filter_LessThan(500),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_filterObjectSizeLessThan(rName, date, 500),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(date),
							names.AttrFilter:                    checkFilter_LessThan(500),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_Filter_ObjectSizeRange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeRange(rName, date, 500, 64000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(date),
							names.AttrFilter: checkProvider5_86_0Filter_And(
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"object_size_greater_than": knownvalue.Int64Exact(500),
									"object_size_less_than":    knownvalue.Int64Exact(64000),
									names.AttrPrefix:           knownvalue.StringExact(""),
									names.AttrTags:             knownvalue.Null(),
								}),
							),
							names.AttrID:                    knownvalue.StringExact(rName),
							"noncurrent_version_expiration": checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                knownvalue.StringExact(""),
							names.AttrStatus:                knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                    checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_filterObjectSizeRange(rName, date, 500, 64000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(date),
							names.AttrFilter: checkFilter_And(
								checkAnd(map[string]knownvalue.Check{
									"object_size_greater_than": knownvalue.Int64Exact(500),
									"object_size_less_than":    knownvalue.Int64Exact(64000),
								}),
							),
							names.AttrID:                    knownvalue.StringExact(rName),
							"noncurrent_version_expiration": checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                knownvalue.StringExact(""),
							names.AttrStatus:                knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                    checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_Filter_ObjectSizeRangeAndPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeRangeAndPrefix(rName, date, 500, 64000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(date),
							names.AttrFilter: checkProvider5_86_0Filter_And(
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"object_size_greater_than": knownvalue.Int64Exact(500),
									"object_size_less_than":    knownvalue.Int64Exact(64000),
									names.AttrPrefix:           knownvalue.StringExact(rName),
									names.AttrTags:             knownvalue.Null(),
								}),
							),
							names.AttrID:                    knownvalue.StringExact(rName),
							"noncurrent_version_expiration": checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                knownvalue.StringExact(""),
							names.AttrStatus:                knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                    checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_filterObjectSizeRangeAndPrefix(rName, date, 500, 64000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(date),
							names.AttrFilter: checkFilter_And(
								checkAnd(map[string]knownvalue.Check{
									"object_size_greater_than": knownvalue.Int64Exact(500),
									"object_size_less_than":    knownvalue.Int64Exact(64000),
									names.AttrPrefix:           knownvalue.StringExact(rName),
								}),
							),
							names.AttrID:                    knownvalue.StringExact(rName),
							"noncurrent_version_expiration": checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                knownvalue.StringExact(""),
							names.AttrStatus:                knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                    checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_Filter_And_Tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_filter_And_Tags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(90),
							names.AttrFilter: checkProvider5_86_0Filter_And(
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"object_size_greater_than": knownvalue.Int64Exact(0),
									"object_size_less_than":    knownvalue.Int64Exact(0),
									names.AttrPrefix:           knownvalue.StringExact(""),
									names.AttrTags: knownvalue.MapExact(map[string]knownvalue.Check{
										acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
										acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
									}),
								}),
							),
							names.AttrID:                    knownvalue.StringExact(rName),
							"noncurrent_version_expiration": checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                knownvalue.StringExact(""),
							names.AttrStatus:                knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                    checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_filter_And_Tags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(90),
							names.AttrFilter: checkFilter_And(
								checkAnd(map[string]knownvalue.Check{
									names.AttrTags: knownvalue.MapExact(map[string]knownvalue.Check{
										acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
										acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
									}),
								}),
							),
							names.AttrID:                    knownvalue.StringExact(rName),
							"noncurrent_version_expiration": checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                knownvalue.StringExact(""),
							names.AttrStatus:                knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                    checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_Filter_Tag(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_filterTag(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(365),
							names.AttrFilter:                    checkProvider5_86_0Filter_Tag(acctest.CtKey1, acctest.CtValue1),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_filterTag(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(365),
							names.AttrFilter:                    checkFilter_Tag(acctest.CtKey1, acctest.CtValue1),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_EmptyFilter_NonCurrentVersions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName, "varies_by_storage_class"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkProvider5_86_0Filter_Prefix(""),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_VersionsAndDays(2, 30),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(30, awstypes.TransitionStorageClassStandardIa),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":     checkTransitions(),
						}),
					})),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName, "varies_by_storage_class"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkFilter_Prefix(""),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_VersionsAndDays(2, 30),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(30, awstypes.TransitionStorageClassStandardIa),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":     checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("varies_by_storage_class")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_EmptyFilter_NonCurrentVersions_WithChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName, "varies_by_storage_class"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkProvider5_86_0Filter_Empty(),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_VersionsAndDays(2, 30),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(30, awstypes.TransitionStorageClassStandardIa),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":     checkTransitions(),
						}),
					})),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName, "all_storage_classes_128K"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkFilter_Prefix(""),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_VersionsAndDays(2, 30),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(30, awstypes.TransitionStorageClassStandardIa),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":     checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_None(),
								names.AttrFilter:                    checkFilter_Prefix(""),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_VersionsAndDays(2, 30),
								"noncurrent_version_transition": checkNoncurrentVersionTransitions(
									checkNoncurrentVersionTransition_Days(30, awstypes.TransitionStorageClassStandardIa),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":     checkTransitions(),
							}),
						})),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_nonCurrentVersionExpiration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_nonCurrentVersionExpiration(rName, 90),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkProvider5_86_0Filter_Prefix("config/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_Days(90),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_nonCurrentVersionExpiration(rName, 90),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkFilter_Prefix("config/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_Days(90),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_ruleAbortIncompleteMultipartUpload(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_ruleAbortIncompleteMultipartUpload(rName, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(7),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkProvider5_86_0Filter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_ruleAbortIncompleteMultipartUpload(rName, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(7),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_RuleExpiration_expireMarkerOnly(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationExpiredDeleteMarker(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_DeleteMarker(true),
							names.AttrFilter:                    checkProvider5_86_0Filter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_ruleExpirationExpiredDeleteMarker(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_DeleteMarker(true),
							names.AttrFilter:                    checkFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_RuleExpiration_emptyBlock(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationEmptyBlock(rName, "prefix/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Empty(),
							names.AttrFilter:                    checkProvider5_86_0Filter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_ruleExpirationEmptyBlock(rName, "prefix/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Empty(),
							names.AttrFilter:                    checkFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_RulePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_rulePrefix(rName, "path1/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(365),
							names.AttrFilter:                    checkFilter_None(),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact("path1/"),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_rulePrefix(rName, "path1/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(365),
							names.AttrFilter:                    checkFilter_None(),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact("path1/"),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_TransitionDate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_dateTransition(rName, date, awstypes.TransitionStorageClassStandardIa),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(1),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkProvider5_86_0Filter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkTransition_Date(date, awstypes.TransitionStorageClassStandardIa),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_dateTransition(rName, date, awstypes.TransitionStorageClassStandardIa),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(1),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkTransition_Date(date, awstypes.TransitionStorageClassStandardIa),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_TransitionStorageClassOnly_intelligentTiering(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_transitionStorageClassOnly(rName, awstypes.TransitionStorageClassIntelligentTiering),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(1),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkProvider5_86_0Filter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkTransition_StorageClass(awstypes.TransitionStorageClassIntelligentTiering),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_transitionStorageClassOnly(rName, awstypes.TransitionStorageClassIntelligentTiering),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(1),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkTransition_StorageClass(awstypes.TransitionStorageClassIntelligentTiering),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_upgradeV5_86_0_TransitionZeroDays_intelligentTiering(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: providerVersion_5_86_0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_zeroDaysTransition(rName, awstypes.TransitionStorageClassIntelligentTiering),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(1),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkProvider5_86_0Filter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_zeroDaysTransition(rName, awstypes.TransitionStorageClassIntelligentTiering),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(1),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func checkProvider5_86_0Filter_Empty() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			schemaProvider5_86_0FilterDefaults(),
		),
	})
}

func checkProvider5_86_0Filter_And(check knownvalue.Check) knownvalue.Check {
	checks := schemaProvider5_86_0FilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"and": knownvalue.ListExact([]knownvalue.Check{
			check,
		}),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkProvider5_86_0Filter_Prefix(prefix string) knownvalue.Check {
	checks := schemaProvider5_86_0FilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		names.AttrPrefix: knownvalue.StringExact(prefix),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkProvider5_86_0Filter_GreaterThan(size int64) knownvalue.Check {
	checks := schemaProvider5_86_0FilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"object_size_greater_than": knownvalue.Int64Exact(size),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkProvider5_86_0Filter_LessThan(size int64) knownvalue.Check {
	checks := schemaProvider5_86_0FilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"object_size_less_than": knownvalue.Int64Exact(size),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkProvider5_86_0Filter_Tag(key, value string) knownvalue.Check {
	checks := schemaProvider5_86_0FilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"tag": knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				names.AttrKey:   knownvalue.StringExact(key),
				names.AttrValue: knownvalue.StringExact(value),
			}),
		}),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func schemaProvider5_86_0FilterDefaults() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"and":                      knownvalue.ListExact([]knownvalue.Check{}),
		"object_size_greater_than": knownvalue.Null(),
		"object_size_less_than":    knownvalue.Null(),
		names.AttrPrefix:           knownvalue.StringExact(""),
		"tag":                      knownvalue.ListExact([]knownvalue.Check{}),
	}
}
