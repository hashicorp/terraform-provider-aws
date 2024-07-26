// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func testAccApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_id", "aws_sagemaker_domain.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "user_profile_name", "aws_sagemaker_user_profile.test", "user_profile_name"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`app/.+`)),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "resource_spec.0.sagemaker_image_arn"),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.0.instance_type", "system"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccApp_space(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_space(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "space_name", "aws_sagemaker_space.test", "space_name"),
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

func testAccApp_resourceSpec(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_resourceSpec(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.0.instance_type", "system"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_spec.0.sagemaker_image_arn"),
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

func testAccApp_resourceSpecLifecycle(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_resourceSpecLifecycle(rName, uName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.0.instance_type", "system"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_spec.0.sagemaker_image_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_spec.0.lifecycle_config_arn", "aws_sagemaker_studio_lifecycle_config.test", names.AttrARN),
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

func testAccApp_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &app),
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
				Config: testAccAppConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAppConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccApp_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &app),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_app" {
				continue
			}

			domainID := rs.Primary.Attributes["domain_id"]
			appType := rs.Primary.Attributes["app_type"]
			appName := rs.Primary.Attributes["app_name"]

			var userProfileOrSpaceName string
			if v, ok := rs.Primary.Attributes["user_profile_name"]; ok {
				userProfileOrSpaceName = v
			}
			if v, ok := rs.Primary.Attributes["space_name"]; ok {
				userProfileOrSpaceName = v
			}

			_, err := tfsagemaker.FindAppByName(ctx, conn, domainID, userProfileOrSpaceName, appType, appName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker App (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppExists(ctx context.Context, n string, v *sagemaker.DescribeAppOutput) resource.TestCheckFunc {
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
		appType := rs.Primary.Attributes["app_type"]
		appName := rs.Primary.Attributes["app_name"]

		var userProfileOrSpaceName string
		if v, ok := rs.Primary.Attributes["user_profile_name"]; ok {
			userProfileOrSpaceName = v
		}
		if v, ok := rs.Primary.Attributes["space_name"]; ok {
			userProfileOrSpaceName = v
		}

		output, err := tfsagemaker.FindAppByName(ctx, conn, domainID, userProfileOrSpaceName, appType, appName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAppConfig_base(rName string) string {
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

resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q
}
`, rName))
}

func testAccAppConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAppConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"
}
`, rName))
}

func testAccAppConfig_space(rName string) string {
	return acctest.ConfigCompose(testAccAppConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_space" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = "%[1]s-space"
}

resource "aws_sagemaker_app" "test" {
  domain_id  = aws_sagemaker_domain.test.id
  app_name   = %[1]q
  app_type   = "JupyterServer"
  space_name = aws_sagemaker_space.test.space_name
}
`, rName))
}

func testAccAppConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAppConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAppConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAppConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAppConfig_resourceSpec(rName string) string {
	return acctest.ConfigCompose(testAccAppConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  resource_spec {
    instance_type = "system"
  }
}
`, rName))
}

func testAccAppConfig_resourceSpecLifecycle(rName, uName string) string {
	return acctest.ConfigCompose(testAccAppConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_studio_lifecycle_config" "test" {
  studio_lifecycle_config_name     = %[1]q
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")
}

resource "aws_sagemaker_user_profile" "lifecycletest" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[2]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_server_app_settings {
      default_resource_spec {
        instance_type        = "system"
        lifecycle_config_arn = aws_sagemaker_studio_lifecycle_config.test.arn
      }

      lifecycle_config_arns = [aws_sagemaker_studio_lifecycle_config.test.arn]
    }
  }
}

resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.lifecycletest.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  resource_spec {
    instance_type        = "system"
    lifecycle_config_arn = aws_sagemaker_studio_lifecycle_config.test.arn
  }
}
`, rName, uName))
}
