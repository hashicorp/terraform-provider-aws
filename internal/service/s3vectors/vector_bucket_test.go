// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3vectors_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3vectors/types"
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
	tfs3vectors "github.com/hashicorp/terraform-provider-aws/internal/service/s3vectors"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3VectorsVectorBucket_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.VectorBucket
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3vectors_vector_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3VectorsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorBucketExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreationTime), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"kms_key_arn": knownvalue.Null(),
							"sse_type":    tfknownvalue.StringExact(awstypes.SseTypeAes256),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vector_bucket_arn"), tfknownvalue.RegionalARNRegexp("s3vectors", regexache.MustCompile(`bucket/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vector_bucket_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						"vector_bucket_arn": tfknownvalue.RegionalARNRegexp("s3vectors", regexache.MustCompile(`bucket/.+`)),
					}),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "vector_bucket_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "vector_bucket_arn",
			},
		},
	})
}

func TestAccS3VectorsVectorBucket_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.VectorBucket
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3vectors_vector_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3VectorsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorBucketExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3vectors.ResourceVectorBucket, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3VectorsVectorBucket_encryptionConfigurationAES256(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.VectorBucket
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3vectors_vector_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3VectorsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorBucketConfig_encryptionConfigurationAES256(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorBucketExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"kms_key_arn": knownvalue.Null(),
							"sse_type":    tfknownvalue.StringExact(awstypes.SseTypeAes256),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "vector_bucket_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "vector_bucket_arn",
			},
			{
				Config: testAccVectorBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorBucketExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccS3VectorsVectorBucket_encryptionConfigurationKMS(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.VectorBucket
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3vectors_vector_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3VectorsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorBucketConfig_encryptionConfigurationKMS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorBucketExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"kms_key_arn": knownvalue.NotNull(),
							"sse_type":    tfknownvalue.StringExact(awstypes.SseTypeAwsKms),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "vector_bucket_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "vector_bucket_arn",
			},
		},
	})
}

func testAccCheckVectorBucketDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3VectorsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3vectors_vector_bucket" {
				continue
			}

			_, err := tfs3vectors.FindVectorBucketByARN(ctx, conn, rs.Primary.Attributes["vector_bucket_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Vectors Vector Bucket %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVectorBucketExists(ctx context.Context, n string, v *awstypes.VectorBucket) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3VectorsClient(ctx)

		output, err := tfs3vectors.FindVectorBucketByARN(ctx, conn, rs.Primary.Attributes["vector_bucket_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return err
	}
}

func testAccVectorBucketConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3vectors_vector_bucket" "test" {
  vector_bucket_name = %[1]q
}
`, rName)
}

func testAccVectorBucketConfig_encryptionConfigurationAES256(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3vectors_vector_bucket" "test" {
  vector_bucket_name = %[1]q

  encryption_configuration {
    sse_type = "AES256"
  }
}
`, rName)
}

func testAccVectorBucketConfig_encryptionConfigurationKMS(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3vectors_vector_bucket" "test" {
  vector_bucket_name = %[1]q

  encryption_configuration {
    kms_key_arn = aws_kms_key.test.arn
    sse_type    = "aws:kms"
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName)
}
