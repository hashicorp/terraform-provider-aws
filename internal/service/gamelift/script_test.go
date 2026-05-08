// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGameLiftScript_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Script
	resourceName := "aws_gamelift_script.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	region := acctest.Region()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScriptDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScriptConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`script/script-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", fmt.Sprintf("prod-gamescale-scripts-%s", region)),
					resource.TestCheckResourceAttrSet(resourceName, "storage_location.0.key"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file"},
			},
			{
				Config: testAccScriptConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`script/script-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccGameLiftScript_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Script

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_gamelift_script.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScriptDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScriptConfig_basicTags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file"},
			},
			{
				Config: testAccScriptConfig_basicTags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccScriptConfig_basicTags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGameLiftScript_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Script

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_gamelift_script.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScriptDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScriptConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(ctx, t, resourceName, &conf),
					acctest.CheckSDKResourceDisappears(ctx, t, tfgamelift.ResourceScript(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfgamelift.ResourceScript(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScriptExists(ctx context.Context, t *testing.T, n string, v *awstypes.Script) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GameLiftClient(ctx)

		output, err := tfgamelift.FindScriptByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckScriptDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GameLiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_gamelift_script" {
				continue
			}

			script, err := tfgamelift.FindScriptByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if script != nil {
				return fmt.Errorf("GameLift Script (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccScriptConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/script.zip"
}
`, rName)
}

func testAccScriptConfig_basicTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/script.zip"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccScriptConfig_basicTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/script.zip"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
