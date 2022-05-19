package s3_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccS3BucketPolicy_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	partition := acctest.Partition()

	expectedPolicyText := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%s:s3:::%s/*",
        "arn:%s:s3:::%s"
      ]
    }
  ]
}`, partition, name, partition, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("aws_s3_bucket.bucket"),
					testAccCheckBucketHasPolicy("aws_s3_bucket.bucket", expectedPolicyText),
				),
			},
			{
				ResourceName:      "aws_s3_bucket_policy.bucket",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_disappears(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	partition := acctest.Partition()
	bucketResourceName := "aws_s3_bucket.bucket"
	resourceName := "aws_s3_bucket_policy.bucket"

	expectedPolicyText := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%s:s3:::%s/*",
        "arn:%s:s3:::%s"
      ]
    }
  ]
}`, partition, name, partition, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					testAccCheckBucketHasPolicy(bucketResourceName, expectedPolicyText),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_disappears_bucket(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	partition := acctest.Partition()
	bucketResourceName := "aws_s3_bucket.bucket"

	expectedPolicyText := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%s:s3:::%s/*",
        "arn:%s:s3:::%s"
      ]
    }
  ]
}`, partition, name, partition, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					testAccCheckBucketHasPolicy(bucketResourceName, expectedPolicyText),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucket(), bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_policyUpdate(t *testing.T) {
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	partition := acctest.Partition()

	expectedPolicyText1 := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%[1]s:s3:::%[2]s/*",
        "arn:%[1]s:s3:::%[2]s"
      ]
    }
  ]
}`, partition, name)

	expectedPolicyText2 := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions"
      ],
      "Resource": [
        "arn:%[1]s:s3:::%[2]s/*",
        "arn:%[1]s:s3:::%[2]s"
      ]
    }
  ]
}`, partition, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("aws_s3_bucket.bucket"),
					testAccCheckBucketHasPolicy("aws_s3_bucket.bucket", expectedPolicyText1),
				),
			},

			{
				Config: testAccBucketPolicyConfig_updated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("aws_s3_bucket.bucket"),
					testAccCheckBucketHasPolicy("aws_s3_bucket.bucket", expectedPolicyText2),
				),
			},

			{
				ResourceName:      "aws_s3_bucket_policy.bucket",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccS3BucketPolicy_IAMRoleOrder_policyDoc(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13144
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/20456
func TestAccS3BucketPolicy_IAMRoleOrder_policyDocNotPrincipal(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config: testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccS3BucketPolicy_IAMRoleOrder_jsonEncode(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName),
				PlanOnly: true,
			},
			{
				Config: testAccBucketPolicyIAMRoleOrderJSONEncodeOrder2Config(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName2),
				PlanOnly: true,
			},
			{
				Config: testAccBucketPolicyIAMRoleOrderJSONEncodeOrder3Config(rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName3),
				PlanOnly: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_migrate_noChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_policy.test"
	bucketResourceName := "aws_s3_bucket.test"
	partition := acctest.Partition()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withPolicy(rName, partition),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					testAccCheckBucketPolicy(bucketResourceName, testAccBucketPolicy(rName, partition)),
				),
			},
			{
				Config: testAccBucketPolicy_Migrate_NoChangeConfig(rName, partition),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					testAccCheckBucketPolicy(resourceName, testAccBucketPolicy(rName, partition)),
				),
			},
		},
	})
}

func TestAccS3BucketPolicy_migrate_withChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_policy.test"
	bucketResourceName := "aws_s3_bucket.test"
	partition := acctest.Partition()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withPolicy(rName, partition),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					testAccCheckBucketPolicy(bucketResourceName, testAccBucketPolicy(rName, partition)),
				),
			},
			{
				Config: testAccBucketPolicy_Migrate_WithChangeConfig(rName, partition),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketPolicy(resourceName, testAccBucketPolicyUpdated(rName, partition)),
				),
			},
		},
	})
}

func testAccCheckBucketHasPolicy(n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		policy, err := conn.GetBucketPolicy(&s3.GetBucketPolicyInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("GetBucketPolicy error: %v", err)
		}

		actualPolicyText := *policy.Policy

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccBucketPolicyConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  tags = {
    TestName = "TestAccS3BucketPolicy_basic"
  }
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
      identifiers = ["*"]
    }
  }
}
`, bucketName)
}

func testAccBucketPolicyConfig_updated(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  tags = {
    TestName = "TestAccS3BucketPolicy_basic"
  }
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
      identifiers = ["*"]
    }
  }
}
`, bucketName)
}

func testAccBucketPolicyIAMRoleOrderBaseConfig(rName string) string {
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

  tags = {
    TestName = %[1]q
  }
}
`, rName)
}

func testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
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

func testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
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

func testAccBucketPolicyIAMRoleOrderJSONEncodeOrder2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
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

func testAccBucketPolicyIAMRoleOrderJSONEncodeOrder3Config(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
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

func testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
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

func testAccBucketPolicy_Migrate_NoChangeConfig(bucketName, partition string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = %[2]s
}
`, bucketName, strconv.Quote(testAccBucketPolicy(bucketName, partition)))
}

func testAccBucketPolicy_Migrate_WithChangeConfig(bucketName, partition string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = %[2]s
}
`, bucketName, strconv.Quote(testAccBucketPolicyUpdated(bucketName, partition)))
}

func testAccBucketPolicyUpdated(bucketName, partition string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:%[1]s:s3:::%[2]s/*"
    }
  ]
}`, partition, bucketName)
}
