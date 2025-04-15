// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"maps"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(365),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Days(365),
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
				},
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
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketLifecycleConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_rule_NoFilterOrPrefix(t *testing.T) {
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
				Config: testAccBucketLifecycleConfigurationConfig_filterWithPrefix(rName, date, "logs/"),
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
							names.AttrFilter:                    checkFilter_Prefix("logs/"),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter:                    checkFilter_Prefix("logs/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterWithPrefix(rName, dateUpdated, "tmp/"),
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
							"expiration":                        checkExpiration_Date(dateUpdated),
							names.AttrFilter:                    checkFilter_Prefix("tmp/"),
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
								"expiration":                        checkExpiration_Date(dateUpdated),
								names.AttrFilter:                    checkFilter_Prefix("tmp/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_filterWithEmptyPrefix(t *testing.T) {
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThan(rName, date, 200),
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
							names.AttrFilter:                    checkFilter_GreaterThan(200),
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
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter:                    checkFilter_GreaterThan(200),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThan(rName, date, 200),
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
							names.AttrFilter:                    checkFilter_GreaterThan(200),
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
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter:                    checkFilter_GreaterThan(200),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(date),
							names.AttrFilter:                    checkFilter_GreaterThan(0),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter:                    checkFilter_GreaterThan(0),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThan(rName, date, 200),
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
							names.AttrFilter:                    checkFilter_GreaterThan(200),
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
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter:                    checkFilter_GreaterThan(200),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeLessThan(rName, date, 5000),
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
							names.AttrFilter:                    checkFilter_LessThan(5000),
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
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter:                    checkFilter_LessThan(5000),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter: checkFilter_And(
									checkAnd(map[string]knownvalue.Check{
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeRange(rName, date, 400, 65000),
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
									"object_size_greater_than": knownvalue.Int64Exact(400),
									"object_size_less_than":    knownvalue.Int64Exact(65000),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter: checkFilter_And(
									checkAnd(map[string]knownvalue.Check{
										"object_size_greater_than": knownvalue.Int64Exact(400),
										"object_size_less_than":    knownvalue.Int64Exact(65000),
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
				},
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter: checkFilter_And(
									checkAnd(map[string]knownvalue.Check{
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
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_Filter_PrefixToAnd(t *testing.T) {
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterWithPrefixAndTags(rName, date, "prefix/"),
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
									"object_size_greater_than": knownvalue.Int64Exact(0),
									"object_size_less_than":    knownvalue.Int64Exact(0),
									names.AttrPrefix:           knownvalue.StringExact("prefix/"),
									names.AttrTags: knownvalue.MapExact(map[string]knownvalue.Check{
										acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Date(date),
								names.AttrFilter: checkFilter_And(
									checkAnd(map[string]knownvalue.Check{
										"object_size_greater_than": knownvalue.Int64Exact(0),
										"object_size_less_than":    knownvalue.Int64Exact(0),
										names.AttrPrefix:           knownvalue.StringExact("prefix/"),
										names.AttrTags: knownvalue.MapExact(map[string]knownvalue.Check{
											acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(365),
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
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_basicStatus(rName, tfs3.LifecycleRuleStatusDisabled),
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
							names.AttrFilter:                    checkFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusDisabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(365),
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
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
									names.AttrPrefix: knownvalue.StringExact("log/"),
									names.AttrTags: knownvalue.MapExact(map[string]knownvalue.Check{
										acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
										acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
									}),
								}),
							),
							names.AttrID:                    knownvalue.StringExact("log"),
							"noncurrent_version_expiration": checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                knownvalue.StringExact(""),
							names.AttrStatus:                knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition": checkTransitions(
								checkTransition_Days(30, types.TransitionStorageClassStandardIa),
								checkTransition_Days(60, types.TransitionStorageClassGlacier),
							),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Date(expirationDate),
							names.AttrFilter:                    checkFilter_Prefix("tmp/"),
							names.AttrID:                        knownvalue.StringExact("tmp"),
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
							names.AttrID:                        knownvalue.StringExact(rName + "-1"),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_DeleteMarker(true),
							names.AttrFilter:                    checkFilter_None(),
							names.AttrID:                        knownvalue.StringExact(rName + "-2"),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_nonCurrentVersionExpiration(rName, 100),
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
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_Days(100),
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
								"expiration":                        checkExpiration_None(),
								names.AttrFilter:                    checkFilter_Prefix("config/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_Days(100),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
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
				Config: testAccBucketLifecycleConfigurationConfig_nonCurrentVersionTransition(rName, 30, 60),
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
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(30, types.TransitionStorageClassStandardIa),
								checkNoncurrentVersionTransition_Days(60, types.TransitionStorageClassGlacier),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_None(),
								names.AttrFilter:                    checkFilter_Prefix("config/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition": checkNoncurrentVersionTransitions(
									checkNoncurrentVersionTransition_Days(30, types.TransitionStorageClassStandardIa),
									checkNoncurrentVersionTransition_Days(60, types.TransitionStorageClassGlacier),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":     checkTransitions(),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_nonCurrentVersionTransition(rName, 45, 90),
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
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(45, types.TransitionStorageClassStandardIa),
								checkNoncurrentVersionTransition_Days(90, types.TransitionStorageClassGlacier),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_None(),
								names.AttrFilter:                    checkFilter_Prefix("config/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition": checkNoncurrentVersionTransitions(
									checkNoncurrentVersionTransition_Days(45, types.TransitionStorageClassStandardIa),
									checkNoncurrentVersionTransition_Days(90, types.TransitionStorageClassGlacier),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":     checkTransitions(),
							}),
						})),
					},
				},
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
func TestAccS3BucketLifecycleConfiguration_RulePrefix(t *testing.T) {
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_rulePrefix(rName, "path2/"),
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
							names.AttrPrefix:                    knownvalue.StringExact("path2/"),
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
								names.AttrPrefix:                    knownvalue.StringExact("path2/"),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
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
func TestAccS3BucketLifecycleConfiguration_RulePrefixToFilter(t *testing.T) {
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterPrefix(rName, "path1/"),
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
							names.AttrFilter:                    checkFilter_Prefix("path1/"),
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
								"expiration":                        checkExpiration_Days(90),
								names.AttrFilter:                    checkFilter_Prefix("path1/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TODO: RulePrefix to FilterAndPrefix

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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterTag(rName, acctest.CtKey2, acctest.CtValue2),
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
							names.AttrFilter:                    checkFilter_Tag(acctest.CtKey2, acctest.CtValue2),
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
								names.AttrFilter:                    checkFilter_Tag(acctest.CtKey2, acctest.CtValue2),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_DeleteMarker(false),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_DeleteMarker(false),
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
				},
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationEmptyBlock(rName, "prefix2/"),
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
							names.AttrFilter:                    checkFilter_Prefix("prefix2/"),
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
								"expiration":                        checkExpiration_Empty(),
								names.AttrFilter:                    checkFilter_Prefix("prefix2/"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(5),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_Days(5),
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
				},
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
	dateUpdated := time.Date(currTime.Year()+1, currTime.Month(), currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_dateTransition(rName, dateUpdated, types.TransitionStorageClassStandardIa),
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
								checkTransition_Date(dateUpdated, types.TransitionStorageClassStandardIa),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
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
									checkTransition_Date(dateUpdated, types.TransitionStorageClassStandardIa),
								),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
				Config: testAccBucketLifecycleConfigurationConfig_dateTransition(rName, date, types.TransitionStorageClassIntelligentTiering),
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
								checkTransition_Date(date, types.TransitionStorageClassIntelligentTiering),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
									checkTransition_Date(date, types.TransitionStorageClassIntelligentTiering),
								),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_transitionStorageClassOnly(rName, types.TransitionStorageClassGlacier),
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
								checkTransition_StorageClass(types.TransitionStorageClassGlacier),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
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
									checkTransition_StorageClass(types.TransitionStorageClassGlacier),
								),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_zeroDaysTransition(rName, types.TransitionStorageClassGlacier),
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
								checkTransition_Days(0, types.TransitionStorageClassGlacier),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
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
									checkTransition_Days(0, types.TransitionStorageClassGlacier),
								),
							}),
						})),
					},
				},
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
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
				},
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_dateTransition(rName, date, types.TransitionStorageClassIntelligentTiering),
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
								checkTransition_Date(date, types.TransitionStorageClassIntelligentTiering),
							),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
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
									checkTransition_Date(date, types.TransitionStorageClassIntelligentTiering),
								),
							}),
						})),
					},
				},
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
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
				},
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
				Config: testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName, "varies_by_storage_class"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkFilter_Prefix(""),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_VersionsAndDays(2, 30),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(30, "STANDARD_IA"),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":     checkTransitions(),
						}),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_None(),
								names.AttrFilter:                    checkFilter_Prefix(""),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_VersionsAndDays(2, 30),
								"noncurrent_version_transition": checkNoncurrentVersionTransitions(
									checkNoncurrentVersionTransition_Days(30, "STANDARD_IA"),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":     checkTransitions(),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName, "all_storage_classes_128K"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_None(),
							names.AttrFilter:                    checkFilter_Prefix(""),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_VersionsAndDays(2, 30),
							"noncurrent_version_transition": checkNoncurrentVersionTransitions(
								checkNoncurrentVersionTransition_Days(30, "STANDARD_IA"),
							),
							names.AttrPrefix: knownvalue.StringExact(""),
							names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":     checkTransitions(),
						}),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_None(),
								names.AttrFilter:                    checkFilter_Prefix(""),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_VersionsAndDays(2, 30),
								"noncurrent_version_transition": checkNoncurrentVersionTransitions(
									checkNoncurrentVersionTransition_Days(30, "STANDARD_IA"),
								),
								names.AttrPrefix: knownvalue.StringExact(""),
								names.AttrStatus: knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":     checkTransitions(),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_migrateFromBucket_noChange(t *testing.T) {
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
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.days", "0"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", acctest.CtTrue),
				),
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_migrateNoChange(rName),
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
							names.AttrFilter:                    checkFilter_None(),
							names.AttrID:                        knownvalue.StringExact("id1"),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_DeleteMarker(true),
								names.AttrFilter:                    checkFilter_None(),
								names.AttrID:                        knownvalue.StringExact("id1"),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact("path1/"),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_migrateFromBucket_withChange(t *testing.T) {
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
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.days", "0"),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(bucketResourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", acctest.CtTrue),
				),
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_migrateChange(rName),
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
							"expiration":                        checkExpiration_DeleteMarker(false),
							names.AttrFilter:                    checkFilter_None(),
							names.AttrID:                        knownvalue.StringExact("id1"),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact("path1/"),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusDisabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("all_storage_classes_128K")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_DeleteMarker(false),
								names.AttrFilter:                    checkFilter_None(),
								names.AttrID:                        knownvalue.StringExact("id1"),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact("path1/"),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusDisabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23884.
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/38551.
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
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThanAndPrefix(rName, "prefix1", 300),
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
									"object_size_greater_than": knownvalue.Int64Exact(300),
									names.AttrPrefix:           knownvalue.StringExact("prefix1"),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Days(90),
								names.AttrFilter: checkFilter_And(
									checkAnd(map[string]knownvalue.Check{
										"object_size_greater_than": knownvalue.Int64Exact(300),
										names.AttrPrefix:           knownvalue.StringExact("prefix1"),
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
				},
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThanAndPrefix(rName, "prefix1", 0),
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
									names.AttrPrefix:           knownvalue.StringExact("prefix1"),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Days(90),
								names.AttrFilter: checkFilter_And(
									checkAnd(map[string]knownvalue.Check{
										"object_size_greater_than": knownvalue.Int64Exact(0),
										"object_size_less_than":    knownvalue.Int64Exact(0),
										names.AttrPrefix:           knownvalue.StringExact("prefix1"),
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
				},
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_filterPrefix(rName, "prefix2"),
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
							names.AttrFilter:                    checkFilter_Prefix("prefix2"),
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
								"expiration":                        checkExpiration_Days(90),
								names.AttrFilter:                    checkFilter_Prefix("prefix2"),
								names.AttrID:                        knownvalue.StringExact(rName),
								"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
								"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
								names.AttrPrefix:                    knownvalue.StringExact(""),
								names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
								"transition":                        checkTransitions(),
							}),
						})),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_directoryBucket(t *testing.T) {
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
				Config: testAccBucketLifecycleConfigurationConfig_directoryBucket(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_directory_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
							"expiration":                        checkExpiration_Days(365),
							names.AttrFilter:                    checkFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Days(365),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_transitionDefaultMinimumObjectSize_update(t *testing.T) {
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
				Config: testAccBucketLifecycleConfigurationConfig_transitionDefaultMinimumObjectSize(rName, "varies_by_storage_class"),
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
							names.AttrFilter:                    checkFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("varies_by_storage_class")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Days(365),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_transitionDefaultMinimumObjectSize(rName, "all_storage_classes_128K"),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Days(365),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_transitionDefaultMinimumObjectSize_remove(t *testing.T) {
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
				Config: testAccBucketLifecycleConfigurationConfig_transitionDefaultMinimumObjectSize(rName, "varies_by_storage_class"),
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
							names.AttrFilter:                    checkFilter_Prefix("prefix/"),
							names.AttrID:                        knownvalue.StringExact(rName),
							"noncurrent_version_expiration":     checkNoncurrentVersionExpiration_None(),
							"noncurrent_version_transition":     checkNoncurrentVersionTransitions(),
							names.AttrPrefix:                    knownvalue.StringExact(""),
							names.AttrStatus:                    knownvalue.StringExact(tfs3.LifecycleRuleStatusEnabled),
							"transition":                        checkTransitions(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transition_default_minimum_object_size"), knownvalue.StringExact("varies_by_storage_class")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"abort_incomplete_multipart_upload": checkAbortIncompleteMultipartUpload_None(),
								"expiration":                        checkExpiration_Days(365),
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
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
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
							"expiration":                        checkExpiration_Days(365),
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
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBucketLifecycleConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

			if rs.Type != "aws_s3_bucket_lifecycle_configuration" {
				continue
			}

			bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			if tfs3.IsDirectoryBucket(bucket) {
				conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
			}

			_, err = tfs3.FindBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)

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

		if tfs3.IsDirectoryBucket(bucket) {
			conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
		}

		_, err = tfs3.FindBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func checkAbortIncompleteMultipartUpload_None() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{})
}

func checkAbortIncompleteMultipartUpload_Days(days int64) knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"days_after_initiation": knownvalue.Int64Exact(days),
		}),
	})
}

func checkExpiration_None() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{})
}

func checkExpiration_Date(date string) knownvalue.Check {
	checks := expirationDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"date": knownvalue.StringExact(date),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkExpiration_Days(days int32) knownvalue.Check {
	checks := expirationDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"days": knownvalue.Int32Exact(days),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkExpiration_DeleteMarker(marker bool) knownvalue.Check {
	checks := expirationDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"expired_object_delete_marker": knownvalue.Bool(marker),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkExpiration_Empty() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			expirationDefaults(),
		),
	})
}

func expirationDefaults() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"date":                         knownvalue.Null(),
		"days":                         knownvalue.Int32Exact(0),
		"expired_object_delete_marker": knownvalue.Bool(false),
	}
}

func checkFilter_None() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{})
}

func checkFilter_And(check knownvalue.Check) knownvalue.Check {
	checks := filterDefaults()
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

func checkFilter_GreaterThan(size int64) knownvalue.Check {
	checks := filterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"object_size_greater_than": knownvalue.Int64Exact(size),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkFilter_LessThan(size int64) knownvalue.Check {
	checks := filterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"object_size_less_than": knownvalue.Int64Exact(size),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkFilter_Prefix(prefix string) knownvalue.Check {
	checks := filterDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		names.AttrPrefix: knownvalue.StringExact(prefix),
	})
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(
			checks,
		),
	})
}

func checkFilter_Tag(key, value string) knownvalue.Check {
	checks := filterDefaults()
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

func filterDefaults() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"and":                      knownvalue.ListExact([]knownvalue.Check{}),
		"object_size_greater_than": knownvalue.Null(),
		"object_size_less_than":    knownvalue.Null(),
		names.AttrPrefix:           knownvalue.Null(),
		"tag":                      knownvalue.ListExact([]knownvalue.Check{}),
	}
}

func checkAnd(attrChecks map[string]knownvalue.Check) knownvalue.Check {
	checks := andDefaults()
	maps.Copy(checks, attrChecks)
	return knownvalue.ObjectExact(
		checks,
	)
}

func andDefaults() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"object_size_greater_than": knownvalue.Int64Exact(0),
		"object_size_less_than":    knownvalue.Int64Exact(0),
		names.AttrPrefix:           knownvalue.StringExact(""),
		names.AttrTags:             knownvalue.Null(),
	}
}

func checkNoncurrentVersionExpiration_None() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{})
}

func checkNoncurrentVersionExpiration_Days(days int32) knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"newer_noncurrent_versions": knownvalue.Null(),
			"noncurrent_days":           knownvalue.Int32Exact(days),
		}),
	})
}

func checkNoncurrentVersionExpiration_VersionsAndDays(versions, days int32) knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"newer_noncurrent_versions": knownvalue.Int32Exact(versions),
			"noncurrent_days":           knownvalue.Int32Exact(days),
		}),
	})
}

func checkNoncurrentVersionTransitions(checks ...knownvalue.Check) knownvalue.Check {
	return knownvalue.SetExact(checks)
}

func checkNoncurrentVersionTransition_Days(days int32, class types.TransitionStorageClass) knownvalue.Check {
	return knownvalue.ObjectExact(map[string]knownvalue.Check{
		"newer_noncurrent_versions": knownvalue.Null(),
		"noncurrent_days":           knownvalue.Int32Exact(days),
		names.AttrStorageClass:      tfknownvalue.StringExact(class),
	})
}

func checkTransitions(checks ...knownvalue.Check) knownvalue.Check {
	return knownvalue.SetExact(checks)
}

func checkTransition_Date(date string, class types.TransitionStorageClass) knownvalue.Check {
	checks := transitionDefaults(class)
	maps.Copy(checks, map[string]knownvalue.Check{
		"date": knownvalue.StringExact(date),
	})
	return knownvalue.ObjectExact(
		checks,
	)
}

func checkTransition_Days(days int64, class types.TransitionStorageClass) knownvalue.Check {
	checks := transitionDefaults(class)
	maps.Copy(checks, map[string]knownvalue.Check{
		"days": knownvalue.Int64Exact(days),
	})
	return knownvalue.ObjectExact(
		checks,
	)
}

func checkTransition_StorageClass(class types.TransitionStorageClass) knownvalue.Check {
	checks := transitionDefaults(class)
	maps.Copy(checks, map[string]knownvalue.Check{
		"days": knownvalue.Int64Exact(0),
	})
	return knownvalue.ObjectExact(
		checks,
	)
}

func transitionDefaults(class types.TransitionStorageClass) map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"date":                 knownvalue.Null(),
		"days":                 knownvalue.Null(),
		names.AttrStorageClass: tfknownvalue.StringExact(class),
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

    filter {
      prefix = "prefix/"
    }
  }
}
`, rName)
}

func testAccBucketLifecycleConfigurationConfig_rule_NoFilterOrPrefix(rName string) string {
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

    filter {
      prefix = "prefix/"
    }
  }
}
`, rName, status)
}

func testAccBucketLifecycleConfigurationConfig_filterWithPrefix(rName, date, prefix string) string {
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

    filter {
      prefix = %[3]q
    }
  }
}
`, rName, date, prefix)
}

func testAccBucketLifecycleConfigurationConfig_filterWithEmptyPrefix(rName, date string) string {
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

    filter {
      prefix = ""
    }
  }
}
`, rName, date)
}

func testAccBucketLifecycleConfigurationConfig_rulePrefix(rName, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id = %[1]q

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

    filter {
      prefix = "prefix/"
    }
  }
}
`, rName, expired)
}

func testAccBucketLifecycleConfigurationConfig_ruleExpirationEmptyBlock(rName, prefix string) string {
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

    filter {
      prefix = %[2]q
    }
  }
}
`, rName, prefix)
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

    filter {
      prefix = "prefix/"
    }
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

func testAccBucketLifecycleConfigurationConfig_nonCurrentVersionExpiration(rName string, nonCurrentDays int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    filter {
      prefix = "config/"
    }

    noncurrent_version_expiration {
      noncurrent_days = %[2]d
    }

    status = "Enabled"
  }
}
`, rName, nonCurrentDays)
}

func testAccBucketLifecycleConfigurationConfig_nonCurrentVersionTransition(rName string, standardDays, glacierDays int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    filter {
      prefix = "config/"
    }

    noncurrent_version_transition {
      noncurrent_days = %[2]d
      storage_class   = "STANDARD_IA"
    }

    noncurrent_version_transition {
      noncurrent_days = %[3]d
      storage_class   = "GLACIER"
    }

    status = "Enabled"
  }
}
`, rName, standardDays, glacierDays)
}

func testAccBucketLifecycleConfigurationConfig_transitionStorageClassOnly(rName string, storageClass types.TransitionStorageClass) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    filter {
      prefix = "prefix/"
    }

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

func testAccBucketLifecycleConfigurationConfig_zeroDaysTransition(rName string, storageClass types.TransitionStorageClass) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    filter {
      prefix = "prefix/"
    }

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

func testAccBucketLifecycleConfigurationConfig_dateTransition(rName, transitionDate string, storageClass types.TransitionStorageClass) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    filter {
      prefix = "prefix/"
    }

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

func testAccBucketLifecycleConfigurationConfig_emptyFilterNonCurrentVersions(rName, transitionDefaultMinimumObjectSize string) string {
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

  transition_default_minimum_object_size = %[2]q
}
`, rName, transitionDefaultMinimumObjectSize)
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

func testAccBucketLifecycleConfigurationConfig_filterObjectSizeGreaterThanAndPrefix(rName, prefix string, objectSizeGT int) string {
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
        object_size_greater_than = %[3]d
        prefix                   = %[2]q
      }
    }

    status = "Enabled"
  }
}`, rName, prefix, objectSizeGT)
}

func testAccBucketLifecycleConfigurationConfig_filterWithPrefixAndTags(rName, date, prefix string) string {
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

    filter {
      and {
        prefix = %[3]q
        tags = {
          key1 = "value1"
        }
      }
    }
  }
}
`, rName, date, prefix)
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
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(rName), fmt.Sprintf(`
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

    filter {
      prefix = "prefix/"
    }
  }
}
`, rName))
}

func testAccBucketLifecycleConfigurationConfig_transitionDefaultMinimumObjectSize(rName, transitionDefaultMinimumObjectSize string) string {
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

    filter {
      prefix = "prefix/"
    }
  }

  transition_default_minimum_object_size = %[2]q
}
`, rName, transitionDefaultMinimumObjectSize)
}
