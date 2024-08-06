// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccUserProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_user_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_profile_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_id", "aws_sagemaker_domain.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", acctest.Ct0),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`user-profile/.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "home_efs_file_system_uid"),
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

func testAccUserProfile_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_user_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserProfileConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccUserProfileConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccUserProfile_tensorboardAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_user_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_tensorBoardAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
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

func testAccUserProfile_tensorboardAppSettingsWithImage(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_user_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_tensorBoardAppSettingsImage(rName, "ml.t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.sagemaker_image_arn", "aws_sagemaker_image.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserProfileConfig_tensorBoardAppSettingsImage(rName, "ml.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.small"),
					resource.TestCheckResourceAttrPair(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.sagemaker_image_arn", "aws_sagemaker_image.test", names.AttrARN),
				),
			},
		},
	})
}

func testAccUserProfile_kernelGatewayAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_user_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_kernelGatewayAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
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

func testAccUserProfile_kernelGatewayAppSettings_lifecycleconfig(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_user_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_kernelGatewayAppSettingsLifecycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.0.lifecycle_config_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.lifecycle_config_arn", "aws_sagemaker_studio_lifecycle_config.test", names.AttrARN),
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

func testAccUserProfile_kernelGatewayAppSettings_imageconfig(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var domain sagemaker.DescribeUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_user_profile.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_kernelGatewayAppSettingsImage(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.0.lifecycle_config_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.sagemaker_image_version_arn", "aws_sagemaker_image_version.test", names.AttrARN),
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

func testAccUserProfile_codeEditorAppSettings_customImage(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var domain sagemaker.DescribeUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_user_profile.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_codeEditorAppSettingsImage(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.code_editor_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.code_editor_app_settings.0.lifecycle_config_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.code_editor_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.code_editor_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "user_settings.0.code_editor_app_settings.0.default_resource_spec.0.sagemaker_image_version_arn", "aws_sagemaker_image_version.test", names.AttrARN),
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

func testAccUserProfile_jupyterServerAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_user_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_jupyterServerAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.jupyter_server_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.jupyter_server_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.jupyter_server_app_settings.0.default_resource_spec.0.instance_type", "system"),
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

func testAccUserProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_user_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &domain),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceUserProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_user_profile" {
				continue
			}

			domainID := rs.Primary.Attributes["domain_id"]
			userProfileName := rs.Primary.Attributes["user_profile_name"]

			userProfile, err := tfsagemaker.FindUserProfileByName(ctx, conn, domainID, userProfileName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker User Profile (%s): %w", rs.Primary.ID, err)
			}

			userProfileArn := aws.StringValue(userProfile.UserProfileArn)
			if userProfileArn == rs.Primary.ID {
				return fmt.Errorf("SageMaker User Profile %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckUserProfileExists(ctx context.Context, n string, userProfile *sagemaker.DescribeUserProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		domainID := rs.Primary.Attributes["domain_id"]
		userProfileName := rs.Primary.Attributes["user_profile_name"]

		resp, err := tfsagemaker.FindUserProfileByName(ctx, conn, domainID, userProfileName)
		if err != nil {
			return err
		}

		*userProfile = *resp

		return nil
	}
}

func testAccUserProfileConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccUserProfileConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccUserProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q
}
`, rName))
}

func testAccUserProfileConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccUserProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccUserProfileConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccUserProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccUserProfileConfig_tensorBoardAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccUserProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    tensor_board_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }
}
`, rName))
}

func testAccUserProfileConfig_tensorBoardAppSettingsImage(rName, instanceType string) string {
	return acctest.ConfigCompose(testAccUserProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn
}

resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    tensor_board_app_settings {
      default_resource_spec {
        instance_type       = %[2]q
        sagemaker_image_arn = aws_sagemaker_image.test.arn
      }
    }
  }
}
`, rName, instanceType))
}

func testAccUserProfileConfig_jupyterServerAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccUserProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_server_app_settings {
      default_resource_spec {
        instance_type = "system"
      }
    }
  }
}
`, rName))
}

func testAccUserProfileConfig_kernelGatewayAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccUserProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }
}
`, rName))
}

func testAccUserProfileConfig_kernelGatewayAppSettingsLifecycle(rName string) string {
	return acctest.ConfigCompose(testAccUserProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_studio_lifecycle_config" "test" {
  studio_lifecycle_config_name     = %[1]q
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")
}

resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      default_resource_spec {
        instance_type        = "ml.t3.micro"
        lifecycle_config_arn = aws_sagemaker_studio_lifecycle_config.test.arn
      }

      lifecycle_config_arns = [aws_sagemaker_studio_lifecycle_config.test.arn]
    }
  }
}
`, rName))
}

func testAccUserProfileConfig_kernelGatewayAppSettingsImage(rName, baseImage string) string {
	return acctest.ConfigCompose(testAccUserProfileConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_image_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[2]q

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      default_resource_spec {
        instance_type               = "ml.t3.micro"
        sagemaker_image_version_arn = aws_sagemaker_image_version.test.arn
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, baseImage))
}

func testAccUserProfileConfig_codeEditorAppSettingsImage(rName, baseImage string) string {
	return acctest.ConfigCompose(testAccUserProfileConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_image_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[2]q

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    code_editor_app_settings {
      default_resource_spec {
        instance_type               = "ml.t3.micro"
        sagemaker_image_version_arn = aws_sagemaker_image_version.test.arn
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, baseImage))
}
