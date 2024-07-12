// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSPublicKeyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kms_key.test"
	datasourceName := "data.aws_kms_public_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccPublicKeyCheckDataSource(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "customer_master_key_spec", resourceName, "customer_master_key_spec"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrKeyID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "key_usage", resourceName, "key_usage"),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrPublicKey),
					resource.TestCheckResourceAttrSet(datasourceName, "public_key_pem"),
				),
			},
		},
	})
}

func TestAccKMSPublicKeyDataSource_encrypt(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kms_key.test"
	datasourceName := "data.aws_kms_public_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyDataSourceConfig_encrypt(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccPublicKeyCheckDataSource(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrKeyID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "customer_master_key_spec", resourceName, "customer_master_key_spec"),
					resource.TestCheckResourceAttrPair(datasourceName, "key_usage", resourceName, "key_usage"),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrPublicKey),
					resource.TestCheckResourceAttrSet(datasourceName, "public_key_pem"),
				),
			},
		},
	})
}

func testAccPublicKeyCheckDataSource(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

func testAccPublicKeyDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description              = %[1]q
  deletion_window_in_days  = 7
  customer_master_key_spec = "RSA_2048"
  key_usage                = "SIGN_VERIFY"
}

data "aws_kms_public_key" "test" {
  key_id = aws_kms_key.test.arn
}
`, rName)
}

func testAccPublicKeyDataSourceConfig_encrypt(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description              = %[1]q
  deletion_window_in_days  = 7
  customer_master_key_spec = "RSA_2048"
  key_usage                = "ENCRYPT_DECRYPT"
}

data "aws_kms_public_key" "test" {
  key_id = aws_kms_key.test.arn
}
`, rName)
}
