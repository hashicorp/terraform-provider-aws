// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"maps"
	"strconv"
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
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfplancheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/plancheck"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	providerVersionSchemaV0 = "5.85.0"
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_rule_NoFilterOrPrefix(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							"expiration":                        checkSchemaV0Expiration_Days(365),
							names.AttrFilter:                    checkSchemaV0Filter_Empty(),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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
						// This is checking the change _after_ the state migration step happens
						tfplancheck.ExpectKnownValueChange(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey(names.AttrFilter),
							checkFilter_Prefix(""),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_FilterWithPrefix(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							"expiration":                        checkSchemaV0Expiration_Date(date),
							names.AttrFilter:                    checkSchemaV0Filter_Prefix("prefix/"),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_filterWithEmptyPrefix(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_filterWithEmptyPrefix(rName, date),
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
							"expiration":                        checkSchemaV0Expiration_Date(date),
							names.AttrFilter:                    checkSchemaV0Filter_Prefix(""),
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
				Config:                   testAccBucketLifecycleConfigurationConfig_filterWithEmptyPrefix(rName, date),
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
							names.AttrFilter:                    checkFilter_Prefix(""),
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
								names.AttrFilter:                    checkFilter_Prefix(""),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_Filter_ObjectSizeGreaterThan(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter:                    checkSchemaV0Filter_GreaterThan(100),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_Filter_ObjectSizeLessThan(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter:                    checkSchemaV0Filter_LessThan(500),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_Filter_ObjectSizeRange(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter: checkSchemaV0Filter_And(
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_Filter_ObjectSizeRangeAndPrefix(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter: checkSchemaV0Filter_And(
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_Filter_And_Tags(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							"expiration":                        checkSchemaV0Expiration_Days(90),
							names.AttrFilter: checkSchemaV0Filter_And(
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
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

// Simulate a change in default value for `transition_default_minimum_object_size`
func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_Filter_And_ZeroLessThan_ChangeOnUpdate(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
					},
				},
				Config: testAccBucketLifecycleConfigurationConfig_Migrate_Filter_And_ZeroLessThan(rName, false),
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
							"expiration":                        checkSchemaV0Expiration_Days(1),
							names.AttrFilter: checkSchemaV0Filter_And(
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"object_size_greater_than": knownvalue.Int64Exact(0),
									"object_size_less_than":    knownvalue.Int64Exact(0),
									names.AttrPrefix:           knownvalue.StringExact("baseline/"),
									names.AttrTags: knownvalue.MapExact(map[string]knownvalue.Check{
										"Key":   knownvalue.StringExact("data-lifecycle-action"),
										"Value": knownvalue.StringExact("delete"),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("varies_by_storage_class")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketLifecycleConfigurationConfig_Migrate_Filter_And_ZeroLessThan(rName, true),
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
							"expiration":                        checkExpiration_Days(1),
							names.AttrFilter: checkFilter_And(
								checkAnd(map[string]knownvalue.Check{
									names.AttrPrefix: knownvalue.StringExact("baseline/"),
									names.AttrTags: knownvalue.MapExact(map[string]knownvalue.Check{
										"Key":   knownvalue.StringExact("data-lifecycle-action"),
										"Value": knownvalue.StringExact("delete"),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Days(1),
								names.AttrFilter: checkFilter_And(
									checkAnd(map[string]knownvalue.Check{
										names.AttrPrefix: knownvalue.StringExact("baseline/"),
										names.AttrTags: knownvalue.MapExact(map[string]knownvalue.Check{
											"Key":   knownvalue.StringExact("data-lifecycle-action"),
											"Value": knownvalue.StringExact("delete"),
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
						tfplancheck.ExpectKnownValueChange(resourceName, tfjsonpath.New("transition_default_minimum_object_size"),
							knownvalue.StringExact("varies_by_storage_class"),
							knownvalue.StringExact("all_storage_classes_128K"),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_Filter_Tag(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							"expiration":                        checkSchemaV0Expiration_Days(365),
							names.AttrFilter:                    checkSchemaV0Filter_Tag(acctest.CtKey1, acctest.CtValue1),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_EmptyFilter_NonCurrentVersions(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter:                    checkSchemaV0Filter_Empty(),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkSchemaV0NoncurrentVersionExpiration_VersionsAndDays(2, 30),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkSchemaV0NoncurrentVersionTransition_Days(30, awstypes.TransitionStorageClassStandardIa),
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
						// This is checking the change _after_ the state migration step happens
						plancheck.ExpectEmptyPlan(),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_EmptyFilter_NonCurrentVersions_WithChange(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter:                    checkSchemaV0Filter_Empty(),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkSchemaV0NoncurrentVersionExpiration_VersionsAndDays(2, 30),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkSchemaV0NoncurrentVersionTransition_Days(30, awstypes.TransitionStorageClassStandardIa),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_nonCurrentVersionExpiration(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter:                    checkSchemaV0Filter_Prefix("config/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkSchemaV0NoncurrentVersionExpiration_Days(90),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_ruleAbortIncompleteMultipartUpload(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter:                    checkSchemaV0Filter_Prefix("prefix/"),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_RuleExpiration_expireMarkerOnly(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							"expiration":                        checkSchemaV0Expiration_DeleteMarker(true),
							names.AttrFilter:                    checkSchemaV0Filter_Prefix("prefix/"),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_RuleExpiration_emptyBlock(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							"expiration":                        checkSchemaV0Expiration_Empty(),
							names.AttrFilter:                    checkSchemaV0Filter_Prefix("prefix/"),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_RulePrefix(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							"expiration":                        checkSchemaV0Expiration_Days(365),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_TransitionDate(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter:                    checkSchemaV0Filter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkSchemaV0NoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkSchemaV0Transition_Date(date, awstypes.TransitionStorageClassStandardIa),
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
									checkNoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition": checkTransitions(
									checkTransition_Date(date, awstypes.TransitionStorageClassStandardIa),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_TransitionStorageClassOnly_intelligentTiering(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter:                    checkSchemaV0Filter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkSchemaV0NoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkSchemaV0Transition_StorageClass(awstypes.TransitionStorageClassIntelligentTiering),
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
									checkNoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition": checkTransitions(
									checkTransition_StorageClass(awstypes.TransitionStorageClassIntelligentTiering),
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

func TestAccS3BucketLifecycleConfiguration_frameworkMigrationV0_TransitionZeroDays_intelligentTiering(t *testing.T) {
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
						VersionConstraint: providerVersionSchemaV0,
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
							names.AttrFilter:                    checkSchemaV0Filter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkSchemaV0NoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkSchemaV0Transition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
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
									checkNoncurrentVersionTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition": checkTransitions(
									checkTransition_Days(0, awstypes.TransitionStorageClassIntelligentTiering),
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

func checkSchemaV0Expiration_Empty() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			schemaV0ExpirationDefaults(),
		),
	})
}

func checkSchemaV0Expiration_Date(date string) knownvalue.Check {
	checks := schemaV0ExpirationDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"date": knownvalue.StringExact(date),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkSchemaV0Expiration_Days(days int64) knownvalue.Check {
	checks := schemaV0ExpirationDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"days": knownvalue.Int64Exact(days),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkSchemaV0Expiration_DeleteMarker(marker bool) knownvalue.Check {
	checks := schemaV0ExpirationDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"expired_object_delete_marker": knownvalue.Bool(marker),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func schemaV0ExpirationDefaults() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"date":                         knownvalue.StringExact(""),
		"days":                         knownvalue.Int64Exact(0),
		"expired_object_delete_marker": knownvalue.Bool(false),
	}
}

func checkSchemaV0Filter_Empty() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			schemaV0FilterDefaults(),
		),
	})
}

func checkSchemaV0Filter_And(check knownvalue.Check) knownvalue.Check {
	checks := schemaV0FilterDefaults()
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

func checkSchemaV0Filter_Prefix(prefix string) knownvalue.Check {
	checks := schemaV0FilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		names.AttrPrefix: knownvalue.StringExact(prefix),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkSchemaV0Filter_GreaterThan(size int64) knownvalue.Check {
	checks := schemaV0FilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"object_size_greater_than": knownvalue.StringExact(strconv.FormatInt(size, 10)),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkSchemaV0Filter_LessThan(size int64) knownvalue.Check {
	checks := schemaV0FilterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"object_size_less_than": knownvalue.StringExact(strconv.FormatInt(size, 10)),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkSchemaV0Filter_Tag(key, value string) knownvalue.Check {
	checks := schemaV0FilterDefaults()
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

func schemaV0FilterDefaults() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"and":                      knownvalue.ListExact([]knownvalue.Check{}),
		"object_size_greater_than": knownvalue.StringExact(""),
		"object_size_less_than":    knownvalue.StringExact(""),
		names.AttrPrefix:           knownvalue.StringExact(""),
		"tag":                      knownvalue.ListExact([]knownvalue.Check{}),
	}
}

func checkSchemaV0NoncurrentVersionExpiration_Days(days int64) knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"newer_noncurrent_versions": knownvalue.StringExact(""),
			"noncurrent_days":           knownvalue.Int64Exact(days),
		}),
	})
}

func checkSchemaV0NoncurrentVersionExpiration_VersionsAndDays(versions, days int64) knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"newer_noncurrent_versions": knownvalue.StringExact(strconv.FormatInt(versions, 10)),
			"noncurrent_days":           knownvalue.Int64Exact(days),
		}),
	})
}

func checkSchemaV0NoncurrentVersionTransition_Days(days int64, class awstypes.TransitionStorageClass) knownvalue.Check {
	return knownvalue.ObjectExact(map[string]knownvalue.Check{
		"newer_noncurrent_versions": knownvalue.StringExact(""),
		"noncurrent_days":           knownvalue.Int64Exact(days),
		names.AttrStorageClass:      tfknownvalue.StringExact(class),
	})
}

func checkSchemaV0Transition_Date(date string, class awstypes.TransitionStorageClass) knownvalue.Check {
	checks := schemaV0TransitionDefaults(class)
	maps.Copy(checks, map[string]knownvalue.Check{
		"date": knownvalue.StringExact(date),
	})
	return knownvalue.ObjectExact(
		checks,
	)
}

func checkSchemaV0Transition_Days(days int64, class awstypes.TransitionStorageClass) knownvalue.Check {
	checks := schemaV0TransitionDefaults(class)
	maps.Copy(checks, map[string]knownvalue.Check{
		"days": knownvalue.Int64Exact(days),
	})
	return knownvalue.ObjectExact(
		checks,
	)
}

func checkSchemaV0Transition_StorageClass(class awstypes.TransitionStorageClass) knownvalue.Check {
	checks := schemaV0TransitionDefaults(class)
	return knownvalue.ObjectExact(
		checks,
	)
}

func schemaV0TransitionDefaults(class awstypes.TransitionStorageClass) map[string]knownvalue.Check {
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

// Set `migrated` to `false` before migration and `true` after
func testAccBucketLifecycleConfigurationConfig_Migrate_Filter_And_ZeroLessThan(rName string, migrated bool) string {
	var transition string
	if !migrated {
		transition = "transition_default_minimum_object_size = \"varies_by_storage_class\""
	}
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    status = "Enabled"

    id = %[1]q

    expiration {
      days = 1
    }

    filter {
      and {
        prefix = "baseline/"
        tags = {
          Key   = "data-lifecycle-action"
          Value = "delete"
        }
      }
    }
  }

  %[2]s
}`, rName, transition)
}
