// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, bucketResourceName, expectedPolicyTemplate, rName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, bucketResourceName, expectedPolicyTemplate, rName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_disappears_bucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, bucketResourceName, expectedPolicyTemplate, rName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucket(), bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_policyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, bucketResourceName, expectedPolicyTemplate1, rName),
				),
			},
			{
				Config: testAccBucketPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, bucketResourceName, expectedPolicyTemplate2, rName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderIAMDoc(rName),
			},
			{
				Config:   testAccBucketPolicyConfig_iamRoleOrderIAMDoc(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyConfig_iamRoleOrderIAMDoc(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyConfig_iamRoleOrderIAMDoc(rName),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13144
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/20456
func TestAccS3BucketPolicy_IAMRoleOrder_policyDocNotPrincipal(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderIAMDocNotPrincipal(rName),
			},
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderIAMDocNotPrincipal(rName),
			},
			{
				Config:   testAccBucketPolicyConfig_iamRoleOrderIAMDocNotPrincipal(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyConfig_iamRoleOrderIAMDocNotPrincipal(rName),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccS3BucketPolicy_IAMRoleOrder_jsonEncode(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderJSONEncode(rName1),
			},
			{
				Config:   testAccBucketPolicyConfig_iamRoleOrderJSONEncode(rName1),
				PlanOnly: true,
			},
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderJSONEncodeOrder2(rName2),
			},
			{
				Config:   testAccBucketPolicyConfig_iamRoleOrderJSONEncode(rName2),
				PlanOnly: true,
			},
			{
				Config: testAccBucketPolicyConfig_iamRoleOrderJSONEncodeOrder3(rName3),
			},
			{
				Config:   testAccBucketPolicyConfig_iamRoleOrderJSONEncode(rName3),
				PlanOnly: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_migrate_noChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_policy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, bucketResourceName, expectedPolicyTemplate, rName),
				),
			},
			{
				Config: testAccBucketPolicyConfig_migrateNoChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, bucketResourceName, expectedPolicyTemplate, rName),
				),
			},
		},
	})
}

func TestAccS3BucketPolicy_migrate_withChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_policy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, bucketResourceName, expectedPolicyTemplate1, rName),
				),
			},
			{
				Config: testAccBucketPolicyConfig_migrateChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketHasPolicy(ctx, bucketResourceName, expectedPolicyTemplate2, rName),
				),
			},
		},
	})
}

func TestAccS3BucketPolicy_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
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

func testAccCheckBucketPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_policy" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
			if tfs3.IsDirectoryBucket(rs.Primary.ID) {
				conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
			}

			_, err := tfs3.FindBucketPolicy(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckBucketHasPolicy(ctx context.Context, n string, expectedPolicyTemplate string, bucketName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.ID) {
			conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
		}

		policy, err := tfs3.FindBucketPolicy(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		// Policy text must be generated inside a resource.TestCheckFunc in order for
		// the acctest.AccountID() helper to function properly.
		expectedPolicyText := fmt.Sprintf(expectedPolicyTemplate, acctest.AccountID(), acctest.Partition(), bucketName)
		equivalent, err := awspolicy.PoliciesAreEquivalent(policy, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, policy)
		}

		return nil
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

resource "aws_iam_role" "test1" {
  name = "%[1]s-sultan"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
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
        Service = "s3.${data.aws_partition.current.dns_suffix}"
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
        Service = "s3.${data.aws_partition.current.dns_suffix}"
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
        Service = "s3.${data.aws_partition.current.dns_suffix}"
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
        Service = "s3.${data.aws_partition.current.dns_suffix}"
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
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), `
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
