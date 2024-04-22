// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/shield"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfshield "github.com/hashicorp/terraform-provider-aws/internal/service/shield"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccDRTAccessLogBucketAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_drt_access_log_bucket_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckLogBucket(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDRTAccessLogBucketAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDRTAccessLogBucketAssociationConfig_basic(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDRTAccessLogBucketAssociationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"role_arn_association_id"},
			},
		},
	})
}

func testAccDRTAccessLogBucketAssociation_multiBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var buckets = []string{}
	for i := 0; i < 2; i++ {
		buckets = append(buckets, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
	}
	resourceName1 := "aws_shield_drt_access_log_bucket_association.test1"
	resourceName2 := "aws_shield_drt_access_log_bucket_association.test2"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckLogBucket(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDRTAccessLogBucketAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDRTAccessLogBucketAssociationConfig_multibucket(rName, buckets),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDRTAccessLogBucketAssociationExists(ctx, resourceName1),
					testAccCheckDRTAccessLogBucketAssociationExists(ctx, resourceName2),
				),
			},
		},
	})
}

func testAccDRTAccessLogBucketAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_drt_access_log_bucket_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckLogBucket(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDRTAccessLogBucketAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDRTAccessLogBucketAssociationConfig_basic(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDRTAccessLogBucketAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfshield.ResourceDRTAccessLogBucketAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDRTAccessLogBucketAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_shield_drt_access_log_bucket_association" {
				continue
			}

			_, err := tfshield.FindDRTLogBucketAssociation(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Shield DRT Role ARN Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDRTAccessLogBucketAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

		_, err := tfshield.FindDRTLogBucketAssociation(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheckLogBucket(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

	input := &shield.DescribeDRTAccessInput{}
	_, err := conn.DescribeDRTAccess(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDRTAccessLogBucketAssociationConfig_basic(rName string, bucket string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      "Sid" : "",
      "Effect" : "Allow",
      "Principal" : {
        "Service" : "drt.shield.amazonaws.com"
      },
      "Action" : "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSShieldDRTAccessPolicy"
}

resource "aws_shield_protection_group" "test" {
  protection_group_id = %[1]q
  aggregation         = "MAX"
  pattern             = "ALL"
}

resource "aws_shield_drt_access_role_arn_association" "test" {
  role_arn = aws_iam_role.test.arn
}

resource "aws_shield_drt_access_log_bucket_association" "test" {
  log_bucket              = aws_s3_bucket.test.id
  role_arn_association_id = aws_shield_drt_access_role_arn_association.test.id
}
`, rName, bucket)
}

func testAccDRTAccessLogBucketAssociationConfig_multibucket(rName string, buckets []string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test1" {
  bucket = %[2]q
}

resource "aws_s3_bucket" "test2" {
  bucket = %[3]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      "Sid" : "",
      "Effect" : "Allow",
      "Principal" : {
        "Service" : "drt.shield.amazonaws.com"
      },
      "Action" : "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSShieldDRTAccessPolicy"
}

resource "aws_shield_protection_group" "test" {
  protection_group_id = %[1]q
  aggregation         = "MAX"
  pattern             = "ALL"
}

resource "aws_shield_drt_access_role_arn_association" "test" {
  role_arn = aws_iam_role.test.arn
}

resource "aws_shield_drt_access_log_bucket_association" "test1" {
  log_bucket              = aws_s3_bucket.test1.id
  role_arn_association_id = aws_shield_drt_access_role_arn_association.test.id
}

resource "aws_shield_drt_access_log_bucket_association" "test2" {
  log_bucket              = aws_s3_bucket.test2.id
  role_arn_association_id = aws_shield_drt_access_role_arn_association.test.id

  depends_on = [aws_shield_drt_access_log_bucket_association.test1]
}
`, rName, buckets[0], buckets[1])
}
