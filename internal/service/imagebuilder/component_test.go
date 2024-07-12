// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfimagebuilder "github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccImageBuilderComponent_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComponentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "imagebuilder", regexache.MustCompile(fmt.Sprintf("component/%s/1.0.0/[1-9][0-9]*", rName))),
					resource.TestCheckResourceAttr(resourceName, "change_description", ""),
					resource.TestMatchResourceAttr(resourceName, "data", regexache.MustCompile(`schemaVersion`)),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwner),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(resourceName, "supported_os_versions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, imagebuilder.ComponentTypeBuild),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1.0.0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccImageBuilderComponent_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComponentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfimagebuilder.ResourceComponent(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderComponent_changeDescription(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComponentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentConfig_changeDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "change_description", "description1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccImageBuilderComponent_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComponentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccImageBuilderComponent_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComponentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccImageBuilderComponent_Platform_windows(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComponentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentConfig_platformWindows(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformWindows),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccImageBuilderComponent_supportedOsVersions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComponentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentConfig_supportedOsVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "supported_os_versions.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccImageBuilderComponent_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComponentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
			{
				Config: testAccComponentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccComponentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccImageBuilderComponent_uri(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComponentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentConfig_uri(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComponentExists(ctx, resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy, names.AttrURI},
			},
		},
	})
}

func testAccCheckComponentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_imagebuilder_component" {
				continue
			}

			input := &imagebuilder.GetComponentInput{
				ComponentBuildVersionArn: aws.String(rs.Primary.ID),
			}

			output, err := conn.GetComponentWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Image Builder Component (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Image Builder Component (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckComponentExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		input := &imagebuilder.GetComponentInput{
			ComponentBuildVersionArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetComponentWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Component (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccComponentConfig_changeDescription(rName string, changeDescription string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  change_description = %[2]q
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"
}
`, rName, changeDescription)
}

func testAccComponentConfig_description(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  description = %[2]q
  name        = %[1]q
  platform    = "Linux"
  version     = "1.0.0"
}
`, rName, description)
}

func testAccComponentConfig_kmsKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  kms_key_id = aws_kms_key.test.arn
  name       = %[1]q
  platform   = "Linux"
  version    = "1.0.0"
}
`, rName)
}

func testAccComponentConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"
}
`, rName)
}

func testAccComponentConfig_platformWindows(rName string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecutePowerShell"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Windows"
  version  = "1.0.0"
}
`, rName)
}

func testAccComponentConfig_supportedOsVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name                  = %[1]q
  platform              = "Linux"
  supported_os_versions = ["Amazon Linux 2"]
  version               = "1.0.0"
}
`, rName)
}

func testAccComponentConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccComponentConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccComponentConfig_uri(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  content = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  key = "test.yml"
}

resource "aws_imagebuilder_component" "test" {
  name     = %[1]q
  platform = "Linux"
  uri      = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  version  = "1.0.0"
}
`, rName)
}
