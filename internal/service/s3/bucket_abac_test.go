// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketABAC_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_abac.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketABACDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketABACConfig_basic(rName, string(awstypes.BucketAbacStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketABACExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("abac_status"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrStatus: tfknownvalue.StringExact(awstypes.BucketAbacStatusEnabled),
						}),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("abac_status"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrStatus: tfknownvalue.StringExact(awstypes.BucketAbacStatusEnabled),
							}),
						})),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrBucket),
				ImportStateVerifyIdentifierAttribute: names.AttrBucket,
			},
		},
	})
}

// There is no standard `_disappears` test because deletion of the resource only disables ABAC.
func TestAccS3BucketABAC_disappears_Bucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_abac.test"
	bucketResourceName := "aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketABACDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketABACConfig_basic(rName, string(awstypes.BucketAbacStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketABACExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3.ResourceBucket(), bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccS3BucketABAC_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_abac.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketABACDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketABACConfig_basic(rName, string(awstypes.BucketAbacStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketABACExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("abac_status"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrStatus: tfknownvalue.StringExact(awstypes.BucketAbacStatusEnabled),
						}),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("abac_status"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrStatus: tfknownvalue.StringExact(awstypes.BucketAbacStatusEnabled),
							}),
						})),
					},
				},
			},
			{
				Config: testAccBucketABACConfig_basic(rName, string(awstypes.BucketAbacStatusDisabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketABACExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("abac_status"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrStatus: knownvalue.StringExact(string(awstypes.BucketAbacStatusDisabled)),
						}),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("abac_status"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrStatus: knownvalue.StringExact(string(awstypes.BucketAbacStatusDisabled)),
							}),
						})),
					},
				},
			},
			{
				Config: testAccBucketABACConfig_basic(rName, string(awstypes.BucketAbacStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketABACExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("abac_status"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrStatus: tfknownvalue.StringExact(awstypes.BucketAbacStatusEnabled),
						}),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("abac_status"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrStatus: tfknownvalue.StringExact(awstypes.BucketAbacStatusEnabled),
							}),
						})),
					},
				},
			},
		},
	})
}

func TestAccS3BucketABAC_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_abac.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketABACDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketABACConfig_expectedBucketOwner(rName, string(awstypes.BucketAbacStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketABACExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrBucket), "aws_s3_bucket.test", tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("abac_status"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrStatus: tfknownvalue.StringExact(awstypes.BucketAbacStatusEnabled),
						}),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("abac_status"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrStatus: tfknownvalue.StringExact(awstypes.BucketAbacStatusEnabled),
							}),
						})),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrBucket),
				ImportStateVerifyIdentifierAttribute: names.AttrBucket,
				ImportStateVerifyIgnore: []string{
					names.AttrExpectedBucketOwner,
				},
			},
		},
	})
}

func testAccCheckBucketABACDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_abac" {
				continue
			}

			bucket := rs.Primary.Attributes[names.AttrBucket]
			expectedBucketOwner := rs.Primary.Attributes[names.AttrExpectedBucketOwner]

			_, err := tfs3.FindBucketABAC(ctx, conn, bucket, expectedBucketOwner)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.S3, create.ErrActionCheckingDestroyed, tfs3.ResNameBucketABAC, bucket, err)
			}

			return create.Error(names.S3, create.ErrActionCheckingDestroyed, tfs3.ResNameBucketABAC, bucket, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckBucketABACExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.S3, create.ErrActionCheckingExistence, tfs3.ResNameBucketABAC, name, errors.New("not found"))
		}

		bucket := rs.Primary.Attributes[names.AttrBucket]
		expectedBucketOwner := rs.Primary.Attributes[names.AttrExpectedBucketOwner]
		if bucket == "" {
			return create.Error(names.S3, create.ErrActionCheckingExistence, tfs3.ResNameBucketABAC, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)
		if tfs3.IsDirectoryBucket(bucket) {
			conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
		}

		_, err := tfs3.FindBucketABAC(ctx, conn, bucket, expectedBucketOwner)
		if err != nil {
			return create.Error(names.S3, create.ErrActionCheckingExistence, tfs3.ResNameBucketABAC, bucket, err)
		}

		return nil
	}
}

func testAccBucketABACConfig_basic(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_abac" "test" {
  bucket = aws_s3_bucket.test.bucket

  abac_status {
    status = %[2]q
  }
}
`, rName, status)
}

func testAccBucketABACConfig_expectedBucketOwner(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_abac" "test" {
  bucket                = aws_s3_bucket.test.bucket
  expected_bucket_owner = data.aws_caller_identity.current.account_id

  abac_status {
    status = %[2]q
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_caller_identity" "current" {}
`, rName, status)
}
