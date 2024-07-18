// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBTableReplica_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
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

func TestAccDynamoDBTableReplica_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdynamodb.ResourceTableReplica(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDynamoDBTableReplica_pitr(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_pitr(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery", acctest.CtTrue),
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

func TestAccDynamoDBTableReplica_pitrKMS(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_pitrKMS(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				Config: testAccTableReplicaConfig_pitrKMS(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				Config: testAccTableReplicaConfig_pitrKMS(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
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

func TestAccDynamoDBTableReplica_pitrDefault(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_pitrDefault(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyARN),
				),
			},
			{
				Config: testAccTableReplicaConfig_pitrDefault(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyARN),
				),
			},
			{
				Config: testAccTableReplicaConfig_pitrDefault(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyARN),
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

func TestAccDynamoDBTableReplica_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_tags1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.tape", "Valladolid"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_tags2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "5"),
					resource.TestCheckResourceAttr(resourceName, "tags.arise", "Melandru"),
					resource.TestCheckResourceAttr(resourceName, "tags.brightest", "Lights"),
					resource.TestCheckResourceAttr(resourceName, "tags.shooting", "Stars"),
					resource.TestCheckResourceAttr(resourceName, "tags.tape", "Valladolid"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_tags3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccDynamoDBTableReplica_tableClass(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_tableClass(rName, "STANDARD"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "table_class_override", "STANDARD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_tableClass(rName, "STANDARD_INFREQUENT_ACCESS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "table_class_override", "STANDARD_INFREQUENT_ACCESS"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_keys(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_keys(rName, "test1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_keys(rName, "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_autoscaleSource(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_autoscaleSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "global_table_arn", "aws_dynamodb_table.source", "arn"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_autoscaleSourceWithoutDependsOn(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTableReplicaConfig_autoscaleSourceWithoutDependsOn(rName),
				ExpectError: regexache.MustCompile("For a DynamoDB Table Replica, the source table must either have a"),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_billingModeProvisionedSource(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTableReplicaConfig_billingModeProvisionedSource(rName),
				ExpectError: regexache.MustCompile("For a DynamoDB Table Replica, the source table must either have a"),
			},
		},
	})
}

func testAccCheckTableReplicaDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)
		replicaRegion := acctest.Region()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dynamodb_table_replica" {
				continue
			}

			tableName, mainRegion, err := tfdynamodb.TableReplicaParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			optFn := func(o *dynamodb.Options) {
				o.Region = mainRegion
			}
			output, err := tfdynamodb.FindTableByName(ctx, conn, tableName, optFn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if tfdynamodb.ReplicaForRegion(output.Replicas, replicaRegion) == nil {
				continue
			}

			return fmt.Errorf("DynamoDB Table Replica %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTableReplicaExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		tableName, mainRegion, err := tfdynamodb.TableReplicaParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		optFn := func(o *dynamodb.Options) {
			o.Region = mainRegion
		}
		_, err = tfdynamodb.FindTableByName(ctx, conn, tableName, optFn)

		return err
	}
}

func testAccTableReplicaConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  global_table_arn = aws_dynamodb_table.source.arn

  tags = {
    Name = %[1]q
    Pozo = "Amargo"
  }
}

resource "aws_dynamodb_table" "source" {
  provider         = "awsalternate"
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [replica]
  }
}
`, rName))
}

func testAccTableReplicaConfig_pitr(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  provider               = "awsalternate"
  global_table_arn       = aws_dynamodb_table.source.arn
  point_in_time_recovery = true
}

resource "aws_dynamodb_table" "source" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [replica]
  }
}
`, rName))
}

func testAccTableReplicaConfig_pitrKMS(rName string, pitr bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  provider               = awsalternate
  global_table_arn       = aws_dynamodb_table.source.arn
  point_in_time_recovery = %[2]t
  kms_key_arn            = aws_kms_key.test.arn
}

resource "aws_kms_key" "test" {
  provider                = awsalternate
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_dynamodb_table" "source" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.source.arn
  }

  lifecycle {
    ignore_changes = [replica]
  }
}

resource "aws_kms_key" "source" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName, pitr))
}

func testAccTableReplicaConfig_pitrDefault(rName string, pitr bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  provider               = awsalternate
  global_table_arn       = aws_dynamodb_table.source.arn
  point_in_time_recovery = %[2]t
}

resource "aws_dynamodb_table" "source" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }

  lifecycle {
    ignore_changes = [replica]
  }
}
`, rName, pitr))
}

func testAccTableReplicaConfig_tags1(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  provider         = "awsalternate"
  global_table_arn = aws_dynamodb_table.source.arn

  tags = {
    Name = %[1]q
    tape = "Valladolid"
  }
}

resource "aws_dynamodb_table" "source" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  tags = {
    Name = %[1]q
  }

  lifecycle {
    ignore_changes = [replica]
  }
}

`, rName))
}

func testAccTableReplicaConfig_tags2(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  provider         = "awsalternate"
  global_table_arn = aws_dynamodb_table.source.arn

  tags = {
    Name      = %[1]q
    tape      = "Valladolid"
    brightest = "Lights"
    arise     = "Melandru"
    shooting  = "Stars"
  }
}

resource "aws_dynamodb_table" "source" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  tags = {
    Name = %[1]q
  }

  lifecycle {
    ignore_changes = [replica]
  }
}
`, rName))
}

func testAccTableReplicaConfig_tags3(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  provider         = "awsalternate"
  global_table_arn = aws_dynamodb_table.source.arn

  tags = {
    Name = %[1]q
  }
}

resource "aws_dynamodb_table" "source" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  tags = {
    Name = %[1]q
  }

  lifecycle {
    ignore_changes = [replica]
  }
}
`, rName))
}

func testAccTableReplicaConfig_tableClass(rName, class string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  global_table_arn     = aws_dynamodb_table.source.arn
  table_class_override = %[2]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_dynamodb_table" "source" {
  provider         = "awsalternate"
  name             = %[1]q
  hash_key         = "ArticLake"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "ArticLake"
    type = "S"
  }

  tags = {
    Name = %[1]q
  }

  lifecycle {
    ignore_changes = [replica]
  }
}
`, rName, class))
}

func testAccTableReplicaConfig_keys(rName, key string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  global_table_arn       = aws_dynamodb_table.source.arn
  kms_key_arn            = aws_kms_key.%[2]s.arn
  point_in_time_recovery = true
}

resource "aws_kms_key" "test1" {
  description             = "Julie test KMS key Z"
  multi_region            = false
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  description             = "Julie test KMS key Z"
  multi_region            = false
  deletion_window_in_days = 7
}

resource "aws_dynamodb_table" "source" {
  provider         = awsalternate
  name             = %[1]q
  hash_key         = "ParticipantId"
  range_key        = "SubscriptionId"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  table_class      = "STANDARD"

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.source.arn
  }

  attribute {
    name = "ParticipantId"
    type = "S"
  }

  attribute {
    name = "SubscriptionId"
    type = "S"
  }

  lifecycle {
    ignore_changes = [replica]
  }
}

resource "aws_kms_key" "source" {
  provider                = awsalternate
  description             = "Julie test KMS key A"
  multi_region            = false
  deletion_window_in_days = 7
}
`, rName, key))
}

func testAccTableReplicaConfig_autoscaleSource(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  global_table_arn = aws_dynamodb_table.source.arn

  depends_on = [
    aws_appautoscaling_policy.source_read,
    aws_appautoscaling_policy.source_write
  ]
}

resource "aws_dynamodb_table" "source" {
  provider       = "awsalternate"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [
      replica,
      read_capacity,
      write_capacity,
    ]
  }
}

resource "aws_appautoscaling_target" "source_read" {
  provider           = "awsalternate"
  max_capacity       = 735
  min_capacity       = 28
  resource_id        = "table/${aws_dynamodb_table.source.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "source_read" {
  provider           = "awsalternate"
  name               = "%[1]s-read"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.source_read.resource_id
  scalable_dimension = aws_appautoscaling_target.source_read.scalable_dimension
  service_namespace  = aws_appautoscaling_target.source_read.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBReadCapacityUtilization"
    }
    target_value = 50
  }
}

resource "aws_appautoscaling_target" "source_write" {
  provider     = "awsalternate"
  max_capacity = 900
  min_capacity = 28

  resource_id        = "table/${aws_dynamodb_table.source.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "source_write" {
  provider           = "awsalternate"
  name               = "%[1]s-write"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.source_write.resource_id
  scalable_dimension = aws_appautoscaling_target.source_write.scalable_dimension
  service_namespace  = aws_appautoscaling_target.source_write.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBWriteCapacityUtilization"
    }
    target_value = 50
  }
}
`, rName))
}

func testAccTableReplicaConfig_autoscaleSourceWithoutDependsOn(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  global_table_arn = aws_dynamodb_table.source.arn

  # Explicitly does not include depends_on
}

resource "aws_dynamodb_table" "source" {
  provider       = "awsalternate"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [
      replica,
      read_capacity,
      write_capacity,
    ]
  }
}

resource "aws_appautoscaling_target" "source_read" {
  provider           = "awsalternate"
  max_capacity       = 735
  min_capacity       = 28
  resource_id        = "table/${aws_dynamodb_table.source.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "source_read" {
  provider           = "awsalternate"
  name               = "%[1]s-read"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.source_read.resource_id
  scalable_dimension = aws_appautoscaling_target.source_read.scalable_dimension
  service_namespace  = aws_appautoscaling_target.source_read.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBReadCapacityUtilization"
    }
    target_value = 50
  }
}

resource "aws_appautoscaling_target" "source_write" {
  provider     = "awsalternate"
  max_capacity = 900
  min_capacity = 28

  resource_id        = "table/${aws_dynamodb_table.source.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "source_write" {
  provider           = "awsalternate"
  name               = "%[1]s-write"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.source_write.resource_id
  scalable_dimension = aws_appautoscaling_target.source_write.scalable_dimension
  service_namespace  = aws_appautoscaling_target.source_write.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBWriteCapacityUtilization"
    }
    target_value = 50
  }
}
`, rName))
}

func testAccTableReplicaConfig_billingModeProvisionedSource(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table_replica" "test" {
  global_table_arn = aws_dynamodb_table.source.arn
}

resource "aws_dynamodb_table" "source" {
  provider       = "awsalternate"
  name           = %[1]q
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [replica]
  }
}
`, rName))
}
