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
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSpace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "space_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_id", "aws_sagemaker_domain.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "space_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "space_sharing_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ownership_settings.#", acctest.Ct0),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`space/.+`)),
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

func testAccSpace_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
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
				Config: testAccSpaceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSpaceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccSpace_customFileSystem(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_customFileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "space_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.custom_file_system.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.custom_file_system.0.efs_file_system.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "space_settings.0.custom_file_system.0.efs_file_system.0.file_system_id", "aws_efs_file_system.test", names.AttrID),
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

func testAccSpace_kernelGatewayAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_kernelGatewayAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "space_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
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

func testAccSpace_kernelGatewayAppSettings_lifecycleconfig(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_kernelGatewayAppSettingsLifecycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "space_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.0.lifecycle_config_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.lifecycle_config_arn", "aws_sagemaker_studio_lifecycle_config.test", names.AttrARN),
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

func testAccSpace_kernelGatewayAppSettings_imageconfig(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_kernelGatewayAppSettingsImage(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "space_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.0.lifecycle_config_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "space_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.sagemaker_image_version_arn", "aws_sagemaker_image_version.test", names.AttrARN),
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

func testAccSpace_storageSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_storageSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "space_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.app_type", "CodeEditor"),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.space_storage_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.space_storage_settings.0.ebs_storage_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.space_storage_settings.0.ebs_storage_settings.0.ebs_volume_size_in_gb", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "space_sharing_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_sharing_settings.0.sharing_type", "Private"),
					resource.TestCheckResourceAttr(resourceName, "ownership_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "ownership_settings.0.owner_user_profile_name", "aws_sagemaker_user_profile.test", "user_profile_name"),
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

func testAccSpace_codeEditorAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_codeEditorAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "space_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.code_editor_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.code_editor_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.code_editor_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
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

func testAccSpace_jupyterLabAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_jupyterLabAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "space_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.app_type", "JupyterLab"),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.jupyter_lab_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.jupyter_lab_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.jupyter_lab_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "space_sharing_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_sharing_settings.0.sharing_type", "Private"),
					resource.TestCheckResourceAttr(resourceName, "ownership_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "ownership_settings.0.owner_user_profile_name", "aws_sagemaker_user_profile.test", "user_profile_name"),
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

func testAccSpace_jupyterServerAppSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_jupyterServerAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "space_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.jupyter_server_app_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.jupyter_server_app_settings.0.default_resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "space_settings.0.jupyter_server_app_settings.0.default_resource_spec.0.instance_type", "system"),
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

func testAccSpace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domain sagemaker.DescribeSpaceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &domain),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceSpace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSpaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_space" {
				continue
			}

			domainID := rs.Primary.Attributes["domain_id"]
			spaceName := rs.Primary.Attributes["space_name"]

			space, err := tfsagemaker.FindSpaceByName(ctx, conn, domainID, spaceName)

			if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker Space (%s): %w", rs.Primary.ID, err)
			}

			spaceArn := aws.StringValue(space.SpaceArn)
			if spaceArn == rs.Primary.ID {
				return fmt.Errorf("SageMaker Space %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckSpaceExists(ctx context.Context, n string, space *sagemaker.DescribeSpaceOutput) resource.TestCheckFunc {
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
		spaceName := rs.Primary.Attributes["space_name"]

		resp, err := tfsagemaker.FindSpaceByName(ctx, conn, domainID, spaceName)
		if err != nil {
			return err
		}

		*space = *resp

		return nil
	}
}

func testAccSpaceConfig_base(rName string) string {
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

  default_space_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}
`, rName))
}

func testAccSpaceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSpaceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q
}
`, rName))
}

func testAccSpaceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccSpaceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccSpaceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccSpaceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccSpaceConfig_storageSettings(rName string) string {
	return acctest.ConfigCompose(testAccSpaceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = "%[1]s-2"
}

resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q

  space_sharing_settings {
    sharing_type = "Private"
  }

  ownership_settings {
    owner_user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  }

  space_settings {
    app_type = "CodeEditor"
    space_storage_settings {
      ebs_storage_settings {
        ebs_volume_size_in_gb = 10
      }
    }
  }
}
`, rName))
}

func testAccSpaceConfig_codeEditorAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccSpaceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = "%[1]s-2"
}

resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q

  space_sharing_settings {
    sharing_type = "Private"
  }

  ownership_settings {
    owner_user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  }

  space_settings {
    app_type = "CodeEditor"
    code_editor_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }
}
`, rName))
}

func testAccSpaceConfig_jupyterLabAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccSpaceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = "%[1]s-2"
}

resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q

  space_sharing_settings {
    sharing_type = "Private"
  }

  ownership_settings {
    owner_user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  }

  space_settings {
    app_type = "JupyterLab"
    jupyter_lab_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }
}
`, rName))
}

func testAccSpaceConfig_jupyterServerAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccSpaceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q

  space_settings {
    jupyter_server_app_settings {
      default_resource_spec {
        instance_type = "system"
      }
    }
  }
}
`, rName))
}

func testAccSpaceConfig_customFileConfig(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_efs(rName), fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = "%[1]s-2"
}

resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q

  space_sharing_settings {
    sharing_type = "Private"
  }

  ownership_settings {
    owner_user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  }

  space_settings {
    app_type = "JupyterLab"
    jupyter_lab_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }

    custom_file_system {
      efs_file_system {
        file_system_id = aws_efs_mount_target.test.file_system_id
      }
    }
  }
}
`, rName))
}

func testAccSpaceConfig_kernelGatewayAppSettings(rName string) string {
	return acctest.ConfigCompose(testAccSpaceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q

  space_settings {
    kernel_gateway_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }
}
`, rName))
}

func testAccSpaceConfig_kernelGatewayAppSettingsLifecycle(rName string) string {
	return acctest.ConfigCompose(testAccSpaceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_studio_lifecycle_config" "test" {
  studio_lifecycle_config_name     = %[1]q
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")
}

resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q

  space_settings {
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

func testAccSpaceConfig_kernelGatewayAppSettingsImage(rName, baseImage string) string {
	return acctest.ConfigCompose(testAccSpaceConfig_base(rName), fmt.Sprintf(`
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

resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = %[1]q

  space_settings {
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
