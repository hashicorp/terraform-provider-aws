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

func TestAccB2BIPartnership_basic(t *testing.T) {
	ctx := acctest2.Context(t)
	resourceName := "aws_b2bi_partnership.test"
	rName := acctest.RandomWithPrefix(acctest2.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest2.PreCheck(ctx, t) },
		ErrorCheck:               acctest2.ErrorCheck(t, names.B2BIServiceID),
		ProtoV5ProviderFactories: acctest2.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartnershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPartnershipConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartnershipExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "email", "test@example.com"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "partnership_id"),
					resource.TestCheckResourceAttrSet(resourceName, "partnership_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "trading_partner_id"),
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

func testAccCheckPartnershipDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest2.ProviderMeta(ctx, t).B2BIClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_b2bi_partnership" {
				continue
			}

			_, err := tfb2bi.FindPartnershipByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("B2BI Partnership %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPartnershipExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest2.ProviderMeta(ctx, t).B2BIClient(ctx)

		_, err := tfb2bi.FindPartnershipByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPartnershipConfig_basic(rName string) string {
	return acctest2.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
resource "aws_b2bi_profile" "test" {
  name          = %[1]q
  business_name = "Test Business"
  phone         = "5555555555"
  logging       = "ENABLED"
}

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

resource "aws_b2bi_partnership" "test" {
  name         = %[1]q
  email        = "test@example.com"
  profile_id   = aws_b2bi_profile.test.profile_id
  capabilities = [aws_b2bi_capability.test.capability_id]
}
`, rName))
}
