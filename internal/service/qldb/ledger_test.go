// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qldb_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/qldb"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfqldb "github.com/hashicorp/terraform-provider-aws/internal/service/qldb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQLDBLedger_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v qldb.DescribeLedgerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "qldb", regexp.MustCompile(`ledger/.+`)),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "permissions_mode", "ALLOW_ALL"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccQLDBLedger_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v qldb.DescribeLedgerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfqldb.ResourceLedger(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQLDBLedger_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v qldb.DescribeLedgerOutput
	resourceName := "aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_nameGenerated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(`tf\d+`)),
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

func TestAccQLDBLedger_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v qldb.DescribeLedgerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions_mode", "ALLOW_ALL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLedgerConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions_mode", "STANDARD"),
				),
			},
			{
				Config: testAccLedgerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions_mode", "ALLOW_ALL"),
				),
			},
		},
	})
}

func TestAccQLDBLedger_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v qldb.DescribeLedgerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_ledger.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_kmsKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLedgerConfig_kmsKeyUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "kms_key", "AWS_OWNED_KMS_KEY"),
				),
			},
		},
	})
}

func TestAccQLDBLedger_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v qldb.DescribeLedgerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLedgerConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLedgerConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckLedgerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QLDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_qldb_ledger" {
				continue
			}

			_, err := tfqldb.FindLedgerByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QLDB Ledger %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLedgerExists(ctx context.Context, n string, v *qldb.DescribeLedgerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No QLDB Ledger ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QLDBClient(ctx)

		output, err := tfqldb.FindLedgerByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLedgerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  permissions_mode    = "ALLOW_ALL"
  deletion_protection = false
}
`, rName)
}

func testAccLedgerConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  permissions_mode    = "STANDARD"
  deletion_protection = true
}
`, rName)
}

func testAccLedgerConfig_nameGenerated() string {
	return `
resource "aws_qldb_ledger" "test" {
  permissions_mode    = "ALLOW_ALL"
  deletion_protection = false
}
`
}

func testAccLedgerConfig_kmsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  permissions_mode    = "ALLOW_ALL"
  deletion_protection = false
  kms_key             = aws_kms_key.test.arn
}
`, rName)
}

func testAccLedgerConfig_kmsKeyUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  permissions_mode    = "ALLOW_ALL"
  deletion_protection = false
  kms_key             = "AWS_OWNED_KMS_KEY"
}
`, rName)
}

func testAccLedgerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  permissions_mode    = "ALLOW_ALL"
  deletion_protection = false

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccLedgerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  permissions_mode    = "ALLOW_ALL"
  deletion_protection = false

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
