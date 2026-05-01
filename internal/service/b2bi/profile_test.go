// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package b2bi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	tfb2bi "github.com/hashicorp/terraform-provider-aws/internal/service/b2bi"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	acctest2 "github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccB2BIProfile_basic(t *testing.T) {
	ctx := acctest2.Context(t)
	resourceName := "aws_b2bi_profile.test"
	rName := acctest.RandomWithPrefix(acctest2.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest2.PreCheck(ctx, t) },
		ErrorCheck:               acctest2.ErrorCheck(t, names.B2BIServiceID),
		ProtoV5ProviderFactories: acctest2.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "business_name", "Test Business"),
					resource.TestCheckResourceAttr(resourceName, "phone", "5555555555"),
					resource.TestCheckResourceAttr(resourceName, "logging", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "profile_id"),
					resource.TestCheckResourceAttrSet(resourceName, "profile_arn"),
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

func TestAccB2BIProfile_update(t *testing.T) {
	ctx := acctest2.Context(t)
	resourceName := "aws_b2bi_profile.test"
	rName := acctest.RandomWithPrefix(acctest2.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(acctest2.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest2.PreCheck(ctx, t) },
		ErrorCheck:               acctest2.ErrorCheck(t, names.B2BIServiceID),
		ProtoV5ProviderFactories: acctest2.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccProfileConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
				),
			},
		},
	})
}

func testAccCheckProfileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest2.ProviderMeta(ctx, t).B2BIClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_b2bi_profile" {
				continue
			}

			_, err := tfb2bi.FindProfileByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("B2BI Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProfileExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest2.ProviderMeta(ctx, t).B2BIClient(ctx)

		_, err := tfb2bi.FindProfileByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccProfileConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_b2bi_profile" "test" {
  name          = %[1]q
  business_name = "Test Business"
  phone         = "5555555555"
  logging       = "ENABLED"
}
`, rName)
}

func TestAccB2BIProfile_tags(t *testing.T) {
	ctx := acctest2.Context(t)
	resourceName := "aws_b2bi_profile.test"
	rName := acctest.RandomWithPrefix(acctest2.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest2.PreCheck(ctx, t) },
		ErrorCheck:               acctest2.ErrorCheck(t, names.B2BIServiceID),
		ProtoV5ProviderFactories: acctest2.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccProfileConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccProfileConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccB2BIProfile_email(t *testing.T) {
	ctx := acctest2.Context(t)
	resourceName := "aws_b2bi_profile.test"
	rName := acctest.RandomWithPrefix(acctest2.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest2.PreCheck(ctx, t) },
		ErrorCheck:               acctest2.ErrorCheck(t, names.B2BIServiceID),
		ProtoV5ProviderFactories: acctest2.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_email(rName, "test@example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", "test@example.com"),
				),
			},
			{
				Config: testAccProfileConfig_email(rName, "updated@example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", "updated@example.com"),
				),
			},
		},
	})
}

func testAccProfileConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_b2bi_profile" "test" {
  name          = %[1]q
  business_name = "Test Business"
  phone         = "5555555555"
  logging       = "ENABLED"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccProfileConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_b2bi_profile" "test" {
  name          = %[1]q
  business_name = "Test Business"
  phone         = "5555555555"
  logging       = "ENABLED"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccProfileConfig_email(rName, email string) string {
	return fmt.Sprintf(`
resource "aws_b2bi_profile" "test" {
  name          = %[1]q
  business_name = "Test Business"
  phone         = "5555555555"
  email         = %[2]q
  logging       = "ENABLED"
}
`, rName, email)
}
