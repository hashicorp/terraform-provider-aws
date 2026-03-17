// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketCORSConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCORSConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": "1",
						"allowed_origins.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, names.AttrExpectedBucketOwner, ""),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cors_rule.0.max_age_seconds",
				},
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCORSConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3.ResourceBucketCorsConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{

			{
				Config: testAccBucketCORSConfigurationConfig_completeSingleRule(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": "1",
						"allowed_methods.#": "3",
						"allowed_origins.#": "1",
						"expose_headers.#":  "1",
						names.AttrID:        rName,
						"max_age_seconds":   "3000",
					}),
				),
			},
			{
				Config: testAccBucketCORSConfigurationConfig_multipleRules(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": "1",
						"allowed_methods.#": "3",
						"allowed_origins.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": "1",
						"allowed_origins.#": "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cors_rule.0.max_age_seconds",
					"cors_rule.1.max_age_seconds",
				},
			},
			{
				Config: testAccBucketCORSConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": "1",
						"allowed_origins.#": "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cors_rule.0.max_age_seconds",
				},
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_SingleRule(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCORSConfigurationConfig_completeSingleRule(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": "1",
						"allowed_methods.#": "3",
						"allowed_origins.#": "1",
						"expose_headers.#":  "1",
						names.AttrID:        rName,
						"max_age_seconds":   "3000",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_headers.*", "*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "POST"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.expose_headers.*", "ETag"),
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

func TestAccS3BucketCORSConfiguration_MultipleRules(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCORSConfigurationConfig_multipleRules(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": "1",
						"allowed_methods.#": "3",
						"allowed_origins.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_headers.*", "*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "POST"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": "1",
						"allowed_origins.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "*"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cors_rule.0.max_age_seconds",
					"cors_rule.1.max_age_seconds",
				},
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_migrate_corsRuleNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_cors(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, t, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_methods.#", "2"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.expose_headers.#", "2"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
			{
				Config: testAccBucketCORSConfigurationConfig_migrateRuleNoChange(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": "1",
						"allowed_methods.#": "2",
						"allowed_origins.#": "1",
						"expose_headers.#":  "2",
						"max_age_seconds":   "3000",
					}),
				),
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_migrate_corsRuleWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_cors(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, t, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_methods.#", "2"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.expose_headers.#", "2"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
			{
				Config: testAccBucketCORSConfigurationConfig_migrateRuleChange(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": "1",
						"allowed_origins.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
				),
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCORSConfigurationConfig_expectedBucketOwner(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cors_rule.0.max_age_seconds",
				},
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_Identity_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				Config: testAccBucketCORSConfigurationConfig_expectedBucketOwner(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner}"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},

			// Step 2: Import command
			{
				Config:            testAccBucketCORSConfigurationConfig_expectedBucketOwner(rName),
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cors_rule.0.max_age_seconds",
				},
			},

			// Step 3: Import block with Import ID
			{
				Config:          testAccBucketCORSConfigurationConfig_expectedBucketOwner(rName),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				Config:          testAccBucketCORSConfigurationConfig_expectedBucketOwner(rName),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Resource Identity was added after v6.9.0
func TestAccS3BucketCORSConfiguration_Identity_ExistingResource_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketCORSConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.9.0",
					},
				},
				Config: testAccBucketCORSConfigurationConfig_expectedBucketOwner(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketCORSConfigurationConfig_expectedBucketOwner(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner}"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},
		},
	})
}

// Resource Identity version 1 was added in version 6.31.0
func TestAccS3BucketCORSConfiguration_Identity_Upgrade_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckBucketCORSConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.30.0",
					},
				},
				Config: testAccBucketCORSConfigurationConfig_expectedBucketOwner(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectHasIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketCORSConfigurationConfig_expectedBucketOwner(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner}"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketCORSConfigurationConfig_directoryBucket(rName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketCORSConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)

			if rs.Type != "aws_s3_bucket_cors_configuration" {
				continue
			}

			bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			if tfs3.IsDirectoryBucket(bucket) {
				conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
			}

			_, err = tfs3.FindCORSRules(ctx, conn, bucket, expectedBucketOwner)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Website Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketCORSConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)
		if tfs3.IsDirectoryBucket(bucket) {
			conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
		}

		_, err = tfs3.FindCORSRules(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketCORSConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}
`, rName)
}

func testAccBucketCORSConfigurationConfig_completeSingleRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST", "DELETE"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["ETag"]
    id              = %[1]q
    max_age_seconds = 3000
  }
}
`, rName)
}

func testAccBucketCORSConfigurationConfig_multipleRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST", "DELETE"]
    allowed_origins = ["https://www.example.com"]
  }

  cors_rule {
    allowed_methods = ["GET"]
    allowed_origins = ["*"]
  }
}
`, rName)
}

func testAccBucketCORSConfigurationConfig_migrateRuleNoChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, rName)
}

func testAccBucketCORSConfigurationConfig_migrateRuleChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}
`, rName)
}

func testAccBucketCORSConfigurationConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.bucket

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}
`)
}

func testAccBucketCORSConfigurationConfig_expectedBucketOwner(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_cors_configuration" "test" {
  bucket                = aws_s3_bucket.test.bucket
  expected_bucket_owner = data.aws_caller_identity.current.account_id

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_caller_identity" "current" {}
`, rName)
}
