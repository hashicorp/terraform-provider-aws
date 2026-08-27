// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lambdamicrovms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambdamicrovms "github.com/hashicorp/terraform-provider-aws/internal/service/lambdamicrovms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func checkImageARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNExact("lambda", "microvm-image:"+name)
}

func TestAccLambdaMicroVMsImage_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v lambdamicrovms.GetMicrovmImageOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdamicrovms_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaMicroVMsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.MicrovmImageStateCreated)),
					resource.TestCheckResourceAttr(resourceName, "code_artifact.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`microvm-image:.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore: []string{
					"base_image_arn",
					"base_image_version",
					"build_role_arn",
					"code_artifact",
					"egress_network_connectors",
					"image_version",
				},
			},
		},
	})
}
func TestAccLambdaMicroVMsImage_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v lambdamicrovms.GetMicrovmImageOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdamicrovms_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaMicroVMsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflambdamicrovms.ResourceImage, resourceName),
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

func TestAccLambdaMicroVMsImage_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v lambdamicrovms.GetMicrovmImageOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdamicrovms_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaMicroVMsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_description(rName, "description one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description one"),
				),
			},
			{
				Config: testAccImageConfig_description(rName, "description two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description two"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckImageDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaMicroVMsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambdamicrovms_image" {
				continue
			}

			_, err := tflambdamicrovms.FindImageByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda MicroVMs Image %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckImageExists(ctx context.Context, t *testing.T, n string, v *lambdamicrovms.GetMicrovmImageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaMicroVMsClient(ctx)

		output, err := tflambdamicrovms.FindImageByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).LambdaMicroVMsClient(ctx)

	input := lambdamicrovms.ListMicrovmImagesInput{}

	_, err := conn.ListMicrovmImages(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// The microvm image needs an IAM role and S3 URI where the zip file with the code and Dockerfile is.
// This creates the pre-requisites required for creating a basic microvm image
func testAccImageConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = ["s3:GetObject"]
      Effect   = "Allow"
      Resource = "${aws_s3_bucket.test.arn}/*"
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "code.zip"
  source = "test-fixtures/code.zip"
}
`, rName)
}

func testAccImageConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccImageConfig_base(rName), fmt.Sprintf(`
resource "aws_lambdamicrovms_image" "test" {
  name           = %[1]q
  base_image_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:aws:microvm-image:al2023-1"
  build_role_arn = aws_iam_role.test.arn

  code_artifact {
    uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  }
}
`, rName))
}

func testAccImageConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccImageConfig_base(rName), fmt.Sprintf(`
resource "aws_lambdamicrovms_image" "test" {
  name           = %[1]q
  description    = %[2]q
  base_image_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:aws:microvm-image:al2023-1"
  build_role_arn = aws_iam_role.test.arn

  code_artifact {
    uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  }
}
`, rName, description))
}
