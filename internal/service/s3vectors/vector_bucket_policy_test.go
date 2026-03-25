// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3vectors_test

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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3vectors "github.com/hashicorp/terraform-provider-aws/internal/service/s3vectors"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3VectorsVectorBucketPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3vectors_vector_bucket_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3VectorsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorBucketPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorBucketPolicyExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPolicy), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vector_bucket_arn"), knownvalue.NotNull()),
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
				ImportStateVerifyIgnore:              []string{names.AttrPolicy},
			},
		},
	})
}

func TestAccS3VectorsVectorBucketPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3vectors_vector_bucket_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3VectorsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorBucketPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorBucketPolicyExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3vectors.ResourceVectorBucketPolicy, resourceName),
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

func testAccCheckVectorBucketPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3VectorsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3vectors_vector_bucket_policy" {
				continue
			}

			_, err := tfs3vectors.FindVectorBucketPolicyByARN(ctx, conn, rs.Primary.Attributes["vector_bucket_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Vectors Vector Bucket Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVectorBucketPolicyExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3VectorsClient(ctx)

		_, err := tfs3vectors.FindVectorBucketPolicyByARN(ctx, conn, rs.Primary.Attributes["vector_bucket_arn"])

		return err
	}
}

func testAccVectorBucketPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3vectors_vector_bucket" "test" {
  vector_bucket_name = %[1]q
}

resource "aws_s3vectors_vector_bucket_policy" "test" {
  vector_bucket_arn = aws_s3vectors_vector_bucket.test.vector_bucket_arn

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "writePolicy",
  "Statement": [{
    "Sid": "writeStatement",
    "Effect": "Allow",
    "Principal": {
      "AWS": "${data.aws_caller_identity.current.account_id}"
    },
    "Action": [
      "s3vectors:PutVectors"
    ],
    "Resource": "*"
  }]
}
EOF
}

data "aws_caller_identity" "current" {}
`, rName)
}
