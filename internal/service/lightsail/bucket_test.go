// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailBucket_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "lightsail", regexache.MustCompile(`Bucket/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttr(resourceName, "bundle_id", "small_1_0"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRegion),
					resource.TestCheckResourceAttrSet(resourceName, "support_code"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrForceDelete),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDelete, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDelete,
				},
			},
		},
	})
}

func TestAccLightsailBucket_BundleId(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_bucket.test"
	bundle1 := "small_1_0"
	bundle2 := "medium_1_0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_bundleId(rName, bundle1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bundle_id", bundle1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDelete,
				},
			},
			{
				Config: testAccBucketConfig_bundleId(rName, bundle2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bundle_id", bundle2),
				),
			},
		},
	})
}

func TestAccLightsailBucket_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflightsail.ResourceBucket(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLightsailBucket_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDelete,
				},
			},
			{
				Config: testAccBucketConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBucketConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccLightsailBucket_keyOnlyTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_tags1(rName, acctest.CtKey1, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDelete,
				},
			},
			{
				Config: testAccBucketConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
			{
				Config: testAccBucketConfig_tags1(rName, acctest.CtKey2, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
		},
	})
}

func testAccCheckBucketExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		out, err := tflightsail.FindBucketById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("Bucket %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_bucket" {
				continue
			}

			_, err := tflightsail.FindBucketById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResBucket, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func TestAccLightsailBucket_forceDelete(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDelete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDelete, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDelete,
				},
			},
		},
	})
}

func testAccBucketConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
}
`, rName)
}

func testAccBucketConfig_bundleId(rName, rBundleId string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = %[2]q
}
`, rName, rBundleId)
}

func testAccBucketConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccBucketConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccBucketConfig_forceDelete(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_bucket" "test" {
  name         = %[1]q
  bundle_id    = "small_1_0"
  force_delete = true
}
`, rName)
}
