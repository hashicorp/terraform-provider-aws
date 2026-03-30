// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_policy.bucket"
	bucketResourceName := "aws_s3_bucket.bucket"

	expectedPolicyTemplate := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[2]s:iam::%[1]s:root"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%[2]s:s3:::%[3]s/*",
        "arn:%[2]s:s3:::%[3]s"
      ]
    }
  ]
}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, t, bucketResourceName, expectedPolicyTemplate, rName),
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

func TestAccS3BucketPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_policy.bucket"
	bucketResourceName := "aws_s3_bucket.bucket"

	expectedPolicyTemplate := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[2]s:iam::%[1]s:root"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%[2]s:s3:::%[3]s/*",
        "arn:%[2]s:s3:::%[3]s"
      ]
    }
  ]
}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, t, bucketResourceName, expectedPolicyTemplate, rName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3.ResourceBucketPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_disappears_bucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.bucket"

	expectedPolicyTemplate := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[2]s:iam::%[1]s:root"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%[2]s:s3:::%[3]s/*",
        "arn:%[2]s:s3:::%[3]s"
      ]
    }
  ]
}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, t, bucketResourceName, expectedPolicyTemplate, rName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3.ResourceBucket(), bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_policyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_policy.bucket"
	bucketResourceName := "aws_s3_bucket.bucket"

	expectedPolicyTemplate1 := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[2]s:iam::%[1]s:root"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%[2]s:s3:::%[3]s/*",
        "arn:%[2]s:s3:::%[3]s"
      ]
    }
  ]
}`

	expectedPolicyTemplate2 := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[2]s:iam::%[1]s:root"
      },
      "Action": [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions"
      ],
      "Resource": [
        "arn:%[2]s:s3:::%[3]s/*",
        "arn:%[2]s:s3:::%[3]s"
      ]
    }
  ]
}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, t, bucketResourceName, expectedPolicyTemplate1, rName),
				),
			},
			{
				Config: testAccBucketPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, t, bucketResourceName, expectedPolicyTemplate2, rName),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccS3BucketPolicy_IAMRoleOrder_policyDoc(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderIAMDoc(rName),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13144
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/20456
func TestAccS3BucketPolicy_IAMRoleOrder_policyDocNotPrincipal(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderIAMDocNotPrincipal(rName),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccS3BucketPolicy_IAMRoleOrder_jsonEncode(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName3 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_policy.bucket"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderJSONEncode(rName1),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderJSONEncode(rName1),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderJSONEncodeOrder2(rName2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderJSONEncode(rName2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderJSONEncodeOrder3(rName3),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderJSONEncode(rName3),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccS3BucketPolicy_migrate_noChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"

	expectedPolicyTemplate := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[2]s:iam::%[1]s:root"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%[2]s:s3:::%[3]s/*",
        "arn:%[2]s:s3:::%[3]s"
      ]
    }
  ]
}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_policy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, t, bucketResourceName, expectedPolicyTemplate, rName),
				),
			},
			{
				Config: testAccBucketPolicyConfig_migrateNoChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, t, bucketResourceName, expectedPolicyTemplate, rName),
				),
			},
		},
	})
}

func TestAccS3BucketPolicy_migrate_withChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"

	expectedPolicyTemplate1 := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[2]s:iam::%[1]s:root"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%[2]s:s3:::%[3]s/*",
        "arn:%[2]s:s3:::%[3]s"
      ]
    }
  ]
}`

	expectedPolicyTemplate2 := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[2]s:iam::%[1]s:root"
      },
      "Action": [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions"
      ],
      "Resource": [
        "arn:%[2]s:s3:::%[3]s/*",
        "arn:%[2]s:s3:::%[3]s"
      ]
    }
  ]
}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_policy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, t, bucketResourceName, expectedPolicyTemplate1, rName),
				),
			},
			{
				Config: testAccBucketPolicyConfig_migrateChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, t, bucketResourceName, expectedPolicyTemplate2, rName),
				),
			},
		},
	})
}

func TestAccS3BucketPolicy_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccDirectoryBucketPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_directoryBucket(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func TestAccS3BucketPolicy_Identity_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_s3_bucket_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccDirectoryBucketPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/BucketPolicy/directory_bucket/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketPolicyExists(ctx, t, resourceName),
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
				ConfigDirectory: config.StaticDirectory("testdata/BucketPolicy/directory_bucket/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				ConfigDirectory: config.StaticDirectory("testdata/BucketPolicy/directory_bucket/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/BucketPolicy/directory_bucket/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},
		},
	})
}

func testAccCheckBucketPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)

			if rs.Type != "aws_s3_bucket_policy" {
				continue
			}

			if tfs3.IsDirectoryBucket(rs.Primary.ID) {
				conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
			}

			_, err := tfs3.FindBucketPolicy(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketHasPolicy(ctx context.Context, t *testing.T, n string, expectedPolicyTemplate string, bucketName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.ID) {
			conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
		}

		policy, err := tfs3.FindBucketPolicy(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		// Policy text must be generated inside a resource.TestCheckFunc in order for
		// the acctest.AccountID(ctx) helper to function properly.
		expectedPolicyText := fmt.Sprintf(expectedPolicyTemplate, acctest.AccountID(ctx), acctest.Partition(), bucketName)
		equivalent, err := awspolicy.PoliciesAreEquivalent(policy, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %w", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, policy)
		}

		return nil
	}
}

func testAccCheckBucketPolicyExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.ID) {
			conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
		}

		_, err := tfs3.FindBucketPolicy(ctx, conn, rs.Primary.ID)
		return err
	}
}

func testAccBucketPolicyConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.bucket.bucket
  policy = data.aws_iam_policy_document.policy.json
}

data "aws_iam_policy_document" "policy" {
  statement {
    effect = "Allow"

    actions = [
      "s3:*",
    ]

    resources = [
      aws_s3_bucket.bucket.arn,
      "${aws_s3_bucket.bucket.arn}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`, bucketName)
}

func testAccBucketPolicyConfig_updated(bucketName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.bucket.bucket
  policy = data.aws_iam_policy_document.policy.json
}

data "aws_iam_policy_document" "policy" {
  statement {
    effect = "Allow"

    actions = [
      "s3:DeleteBucket",
      "s3:ListBucket",
      "s3:ListBucketVersions",
    ]

    resources = [
      aws_s3_bucket.bucket.arn,
      "${aws_s3_bucket.bucket.arn}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`, bucketName)
}

func testAccBucketPolicyIAMRoleOrderConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_service_principal" "current" {
  service_name = "s3"
}

resource "aws_iam_role" "test1" {
  name = "%[1]s-sultan"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = data.aws_service_principal.current.name
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-shepard"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = data.aws_service_principal.current.name
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test3" {
  name = "%[1]s-tritonal"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = data.aws_service_principal.current.name
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test4" {
  name = "%[1]s-artlec"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = data.aws_service_principal.current.name
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test5" {
  name = "%[1]s-cazzette"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = data.aws_service_principal.current.name
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketPolicyConfig_iamRoleOrderIAMDoc(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderConfig_base(rName),
		fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  policy_id = %[1]q

  statement {
    actions = [
      "s3:DeleteBucket",
      "s3:ListBucket",
      "s3:ListBucketVersions",
    ]
    effect = "Allow"
    principals {
      identifiers = [
        aws_iam_role.test2.arn,
        aws_iam_role.test1.arn,
        aws_iam_role.test4.arn,
        aws_iam_role.test3.arn,
        aws_iam_role.test5.arn,
      ]
      type = "AWS"
    }
    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]
  }
}

resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}
`, rName))
}

func testAccBucketPolicyConfig_iamRoleOrderJSONEncode(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderConfig_base(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket

  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Action = [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions",
      ]
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test2.arn,
          aws_iam_role.test1.arn,
          aws_iam_role.test4.arn,
          aws_iam_role.test3.arn,
          aws_iam_role.test5.arn,
        ]
      }

      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName))
}

func testAccBucketPolicyConfig_iamRoleOrderJSONEncodeOrder2(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderConfig_base(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket

  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Action = [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions",
      ]
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test2.arn,
          aws_iam_role.test3.arn,
          aws_iam_role.test5.arn,
          aws_iam_role.test1.arn,
          aws_iam_role.test4.arn,
        ]
      }

      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName))
}

func testAccBucketPolicyConfig_iamRoleOrderJSONEncodeOrder3(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderConfig_base(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket

  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Action = [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions",
      ]
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test4.arn,
          aws_iam_role.test1.arn,
          aws_iam_role.test3.arn,
          aws_iam_role.test5.arn,
          aws_iam_role.test2.arn,
        ]
      }

      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName))
}

func testAccBucketPolicyConfig_iamRoleOrderIAMDocNotPrincipal(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderConfig_base(rName),
		`
data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "DenyInfected"
    actions = [
      "s3:GetObject",
      "s3:PutObjectTagging",
    ]
    effect = "Deny"
    not_principals {
      identifiers = [
        aws_iam_role.test2.arn,
        aws_iam_role.test3.arn,
        aws_iam_role.test4.arn,
        aws_iam_role.test1.arn,
        aws_iam_role.test5.arn,
        data.aws_caller_identity.current.arn,
      ]
      type = "AWS"
    }
    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
    condition {
      test     = "StringEquals"
      variable = "s3:ExistingObjectTag/av-status"
      values   = ["INFECTED"]
    }
  }
}

resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}
`)
}

func testAccBucketPolicyConfig_migrateNoChange(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_iam_policy_document" "policy" {
  statement {
    effect = "Allow"

    actions = [
      "s3:*",
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.policy.json
}
`, rName)
}

func testAccBucketPolicyConfig_migrateChange(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_iam_policy_document" "policy" {
  statement {
    effect = "Allow"

    actions = [
      "s3:DeleteBucket",
      "s3:ListBucket",
      "s3:ListBucketVersions",
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.policy.json
}
`, rName)
}

func testAccBucketPolicyConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(rName), `
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "s3express:*",
    ]

    resources = [
      aws_s3_directory_bucket.test.arn,
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`)
}
