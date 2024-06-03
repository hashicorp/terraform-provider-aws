// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftimestreamwrite "github.com/hashicorp/terraform-provider-aws/internal/service/timestreamwrite"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamWriteDatabase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_timestreamwrite_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "timestream", fmt.Sprintf("database/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
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

func TestAccTimestreamWriteDatabase_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_timestreamwrite_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftimestreamwrite.ResourceDatabase(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTimestreamWriteDatabase_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_timestreamwrite_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsResourceName, names.AttrARN),
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

func TestAccTimestreamWriteDatabase_updateKMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_timestreamwrite_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
				),
			},
			{
				Config: testAccDatabaseConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
				),
			},
		},
	})
}

func TestAccTimestreamWriteDatabase_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_timestreamwrite_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				Config: testAccDatabaseConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
				),
			},
			{
				Config: testAccDatabaseConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
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

func testAccCheckDatabaseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamWriteClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_timestreamwrite_database" {
				continue
			}

			_, err := tftimestreamwrite.FindDatabaseByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Timestream Database %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDatabaseExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamWriteClient(ctx)

		_, err := tftimestreamwrite.FindDatabaseByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamWriteClient(ctx)

	input := &timestreamwrite.ListDatabasesInput{}

	_, err := conn.ListDatabases(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDatabaseConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}
`, rName)
}

func testAccDatabaseConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDatabaseConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDatabaseConfig_kmsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  policy = <<POLICY
{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Effect": "Allow",
     "Principal": {
       "AWS": "*"
     },
     "Action": "kms:*",
     "Resource": "*"
   }
 ]
}
POLICY
}

resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
  kms_key_id    = aws_kms_key.test.arn
}
`, rName)
}
