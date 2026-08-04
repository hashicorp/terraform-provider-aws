// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestImageVersionFromARN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arn  string
		want int32
	}{
		{
			name: "valid ARN with version",
			arn:  "arn:aws:sagemaker:us-west-2:123456789012:image-version/my-image/42", //lintignore:AWSAT003,AWSAT005
			want: 42,
		},
		{
			name: "valid ARN with large version",
			arn:  "arn:aws:sagemaker:us-west-2:123456789012:image-version/my-image/999999", //lintignore:AWSAT003,AWSAT005
			want: 999999,
		},
		{
			name: "invalid ARN - too few parts",
			arn:  "arn:aws:sagemaker:us-west-2:123456789012", //lintignore:AWSAT003,AWSAT005
			want: 0,
		},
		{
			name: "invalid ARN - non-numeric version",
			arn:  "arn:aws:sagemaker:us-west-2:123456789012:image-version/my-image/latest", //lintignore:AWSAT003,AWSAT005
			want: 0,
		},
		{
			name: "empty ARN",
			arn:  "",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tfsagemaker.ImageVersionFromARN(tt.arn); got != tt.want {
				t.Errorf("ImageVersionFromARN(%q) = %d, want %d", tt.arn, got, tt.want)
			}
		})
	}
}

const (
	// imageVersionBaseImageEnvVar is the environment variable which must be
	// set to an ECR image URI for certain acceptance tests to run
	//
	// Follow this guide to set up a private ECR repository and push a simple
	// "hello world" image to it:
	// https://docs.aws.amazon.com/AmazonECR/latest/userguide/getting-started-cli.html
	imageVersionBaseImageEnvVar = "SAGEMAKER_IMAGE_VERSION_BASE_IMAGE"

	// imageVersionConcurrentCountEnvVar is the environment variable which can be
	// set to control the number of concurrent image versions created in the test.
	//
	// Defaults to 10 if not set.
	imageVersionConcurrentCountEnvVar = "SAGEMAKER_IMAGE_VERSION_CONCURRENT_COUNT"
)

func TestAccSageMakerImageVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "image_arn", "sagemaker", fmt.Sprintf("image/%s", rName)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("image-version/%s/1", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "container_image"),
					resource.TestCheckResourceAttr(resourceName, "horovod", acctest.CtFalse),
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

func TestAccSageMakerImageVersion_update(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdate := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_full(rName, baseImage, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "image_arn", "sagemaker", fmt.Sprintf("image/%s", rName)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("image-version/%s/1", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "container_image"),
					resource.TestCheckResourceAttr(resourceName, "horovod", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "processor", "CPU"),
					resource.TestCheckResourceAttr(resourceName, "vendor_guidance", "STABLE"),
					resource.TestCheckResourceAttr(resourceName, "release_notes", rName),
					resource.TestCheckResourceAttr(resourceName, "job_type", "TRAINING"),
					resource.TestCheckResourceAttr(resourceName, "ml_framework", "TensorFlow 1.1"),
					resource.TestCheckResourceAttr(resourceName, "programming_lang", "Python 3.8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImageVersionConfig_full(rName, baseImage, rNameUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "image_arn", "sagemaker", fmt.Sprintf("image/%s", rName)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("image-version/%s/1", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "container_image"),
					resource.TestCheckResourceAttr(resourceName, "horovod", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "processor", "CPU"),
					resource.TestCheckResourceAttr(resourceName, "vendor_guidance", "STABLE"),
					resource.TestCheckResourceAttr(resourceName, "release_notes", rNameUpdate),
					resource.TestCheckResourceAttr(resourceName, "job_type", "TRAINING"),
					resource.TestCheckResourceAttr(resourceName, "ml_framework", "TensorFlow 1.1"),
					resource.TestCheckResourceAttr(resourceName, "programming_lang", "Python 3.8"),
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

func TestAccSageMakerImageVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsagemaker.ResourceImageVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccSageMakerImageVersion_Disappears_image(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	imageResourceName := "aws_sagemaker_image.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsagemaker.ResourceImage(), imageResourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(imageResourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// TestAccSageMakerImageVersion_multiple verifies multiple image versions by the
// same name can co-exist in the same configuration without overwriting one another
//
// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/40597
func TestAccSageMakerImageVersion_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	var image, imageV2 sagemaker.DescribeImageVersionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	resourceNameV2 := "aws_sagemaker_image_version.test_v2"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_multiple(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					testAccCheckImageVersionExists(ctx, t, resourceNameV2, &imageV2),
					resource.TestCheckResourceAttr(resourceNameV2, "image_name", rName),
					resource.TestCheckResourceAttr(resourceNameV2, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceNameV2, names.AttrVersion, "2"),
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

func TestAccSageMakerImageVersion_aliases(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_aliases(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aliases.*", "latest"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aliases.*", "stable"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "0"),
				),
			},
			{
				Config: testAccImageVersionConfig_aliases(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aliases.*", "latest"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aliases.*", "stable"),
				),
			},
		},
	})
}

func TestAccSageMakerImageVersion_upgrade_V5_98_0(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.SageMakerServiceID),
		CheckDestroy: testAccCheckImageVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// After v5.97.0, id was change to a multi-part key
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.97.0",
					},
				},
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, t, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// Testing behavior during highly concurrent image version creation operations
//
// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/44693
func TestAccSageMakerImageVersion_concurrentCreation(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	count := 10
	if v := os.Getenv(imageVersionConcurrentCountEnvVar); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			count = parsed
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_concurrent(rName, baseImage, count),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_sagemaker_image_version.test.0", "image_name", rName),
				),
			},
		},
	})
}

func testAccCheckImageVersionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_image_version" {
				continue
			}

			name := rs.Primary.Attributes["image_name"]
			version := flex.StringValueToInt32Value(rs.Primary.Attributes[names.AttrVersion])
			_, err := tfsagemaker.FindImageVersionByTwoPartKey(ctx, conn, name, version)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker AI Image Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckImageVersionExists(ctx context.Context, t *testing.T, n string, v *sagemaker.DescribeImageVersionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		name := rs.Primary.Attributes["image_name"]
		version := flex.StringValueToInt32Value(rs.Primary.Attributes[names.AttrVersion])
		output, err := tfsagemaker.FindImageVersionByTwoPartKey(ctx, conn, name, version)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccImageVersionConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccImageVersionConfig_basic(rName, baseImage string) string {
	return acctest.ConfigCompose(
		testAccImageVersionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[1]q
}
`, baseImage))
}

func testAccImageVersionConfig_full(rName, baseImage, notes string) string {
	return acctest.ConfigCompose(
		testAccImageVersionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_image_version" "test" {
  image_name       = aws_sagemaker_image.test.id
  base_image       = %[1]q
  job_type         = "TRAINING"
  processor        = "CPU"
  release_notes    = %[2]q
  vendor_guidance  = "STABLE"
  ml_framework     = "TensorFlow 1.1"
  programming_lang = "Python 3.8"
}
`, baseImage, notes))
}

func testAccImageVersionConfig_multiple(rName, baseImage string) string {
	return acctest.ConfigCompose(
		testAccImageVersionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[1]q
}

resource "aws_sagemaker_image_version" "test_v2" {
  image_name = aws_sagemaker_image_version.test.image_name
  base_image = %[1]q
}
`, baseImage))
}

func testAccImageVersionConfig_aliases(rName, baseImage string) string {
	return acctest.ConfigCompose(
		testAccImageVersionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[1]q
  aliases    = ["latest", "stable"]
}
`, baseImage))
}

func testAccImageVersionConfig_concurrent(rName, baseImage string, count int) string {
	return acctest.ConfigCompose(
		testAccImageVersionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_image_version" "test" {
  count      = %[3]d
  image_name = aws_sagemaker_image.test.image_name
  base_image = %[2]q
}
`, rName, baseImage, count))
}
