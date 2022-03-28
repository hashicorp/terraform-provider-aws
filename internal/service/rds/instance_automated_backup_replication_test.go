package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

const (
	dbInstanceAutomatedBackupReplicationRetained = "retained"
)

func TestAccRDSInstanceAutomatedBackupReplication_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_automated_backup_replication.test"

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInstanceAutomatedBackupReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAutomatedBackupReplicationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "retention_period", "7"),
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

func TestAccRDSInstanceAutomatedBackupReplication_retentionPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_automated_backup_replication.test"

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInstanceAutomatedBackupReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAutomatedBackupReplicationConfig_retentionPeriod(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "retention_period", "14"),
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

func TestAccRDSInstanceAutomatedBackupReplication_kmsEncrypted(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_automated_backup_replication.test"

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInstanceAutomatedBackupReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAutomatedBackupReplicationConfig_kmsEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "retention_period", "7"),
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

func testAccInstanceAutomatedBackupReplicationConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 10
  identifier              = %[1]q
  engine                  = "postgres"
  engine_version          = "13.4"
  instance_class          = "db.t3.micro"
  name                    = "mydb"
  username                = "masterusername"
  password                = "mustbeeightcharacters"
  backup_retention_period = 7
  skip_final_snapshot     = true
}

resource "aws_db_instance_automated_backup_replication" "test" {
  source_db_instance_arn = aws_db_instance.test.arn

  provider = "awsalternate"
}
`, rName))
}

func testAccInstanceAutomatedBackupReplicationConfig_retentionPeriod(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 10
  identifier              = %[1]q
  engine                  = "postgres"
  engine_version          = "13.4"
  instance_class          = "db.t3.micro"
  name                    = "mydb"
  username                = "masterusername"
  password                = "mustbeeightcharacters"
  backup_retention_period = 7
  skip_final_snapshot     = true
}

resource "aws_db_instance_automated_backup_replication" "test" {
  source_db_instance_arn = aws_db_instance.test.arn
  retention_period       = 14

  provider = "awsalternate"
}
`, rName))
}

func testAccInstanceAutomatedBackupReplicationConfig_kmsEncrypted(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
  provider    = "awsalternate"
}

resource "aws_db_instance" "test" {
  allocated_storage       = 10
  identifier              = %[1]q
  engine                  = "postgres"
  engine_version          = "13.4"
  instance_class          = "db.t3.micro"
  name                    = "mydb"
  username                = "masterusername"
  password                = "mustbeeightcharacters"
  backup_retention_period = 7
  storage_encrypted       = true
  skip_final_snapshot     = true
}

resource "aws_db_instance_automated_backup_replication" "test" {
  source_db_instance_arn = aws_db_instance.test.arn
  kms_key_id             = aws_kms_key.test.arn

  provider = "awsalternate"
}
`, rName))
}

func testAccCheckInstanceAutomatedBackupReplicationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_cluster" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeDBInstanceAutomatedBackups(
			&rds.DescribeDBInstanceAutomatedBackupsInput{
				DBInstanceAutomatedBackupsArn: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBInstanceAutomatedBackups) != 0 &&
				*resp.DBInstanceAutomatedBackups[0].DBInstanceAutomatedBackupsArn == rs.Primary.ID {
				return fmt.Errorf("DB Instance Automated Backup still exists, %s", rs.Primary.ID)
			}
		}

		// Return nil if the cluster is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == rds.ErrCodeDBInstanceAutomatedBackupNotFoundFault {
				return nil
			}
		}

		return err
	}

	return nil
}
