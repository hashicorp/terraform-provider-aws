// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlAccessPointForDirectoryBucketPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_directory_access_point_policy.test_policy"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketPolicyConfig_basic(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketPolicyExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`s3express:CreateSession`)),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					arn := rs.Primary.Attributes["access_point_arn"]

					if arn == "" {
						return "", fmt.Errorf("missing arn in state")
					}

					resourceID, err := tfs3control.AccessPointForDirectoryBucketCreateResourceIDFromARN(arn)
					if err != nil {
						return "", fmt.Errorf("could not parse access point ARN")
					}

					return resourceID, nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointForDirectoryBucketPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_directory_access_point_policy.test_policy"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketPolicyConfig_basic(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessPointForDirectoryBucketPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointForDirectoryBucketPolicy_disappears_AccessPoint(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_directory_access_point_policy.test_policy"
	accessPointResourceName := "aws_s3_directory_access_point.test_ap"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketPolicyConfig_basic(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessPointForDirectoryBucket(), accessPointResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointForDirectoryBucketPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_directory_access_point_policy.test_policy"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketPolicyConfig_basic(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketPolicyExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`s3express:CreateSession`)),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					arn := rs.Primary.Attributes["access_point_arn"]

					if arn == "" {
						return "", fmt.Errorf("missing arn in state")
					}

					resourceID, err := tfs3control.AccessPointForDirectoryBucketCreateResourceIDFromARN(arn)
					if err != nil {
						return "", fmt.Errorf("could not parse access point ARN")
					}

					return resourceID, nil
				},
				ImportStateVerify: true,
			},
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketPolicyConfig_updated(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketPolicyExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`s3express:CreateSession`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"s3express:DataAccessPointArn"`)),
				),
			},
		},
	})
}

func testAccCheckAccessPointForDirectoryBucketPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_directory_access_point_policy" {
				continue
			}

			name, accountID, err := tfs3control.AccessPointForDirectoryBucketParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3control.FindAccessPointForDirectoryBucketPolicyByTwoPartKey(ctx, conn, accountID, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Access Point for Directory Bucket Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessPointForDirectoryBucketPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		name, accountID, err := tfs3control.AccessPointForDirectoryBucketParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		_, err = tfs3control.FindAccessPointForDirectoryBucketPolicyByTwoPartKey(ctx, conn, accountID, name)

		return err
	}
}

func testAccAccessPointForDirectoryBucketPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAccessPointForDirectoryBucketConfig_basic(rName), `
resource "aws_s3control_directory_access_point_policy" "test_policy" {
  access_point_arn = aws_s3_directory_access_point.test_ap.arn

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      "Action": "s3express:CreateSession"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Resource = "${aws_s3_directory_access_point.test_ap.arn}"
    }]
  })
}
`)
}

func testAccAccessPointForDirectoryBucketPolicyConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccAccessPointForDirectoryBucketConfig_basic(rName), `
resource "aws_s3control_directory_access_point_policy" "test_policy" {
  access_point_arn = aws_s3_directory_access_point.test_ap.arn

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3express:CreateSession"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Resource = "${aws_s3_directory_access_point.test_ap.arn}"
      Condition = {
        StringLike = {
          "s3express:DataAccessPointArn": "${aws_s3_directory_access_point.test_ap.arn}"
        }
      }
    }]
  })
}
`)

}
