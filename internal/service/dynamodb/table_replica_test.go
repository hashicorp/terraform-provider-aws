package dynamodb_test

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBTableReplica_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdynamodb.ResourceTableReplica(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDynamoDBTableReplica_pitr(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_pitr(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery", "true"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_tags1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
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
					testAccCheckTableReplicaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "5"),
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
					testAccCheckTableReplicaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_dynamodb_table_replica.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_tableClass(rName, "STANDARD"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicaExists(resourceName),
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
					testAccCheckTableReplicaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "table_class_override", "STANDARD_INFREQUENT_ACCESS"),
				),
			},
		},
	})
}

func testAccCheckTableReplicaDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn
	replicaRegion := aws.StringValue(conn.Config.Region)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_table_replica" {
			continue
		}

		log.Printf("[DEBUG] Checking if DynamoDB table replica %s was destroyed", rs.Primary.ID)

		if rs.Primary.ID == "" {
			return create.Error(names.DynamoDB, create.ErrActionCheckingDestroyed, tfdynamodb.ResNameTableReplica, rs.Primary.ID, errors.New("no ID"))
		}

		tableName, mainRegion, err := tfdynamodb.TableReplicaParseID(rs.Primary.ID)
		if err != nil {
			return create.Error(names.DynamoDB, create.ErrActionCheckingDestroyed, tfdynamodb.ResNameTableReplica, rs.Primary.ID, err)
		}

		session, err := conns.NewSessionForRegion(&conn.Config, mainRegion, acctest.Provider.Meta().(*conns.AWSClient).TerraformVersion)
		if err != nil {
			return create.Error(names.DynamoDB, create.ErrActionCheckingDestroyed, tfdynamodb.ResNameTableReplica, rs.Primary.ID, fmt.Errorf("region %s: %w", mainRegion, err))
		}

		conn = dynamodb.New(session) // now global table region

		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		}

		result, err := conn.DescribeTable(params)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return create.Error(names.DynamoDB, create.ErrActionCheckingDestroyed, tfdynamodb.ResNameTableReplica, rs.Primary.ID, err)
		}

		if result == nil || result.Table == nil {
			continue
		}

		if _, err := tfdynamodb.FilterReplicasByRegion(result.Table.Replicas, replicaRegion); err == nil {
			return create.Error(names.DynamoDB, create.ErrActionCheckingDestroyed, tfdynamodb.ResNameTableReplica, rs.Primary.ID, errors.New("still exists"))
		}

		return err
	}

	return nil
}

func testAccCheckTableReplicaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Printf("[DEBUG] Trying to create initial table replica state!")
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.DynamoDB, create.ErrActionCheckingExistence, tfdynamodb.ResNameTableReplica, rs.Primary.ID, fmt.Errorf("not found: %s", n))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DynamoDB, create.ErrActionCheckingExistence, tfdynamodb.ResNameTableReplica, rs.Primary.ID, errors.New("no ID"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn

		tableName, mainRegion, err := tfdynamodb.TableReplicaParseID(rs.Primary.ID)
		if err != nil {
			return create.Error(names.DynamoDB, create.ErrActionCheckingExistence, tfdynamodb.ResNameTableReplica, rs.Primary.ID, err)
		}

		session, err := conns.NewSessionForRegion(&conn.Config, mainRegion, acctest.Provider.Meta().(*conns.AWSClient).TerraformVersion)
		if err != nil {
			return create.Error(names.DynamoDB, create.ErrActionCheckingExistence, tfdynamodb.ResNameTableReplica, rs.Primary.ID, fmt.Errorf("region %s: %w", mainRegion, err))
		}

		conn = dynamodb.New(session) // now global table region

		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		}

		_, err = conn.DescribeTable(params)
		if err != nil {
			return create.Error(names.DynamoDB, create.ErrActionCheckingExistence, tfdynamodb.ResNameTableReplica, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccTableReplicaConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
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

resource "aws_dynamodb_table_replica" "test" {
  global_table_arn = aws_dynamodb_table.test.arn

  tags = {
    Name = %[1]q
    Pozo = "Amargo"
  }
}
`, rName))
}

func testAccTableReplicaConfig_pitr(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
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

resource "aws_dynamodb_table_replica" "test" {
  provider               = "awsalternate"
  global_table_arn       = aws_dynamodb_table.test.arn
  point_in_time_recovery = true
}
`, rName))
}

func testAccTableReplicaConfig_tags1(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
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

resource "aws_dynamodb_table_replica" "test" {
  provider         = "awsalternate"
  global_table_arn = aws_dynamodb_table.test.arn

  tags = {
    Name = %[1]q
    tape = "Valladolid"
  }
}
`, rName))
}

func testAccTableReplicaConfig_tags2(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
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

resource "aws_dynamodb_table_replica" "test" {
  provider         = "awsalternate"
  global_table_arn = aws_dynamodb_table.test.arn

  tags = {
    Name      = %[1]q
    tape      = "Valladolid"
    brightest = "Lights"
    arise     = "Melandru"
    shooting  = "Stars"
  }
}
`, rName))
}

func testAccTableReplicaConfig_tags3(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
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

resource "aws_dynamodb_table_replica" "test" {
  provider         = "awsalternate"
  global_table_arn = aws_dynamodb_table.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTableReplicaConfig_tableClass(rName, class string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
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

resource "aws_dynamodb_table_replica" "test" {
  global_table_arn     = aws_dynamodb_table.test.arn
  table_class_override = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, class))
}
