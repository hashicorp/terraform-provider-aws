// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package paymentcryptography_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/paymentcryptography/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkeyalias "github.com/hashicorp/terraform-provider-aws/internal/service/paymentcryptography"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPaymentCryptographyKeyAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("alias/")
	resourceName := "aws_paymentcryptography_key_alias.test"
	var v awstypes.Alias

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.PaymentCryptographyEndpointID) // Causing a false negative
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PaymentCryptographyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyAliasConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "alias_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "key_arn"),
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

func TestAccPaymentCryptographyKeyAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("alias/")
	resourceName := "aws_paymentcryptography_key_alias.test"
	var v awstypes.Alias

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.PaymentCryptographyEndpointID) // Causing a false negative
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PaymentCryptographyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyAliasExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfkeyalias.ResourceKeyAlias, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPaymentCryptographyKeyAlias_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("alias/")
	resourceName := "aws_paymentcryptography_key_alias.test"
	var v awstypes.Alias

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.PaymentCryptographyEndpointID) // Causing a false negative
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PaymentCryptographyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyAliasConfig_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "alias_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "key_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyAliasConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "alias_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "key_arn"),
				),
			},
		},
	})
}

func TestAccPaymentCryptographyKeyAlias_updateName(t *testing.T) {
	ctx := acctest.Context(t)
	rOldName := sdkacctest.RandomWithPrefix("alias/")
	rNewName := sdkacctest.RandomWithPrefix("alias/")
	resourceName := "aws_paymentcryptography_key_alias.test"
	var v awstypes.Alias

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.PaymentCryptographyEndpointID) // Causing a false negative
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PaymentCryptographyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyAliasConfig_updateName(rOldName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "alias_name", rOldName),
					resource.TestCheckResourceAttrSet(resourceName, "key_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyAliasConfig_updateName(rNewName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "alias_name", rNewName),
					resource.TestCheckResourceAttrSet(resourceName, "key_arn"),
				),
			},
		},
	})
}

func testAccCheckKeyAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PaymentCryptographyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_paymentcryptography_key_alias" {
				continue
			}

			_, err := tfkeyalias.FindKeyAliasByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Key Alias %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckKeyAliasExists(ctx context.Context, n string, v *awstypes.Alias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PaymentCryptographyClient(ctx)

		output, err := tfkeyalias.FindKeyAliasByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccKeyAliasConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_paymentcryptography_key" "test" {
  exportable = true
  key_attributes {
    key_algorithm = "TDES_3KEY"
    key_class     = "SYMMETRIC_KEY"
    key_usage     = "TR31_P0_PIN_ENCRYPTION_KEY"
    key_modes_of_use {
      decrypt = true
      encrypt = true
      wrap    = true
      unwrap  = true
    }
  }
}
resource "aws_paymentcryptography_key_alias" "test" {
  alias_name = %[1]q
  key_arn    = aws_paymentcryptography_key.test.arn
}
`, name)
}

func testAccKeyAliasConfig_update(name string) string {
	return fmt.Sprintf(`
resource "aws_paymentcryptography_key_alias" "test" {
  alias_name = %[1]q
  key_arn    = aws_paymentcryptography_key.test.arn
}

resource "aws_paymentcryptography_key" "test" {
  exportable = true
  key_attributes {
    key_algorithm = "TDES_3KEY"
    key_class     = "SYMMETRIC_KEY"
    key_usage     = "TR31_P0_PIN_ENCRYPTION_KEY"
    key_modes_of_use {
      decrypt = true
      encrypt = true
      wrap    = true
      unwrap  = true
    }
  }
}
`, name)
}

func testAccKeyAliasConfig_updateName(name string) string {
	return fmt.Sprintf(`
resource "aws_paymentcryptography_key_alias" "test" {
  alias_name = %[1]q
  key_arn    = aws_paymentcryptography_key.test.arn
}

resource "aws_paymentcryptography_key" "test" {
  exportable = true
  key_attributes {
    key_algorithm = "TDES_3KEY"
    key_class     = "SYMMETRIC_KEY"
    key_usage     = "TR31_P0_PIN_ENCRYPTION_KEY"
    key_modes_of_use {
      decrypt = true
      encrypt = true
      wrap    = true
      unwrap  = true
    }
  }
}
`, name)
}
