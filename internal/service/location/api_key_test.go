// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLocationAPIKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_api_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "key_arn", "geo", fmt.Sprintf("api-key/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "key_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "key_value"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "no_expiry", "true"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.0", "geo:GetMap*"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_resources.#", "1"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreateTime),
					acctest.CheckResourceAttrRFC3339(resourceName, "update_time"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// key_value is sensitive and may drift on import if not returned by Describe for some SDKs;
				// no_expiry is derived state not returned directly by DescribeKey.
				ImportStateVerifyIgnore: []string{"no_expiry"},
			},
		},
	})
}

func TestAccLocationAPIKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_api_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflocation.ResourceAPIKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLocationAPIKey_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_api_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"no_expiry"},
			},
			{
				Config: testAccAPIKeyConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccLocationAPIKey_restrictions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_api_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_restrictions(rName, "geo:GetMap*", "geo:SearchPlaceIndexForText"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.0", "geo:GetMap*"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.allow_actions.1", "geo:SearchPlaceIndexForText"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"no_expiry"},
			},
		},
	})
}

func TestAccLocationAPIKey_expireTime(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_api_key.test"
	expireTime := "2099-01-01T00:00:00Z"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_expireTime(rName, expireTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "expire_time"),
					resource.TestCheckResourceAttr(resourceName, "no_expiry", "false"),
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

func TestAccLocationAPIKey_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_api_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"no_expiry"},
			},
			{
				Config: testAccAPIKeyConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAPIKeyConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAPIKeyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LocationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_location_api_key" {
				continue
			}

			_, err := conn.DescribeKey(ctx, &location.DescribeKeyInput{
				KeyName: aws.String(rs.Primary.ID),
			})

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Location Service API Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAPIKeyExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).LocationClient(ctx)
		_, err := conn.DescribeKey(ctx, &location.DescribeKeyInput{
			KeyName: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccAPIKeyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_api_key" "test" {
  key_name = %[1]q

  restrictions {
    allow_actions   = ["geo:GetMap*"]
    allow_resources = ["arn:aws:geo:*:*:map/*"]
  }
}
`, rName)
}

func testAccAPIKeyConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_location_api_key" "test" {
  key_name    = %[1]q
  description = %[2]q

  restrictions {
    allow_actions   = ["geo:GetMap*"]
    allow_resources = ["arn:aws:geo:*:*:map/*"]
  }
}
`, rName, description)
}

func testAccAPIKeyConfig_restrictions(rName, action1, action2 string) string {
	return fmt.Sprintf(`
resource "aws_location_api_key" "test" {
  key_name = %[1]q

  restrictions {
    allow_actions   = [%[2]q, %[3]q]
    allow_resources = ["arn:aws:geo:*:*:map/*", "arn:aws:geo:*:*:place-index/*"]
  }
}
`, rName, action1, action2)
}

func testAccAPIKeyConfig_expireTime(rName, expireTime string) string {
	return fmt.Sprintf(`
resource "aws_location_api_key" "test" {
  key_name    = %[1]q
  expire_time = %[2]q

  restrictions {
    allow_actions   = ["geo:GetMap*"]
    allow_resources = ["arn:aws:geo:*:*:map/*"]
  }
}
`, rName, expireTime)
}

func testAccAPIKeyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_location_api_key" "test" {
  key_name = %[1]q

  restrictions {
    allow_actions   = ["geo:GetMap*"]
    allow_resources = ["arn:aws:geo:*:*:map/*"]
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAPIKeyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_location_api_key" "test" {
  key_name = %[1]q

  restrictions {
    allow_actions   = ["geo:GetMap*"]
    allow_resources = ["arn:aws:geo:*:*:map/*"]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
