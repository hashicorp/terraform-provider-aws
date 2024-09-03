// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qldb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "qldb", regexache.MustCompile(`ledger/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKey, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "permissions_mode", "ALLOW_ALL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_nameGenerated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, names.AttrName, regexache.MustCompile(`tf\d+`)),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "permissions_mode", "STANDARD"),
				),
			},
			{
				Config: testAccLedgerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_kmsKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKey, kmsKeyResourceName, names.AttrARN),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKey, "AWS_OWNED_KMS_KEY"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLedgerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLedgerConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLedgerConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLedgerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
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
