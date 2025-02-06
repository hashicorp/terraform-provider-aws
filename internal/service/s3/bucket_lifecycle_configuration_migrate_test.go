// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"maps"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfplancheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/plancheck"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Notes on migration testing
//
// Because this migration includes a schema change, the standard testing pattern described at
// https://hashicorp.github.io/terraform-provider-aws/terraform-plugin-migrations/#testing
// cannot be used.
//
// To ensure that the planned change is as expected, first create the test using the standard pattern.
// This will either fail with a non-empty plan if there are changes or an error like
// > Error retrieving state, there may be dangling resources: exit status 1
// > Failed to marshal state to json: schema version 0 for aws_s3_bucket_lifecycle_configuration.test in state does not match version 1 from the provider
// if there are no changes.
//
// Once the plan is as expected, remove the `PlanOnly` parameter and add `ConfigStateChecks` and `ConfigPlanChecks` checks

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_basic(t *testing.T) {
	// Expected change: removes `rule[0].filter` from state
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
						VersionConstraint: "5.85.0",
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_basic(rName),
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
							"expiration":                        checkOldExpiration_Days(365),
							names.AttrFilter:                    checkOldFilter_Empty(),
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
				Config:                   testAccBucketLifecycleConfigurationConfig_basic(rName),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Days(365),
								// names.AttrFilter:
								names.AttrID:                    knownvalue.StringExact(rName),
								"noncurrent_version_expiration": checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition": checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                knownvalue.StringExact(""),
								names.AttrStatus:                knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                    checkTransitions(),
							}),
						})),
						// This is checking the change _after_ the state migration step happens
						tfplancheck.ExpectKnownValueChange(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey(names.AttrFilter),
							checkFilter_Empty(),
							checkFilter_None(),
						),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_FilterWithPrefix(t *testing.T) {
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
						VersionConstraint: "5.85.0",
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_basicUpdate(rName, date, "prefix/"),
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
							"expiration":                        checkOldExpiration_Date(date),
							names.AttrFilter:                    checkOldFilter_Prefix("prefix/"),
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
				Config:                   testAccBucketLifecycleConfigurationConfig_basicUpdate(rName, date, "prefix/"),
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter:                    checkFilter_Prefix("prefix/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_Filter_ObjectSizeGreaterThan(t *testing.T) {
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
						VersionConstraint: "5.85.0",
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
							names.AttrFilter:                    checkOldFilter_GreaterThan(100),
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_Filter_ObjectSizeLessThan(t *testing.T) {
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
						VersionConstraint: "5.85.0",
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
							names.AttrFilter:                    checkOldFilter_LessThan(500),
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_Filter_ObjectSizeRange(t *testing.T) {
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
						VersionConstraint: "5.85.0",
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
							names.AttrFilter: checkOldFilter_And(
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter: checkFilter_And(
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_Filter_ObjectSizeRangeAndPrefix(t *testing.T) {
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
						VersionConstraint: "5.85.0",
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
							names.AttrFilter: checkOldFilter_And(
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter: checkFilter_And(
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_Filter_And_Tags(t *testing.T) {
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
						VersionConstraint: "5.85.0",
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
							"expiration":                        checkOldExpiration_Days(90),
							names.AttrFilter: checkOldFilter_And(
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Days(90),
								names.AttrFilter: checkFilter_And(
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_Filter_Tag(t *testing.T) {
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
						VersionConstraint: "5.85.0",
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
							"expiration":                        checkOldExpiration_Days(365),
							names.AttrFilter:                    checkOldFilter_Tag(acctest.CtKey1, acctest.CtValue1),
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_EmptyFilter_NonCurrentVersions(t *testing.T) {
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
						VersionConstraint: "5.85.0",
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkOldFilter_Empty(),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"newer_noncurrent_versions": knownvalue.StringExact("2"),
									"noncurrent_days":           knownvalue.Int64Exact(30),
								}),
							}),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkOldNoncurrentVersionTransition_Days(30, "STANDARD_IA"),
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
				Config:                   testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName),
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
							names.AttrFilter:                    checkFilter_Empty(),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"newer_noncurrent_versions": knownvalue.Int32Exact(2),
									"noncurrent_days":           knownvalue.Int64Exact(30),
								}),
							}),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(30, "STANDARD_IA"),
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
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_None(),
								names.AttrFilter:                    checkFilter_Empty(),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"newer_noncurrent_versions": knownvalue.Int32Exact(2),
										"noncurrent_days":           knownvalue.Int64Exact(30),
									}),
								}),
								"noncurrent_version_transition": checkNoncurrentVersionTransitions(
									checkNoncurrentVersionTransition_Days(30, "STANDARD_IA"),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_nonCurrentVersionExpiration(t *testing.T) {
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
						VersionConstraint: "5.85.0",
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_nonCurrentVersionExpiration(rName),
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
							names.AttrFilter:                    checkOldFilter_Prefix("config/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkOldNoncurrentVersionExpiration_Days(90),
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
				Config:                   testAccBucketLifecycleConfigurationConfig_nonCurrentVersionExpiration(rName),
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_ruleAbortIncompleteMultipartUpload(t *testing.T) {
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
						VersionConstraint: "5.85.0",
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
							names.AttrFilter:                    checkOldFilter_Prefix("prefix/"),
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_RuleExpiration_expireMarkerOnly(t *testing.T) {
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
						VersionConstraint: "5.85.0",
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
							"expiration":                        checkOldExpiration_DeleteMarker(true),
							names.AttrFilter:                    checkOldFilter_Prefix("prefix/"),
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_RuleExpiration_emptyBlock(t *testing.T) {
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
						VersionConstraint: "5.85.0",
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationEmptyBlock(rName),
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
							"expiration":                        checkOldExpiration_Empty(),
							names.AttrFilter:                    checkOldFilter_Prefix("prefix/"),
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
				Config:                   testAccBucketLifecycleConfigurationConfig_ruleExpirationEmptyBlock(rName),
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_RulePrefix(t *testing.T) {
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
						VersionConstraint: "5.85.0",
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
							"expiration":                        checkOldExpiration_Days(365),
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_TransitionDate(t *testing.T) {
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
						VersionConstraint: "5.85.0",
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_dateTransition(rName, date, types.TransitionStorageClassStandardIa),
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
							names.AttrFilter:                    checkOldFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkOldNoncurrentVersionTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkOldTransition_Date(date, types.TransitionStorageClassStandardIa),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_dateTransition(rName, date, types.TransitionStorageClassStandardIa),
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
								checkNoncurrentVersionTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkTransition_Date(date, types.TransitionStorageClassStandardIa),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(1),
								"expiration":                        checkExpiration_None(),
								names.AttrFilter:                    checkFilter_Prefix("prefix/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition": checkNoncurrentVersionTransitions(
									checkNoncurrentVersionTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition": checkTransitions(
									checkTransition_Date(date, types.TransitionStorageClassStandardIa),
								),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_TransitionStorageClassOnly_intelligentTiering(t *testing.T) {
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
						VersionConstraint: "5.85.0",
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_transitionStorageClassOnly(rName, types.TransitionStorageClassIntelligentTiering),
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
							names.AttrFilter:                    checkOldFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkOldNoncurrentVersionTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkOldTransition_StorageClass(types.TransitionStorageClassIntelligentTiering),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_transitionStorageClassOnly(rName, types.TransitionStorageClassIntelligentTiering),
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
								checkNoncurrentVersionTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkTransition_StorageClass(types.TransitionStorageClassIntelligentTiering),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(1),
								"expiration":                        checkExpiration_None(),
								names.AttrFilter:                    checkFilter_Prefix("prefix/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition": checkNoncurrentVersionTransitions(
									checkNoncurrentVersionTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition": checkTransitions(
									checkTransition_StorageClass(types.TransitionStorageClassIntelligentTiering),
								),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigration_TransitionZeroDays_intelligentTiering(t *testing.T) {
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
						VersionConstraint: "5.85.0",
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_zeroDaysTransition(rName, types.TransitionStorageClassIntelligentTiering),
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
							names.AttrFilter:                    checkOldFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkOldNoncurrentVersionTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkOldTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_zeroDaysTransition(rName, types.TransitionStorageClassIntelligentTiering),
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
								checkNoncurrentVersionTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(1),
								"expiration":                        checkExpiration_None(),
								names.AttrFilter:                    checkFilter_Prefix("prefix/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition": checkNoncurrentVersionTransitions(
									checkNoncurrentVersionTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition": checkTransitions(
									checkTransition_Days(0, types.TransitionStorageClassIntelligentTiering),
								),
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

func checkOldExpiration_Empty() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			oldExpirationDefaults(),
		),
	})
}

func checkOldExpiration_Date(date string) knownvalue.Check {
	checks := oldExpirationDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"date": knownvalue.StringExact(date),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkOldExpiration_Days(days int64) knownvalue.Check {
	checks := oldExpirationDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"days": knownvalue.Int64Exact(days),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkOldExpiration_DeleteMarker(marker bool) knownvalue.Check {
	checks := oldExpirationDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"expired_object_delete_marker": knownvalue.Bool(marker),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func oldExpirationDefaults() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"date":                         knownvalue.StringExact(""),
		"days":                         knownvalue.Int64Exact(0),
		"expired_object_delete_marker": knownvalue.Bool(false),
	}
}

func checkOldFilter_Empty() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			oldFilterDefaults(),
		),
	})
}

func checkOldFilter_And(check knownvalue.Check) knownvalue.Check {
	checks := oldFilterDefaults()
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

func checkOldFilter_Prefix(prefix string) knownvalue.Check {
	checks := oldFilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		names.AttrPrefix: knownvalue.StringExact(prefix),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkOldFilter_GreaterThan(size int64) knownvalue.Check {
	checks := oldFilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"object_size_greater_than": knownvalue.StringExact(strconv.FormatInt(size, 10)),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkOldFilter_LessThan(size int64) knownvalue.Check {
	checks := oldFilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"object_size_less_than": knownvalue.StringExact(strconv.FormatInt(size, 10)),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkOldFilter_Tag(key, value string) knownvalue.Check {
	checks := oldFilterDefaults()
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

func oldFilterDefaults() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"and":                      knownvalue.ListExact([]knownvalue.Check{}),
		"object_size_greater_than": knownvalue.StringExact(""),
		"object_size_less_than":    knownvalue.StringExact(""),
		names.AttrPrefix:           knownvalue.StringExact(""),
		"tag":                      knownvalue.ListExact([]knownvalue.Check{}),
	}
}

func checkOldNoncurrentVersionExpiration_Days(days int64) knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"newer_noncurrent_versions": knownvalue.StringExact(""),
			"noncurrent_days":           knownvalue.Int64Exact(days),
		}),
	})
}

func checkOldNoncurrentVersionTransition_Days(days int64, class types.TransitionStorageClass) knownvalue.Check {
	return knownvalue.ObjectExact(map[string]knownvalue.Check{
		"newer_noncurrent_versions": knownvalue.StringExact(""),
		"noncurrent_days":           knownvalue.Int64Exact(days),
		names.AttrStorageClass:      tfknownvalue.StringExact(class),
	})
}

func checkOldTransition_Date(date string, class types.TransitionStorageClass) knownvalue.Check {
	checks := oldTransitionDefaults(class)
	maps.Copy(checks, map[string]knownvalue.Check{
		"date": knownvalue.StringExact(date),
	})
	return knownvalue.ObjectExact(
		checks,
	)
}

func checkOldTransition_Days(days int64, class types.TransitionStorageClass) knownvalue.Check {
	checks := oldTransitionDefaults(class)
	maps.Copy(checks, map[string]knownvalue.Check{
		"days": knownvalue.Int64Exact(days),
	})
	return knownvalue.ObjectExact(
		checks,
	)
}

func checkOldTransition_StorageClass(class types.TransitionStorageClass) knownvalue.Check {
	checks := oldTransitionDefaults(class)
	return knownvalue.ObjectExact(
		checks,
	)
}

func oldTransitionDefaults(class types.TransitionStorageClass) map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"date":                 knownvalue.StringExact(""),
		"days":                 knownvalue.Null(),
		names.AttrStorageClass: tfknownvalue.StringExact(class),
	}
}

func testAccBucketLifecycleConfigurationConfig_filter_And_Tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    expiration {
      days = 90
    }

    filter {
      and {
        tags = {
          key1 = "value1"
          key2 = "value2"
        }
      }
    }

    status = "Enabled"
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}
