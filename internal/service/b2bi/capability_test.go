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

func TestAccB2BICapability_basic(t *testing.T) {
	ctx := acctest2.Context(t)
	resourceName := "aws_b2bi_capability.test"
	rName := acctest.RandomWithPrefix(acctest2.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest2.PreCheck(ctx, t) },
		ErrorCheck:               acctest2.ErrorCheck(t, names.B2BIServiceID),
		ProtoV5ProviderFactories: acctest2.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapabilityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapabilityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "edi"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.edi.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "capability_id"),
					resource.TestCheckResourceAttrSet(resourceName, "capability_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configuration.0.edi.0.type",
				},
			},
		},
	})
}

func testAccCheckCapabilityDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest2.ProviderMeta(ctx, t).B2BIClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_b2bi_capability" {
				continue
			}

			_, err := tfb2bi.FindCapabilityByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("B2BI Capability %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCapabilityExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest2.ProviderMeta(ctx, t).B2BIClient(ctx)

		_, err := tfb2bi.FindCapabilityByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCapabilityConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_b2bi_transformer" "test" {
  name   = %[1]q
  status = "active"

  input_conversion {
    from_format = "X12"

    format_options {
      x12 {
        transaction_set = "X12_110"
        version         = "VERSION_4010"
      }
    }
  }

  mapping {
    template_language = "JSONATA"
    template          = "{}"
  }
}
`, rName)
}

func testAccCapabilityConfig_basic(rName string) string {
	return acctest2.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
resource "aws_b2bi_capability" "test" {
  name = %[1]q
  type = "edi"

  configuration {
    edi {
      input_location {
        bucket_name = aws_s3_bucket.test.bucket
        key         = "input/"
      }

      output_location {
        bucket_name = aws_s3_bucket.test.bucket
        key         = "output/"
      }

      transformer_id = aws_b2bi_transformer.test.transformer_id

      type {
        x12_details {
          transaction_set = "X12_110"
          version         = "VERSION_4010"
        }
      }
    }
  }
}
`, rName))
}

func TestAccB2BICapability_tags(t *testing.T) {
	ctx := acctest2.Context(t)
	resourceName := "aws_b2bi_capability.test"
	rName := acctest.RandomWithPrefix(acctest2.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest2.PreCheck(ctx, t) },
		ErrorCheck:               acctest2.ErrorCheck(t, names.B2BIServiceID),
		ProtoV5ProviderFactories: acctest2.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapabilityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapabilityConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccCapabilityConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCapabilityConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCapabilityConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest2.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
resource "aws_b2bi_capability" "test" {
  name = %[1]q
  type = "edi"

  configuration {
    edi {
      input_location {
        bucket_name = aws_s3_bucket.test.bucket
        key         = "input/"
      }

      output_location {
        bucket_name = aws_s3_bucket.test.bucket
        key         = "output/"
      }

      transformer_id = aws_b2bi_transformer.test.transformer_id

      type {
        x12_details {
          transaction_set = "X12_110"
          version         = "VERSION_4010"
        }
      }
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccCapabilityConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest2.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
resource "aws_b2bi_capability" "test" {
  name = %[1]q
  type = "edi"

  configuration {
    edi {
      input_location {
        bucket_name = aws_s3_bucket.test.bucket
        key         = "input/"
      }

      output_location {
        bucket_name = aws_s3_bucket.test.bucket
        key         = "output/"
      }

      transformer_id = aws_b2bi_transformer.test.transformer_id

      type {
        x12_details {
          transaction_set = "X12_110"
          version         = "VERSION_4010"
        }
      }
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
