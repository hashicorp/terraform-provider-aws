// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftimestreamwrite "github.com/hashicorp/terraform-provider-aws/internal/service/timestreamwrite"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamWriteDatabase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "timestream", fmt.Sprintf("database/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tftimestreamwrite.ResourceDatabase(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTimestreamWriteDatabase_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	kmsResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, t, resourceName),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	kmsResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
				),
			},
			{
				Config: testAccDatabaseConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, t, resourceName),
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
					testAccCheckDatabaseExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
				),
			},
		},
	})
}

func TestAccTimestreamWriteDatabase_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				Config: testAccDatabaseConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
				),
			},
			{
				Config: testAccDatabaseConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
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

func testAccCheckDatabaseDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).TimestreamWriteClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_timestreamwrite_database" {
				continue
			}

			_, err := tftimestreamwrite.FindDatabaseByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckDatabaseExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).TimestreamWriteClient(ctx)

		_, err := tftimestreamwrite.FindDatabaseByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).TimestreamWriteClient(ctx)

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
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true

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
