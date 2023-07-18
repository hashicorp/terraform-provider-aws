// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccS3ControlBucketPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName, "s3-outposts:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3control_bucket.test", "arn"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`s3-outposts:\*`)),
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

func TestAccS3ControlBucketPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName, "s3-outposts:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceBucketPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlBucketPolicy_policy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(rName, "s3-outposts:GetObject"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPolicyExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`s3-outposts:GetObject`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketPolicyConfig_basic(rName, "s3-outposts:PutObject"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPolicyExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`s3-outposts:PutObject`)),
				),
			},
		},
	})
}

func testAccCheckBucketPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_bucket_policy" {
				continue
			}

			parsedArn, err := arn.Parse(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfs3control.FindBucketPolicyByTwoPartKey(ctx, conn, parsedArn.AccountID, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Control Bucket Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Control Bucket Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn(ctx)

		parsedArn, err := arn.Parse(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfs3control.FindBucketPolicyByTwoPartKey(ctx, conn, parsedArn.AccountID, rs.Primary.ID)

		return err
	}
}

func testAccBucketPolicyConfig_basic(rName, action string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_s3control_bucket_policy" "test" {
  bucket = aws_s3control_bucket.test.arn
  policy = jsonencode({
    Id = "testBucketPolicy"
    Statement = [
      {
        Action = %[2]q
        Effect = "Deny"
        Principal = {
          AWS = "*"
        }
        Resource = "${aws_s3control_bucket.test.arn}/object/test"
        Sid      = "st1"
      }
    ]
    Version = "2012-10-17"
  })
}
`, rName, action)
}
