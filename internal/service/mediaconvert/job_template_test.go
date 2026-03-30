// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mediaconvert_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmediaconvert "github.com/hashicorp/terraform-provider-aws/internal/service/mediaconvert"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaConvertJobTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var jobTemplate types.JobTemplate
	resourceName := "aws_media_convert_job_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, t, resourceName, &jobTemplate),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediaconvert", regexache.MustCompile(`jobTemplates/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "0"),
					resource.TestCheckResourceAttr(resourceName, "acceleration_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "hop_destinations.#", "0"),
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

func TestAccMediaConvertJobTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var jobTemplate types.JobTemplate
	resourceName := "aws_media_convert_job_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, t, resourceName, &jobTemplate),
					acctest.CheckSDKResourceDisappears(ctx, t, tfmediaconvert.ResourceJobTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMediaConvertJobTemplate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var jobTemplate types.JobTemplate
	resourceName := "aws_media_convert_job_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, t, resourceName, &jobTemplate),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobTemplateConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, t, resourceName, &jobTemplate),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccJobTemplateConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, t, resourceName, &jobTemplate),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccMediaConvertJobTemplate_description(t *testing.T) {
	ctx := acctest.Context(t)
	var jobTemplate types.JobTemplate
	resourceName := "aws_media_convert_job_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description1 := acctest.RandomWithPrefix(t, "Description: ")
	description2 := acctest.RandomWithPrefix(t, "Description: ")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_description(rName, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, t, resourceName, &jobTemplate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description1),
				),
			},
			{
				Config: testAccJobTemplateConfig_description(rName, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, t, resourceName, &jobTemplate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description2),
				),
			},
		},
	})
}

func TestAccMediaConvertJobTemplate_accelerationSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var jobTemplate types.JobTemplate
	resourceName := "aws_media_convert_job_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_accelerationSettings(rName, string(types.AccelerationModeDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, t, resourceName, &jobTemplate),
					resource.TestCheckResourceAttr(resourceName, "acceleration_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "acceleration_settings.0.mode", string(types.AccelerationModeDisabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobTemplateConfig_accelerationSettings(rName, string(types.AccelerationModeEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, t, resourceName, &jobTemplate),
					resource.TestCheckResourceAttr(resourceName, "acceleration_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "acceleration_settings.0.mode", string(types.AccelerationModeEnabled)),
				),
			},
		},
	})
}

func testAccCheckJobTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_convert_job_template" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).MediaConvertClient(ctx)

			_, err := tfmediaconvert.FindJobTemplateByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Media Convert Job Template %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckJobTemplateExists(ctx context.Context, t *testing.T, n string, v *types.JobTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).MediaConvertClient(ctx)

		output, err := tfmediaconvert.FindJobTemplateByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccJobTemplateConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_job_template" "test" {
  name          = %[1]q
  settings_json = "{}"
}
`, rName)
}

func testAccJobTemplateConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_job_template" "test" {
  name          = %[1]q
  settings_json = "{}"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccJobTemplateConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_job_template" "test" {
  name          = %[1]q
  settings_json = "{}"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccJobTemplateConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_job_template" "test" {
  name          = %[1]q
  description   = %[2]q
  settings_json = "{}"
}
`, rName, description)
}

func testAccJobTemplateConfig_accelerationSettings(rName, mode string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_job_template" "test" {
  name          = %[1]q
  settings_json = "{}"

  acceleration_settings {
    mode = %[2]q
  }
}
`, rName, mode)
}
