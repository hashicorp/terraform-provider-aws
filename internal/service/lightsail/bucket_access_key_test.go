// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailBucketAccessKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_bucket_access_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAccessKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccessKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccessKeyExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "access_key_id", regexache.MustCompile(`((?:ASIA|AKIA|AROA|AIDA)([0-7A-Z]{16}))`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestMatchResourceAttr(resourceName, "secret_access_key", regexache.MustCompile(`([0-9A-Za-z+/]{40})`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_access_key", names.AttrBucketName},
			},
		},
	})
}

func TestAccLightsailBucketAccessKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lightsail_bucket_access_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAccessKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccessKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccessKeyExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflightsail.ResourceBucketAccessKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBucketAccessKeyExists(ctx context.Context, t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

		out, err := tflightsail.FindBucketAccessKeyById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("BucketAccessKey %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketAccessKeyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_bucket_access_key" {
				continue
			}

			_, err := tflightsail.FindBucketAccessKeyById(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResBucketAccessKey, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccBucketAccessKeyConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
}
`, rName)
}

func testAccBucketAccessKeyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccBucketAccessKeyConfig_base(rName), `
resource "aws_lightsail_bucket_access_key" "test" {
  bucket_name = aws_lightsail_bucket.test.id
}
`)
}
