// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconvert_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmediaconvert "github.com/hashicorp/terraform-provider-aws/internal/service/mediaconvert"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaConvertJobTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var job_template types.JobTemplate
	resourceName := "aws_media_convert_job_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &job_template),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "mediaconvert", regexache.MustCompile(`job_template/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "settings", string(types.JobTemplateSettings)),
					resource.TestCheckResourceAttr(resourceName, "acceleration_settings", string(types.AccelerationSettings)),
					resource.TestCheckResourceAttr(resourceName, "category", string(types.AccelerationSettings)),
					resource.TestCheckResourceAttr(resourceName, "description", string(types.AccelerationSettings)),
					resource.TestCheckResourceAttr(resourceName, "hop_destinations", string(types.HopDestination)),
					resource.TestCheckResourceAttr(resourceName, "queue", string(types.string)),
					resource.TestCheckResourceAttr(resourceName, "priority", string(types.int32)),
					resource.TestCheckResourceAttr(resourceName, "status_update_interval", string(types.StatusUpdateInterval)),
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
	var job_template types.JobTemplate
	resourceName := "aws_media_convert_job_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &job_template),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmediaconvert.ResourceJobTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMediaConvertJobTemplate_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var job_template types.JobTemplate
	resourceName := "aws_media_convert_job_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &job_template),
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
				Config: testAccJobTemplateConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &job_template),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccJobTemplateConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &job_template),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccMediaConvertJobTemplate_withDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var job_template types.JobTemplate
	resourceName := "aws_media_convert_job_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description1 := sdkacctest.RandomWithPrefix("Description: ")
	description2 := sdkacctest.RandomWithPrefix("Description: ")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_description(rName, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &job_template),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description1),
				),
			},
			{
				Config: testAccJobTemplateConfig_description(rName, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &job_template),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description2),
				),
			},
		},
	})
}

func testAccCheckJobTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_convert_job_template" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).MediaConvertClient(ctx)

			_, err := tfmediaconvert.FindJobTemplateByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckJobTemplateExists(ctx context.Context, n string, v *types.JobTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaConvertClient(ctx)

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
  name = %[1]q
}
`, rName)
}

func testAccJobTemplateConfig_category(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_job_template" "test" {
  name     = %[1]q
  category = %[2]q
}
`, rName, category)
}

func testAccJobTemplateConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_job_template" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccJobTemplateConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_job_template" "test" {
  name = %[1]q

  tags = {
    %[2]s = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccJobTemplateConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_job_template" "test" {
  name = %[1]q

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccJobTemplateConfig_reserved(rName, commitment, renewalType string, reservedSlots int) string {
	return fmt.Sprintf(`
resource "aws_media_convert_job_template" "test" {
  name  = %[1]q
  queue = %[2]q

  settings = %[3]q
}
`, rName, queue, string(types.JobTemplateSettings))
}
