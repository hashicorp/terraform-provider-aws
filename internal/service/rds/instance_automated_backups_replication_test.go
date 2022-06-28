package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRDSInstanceAutomatedBackupsReplication_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_automated_backups_replication.test"

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInstanceAutomatedBackupsReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAutomatedBackupsReplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAutomatedBackupsReplicationExist(resourceName),
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

func TestAccRDSInstanceAutomatedBackupsReplication_retentionPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_automated_backups_replication.test"

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInstanceAutomatedBackupsReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAutomatedBackupsReplicationConfig_retentionPeriod(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAutomatedBackupsReplicationExist(resourceName),
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

func TestAccRDSInstanceAutomatedBackupsReplication_kmsEncrypted(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_automated_backups_replication.test"

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInstanceAutomatedBackupsReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAutomatedBackupsReplicationConfig_kmsEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAutomatedBackupsReplicationExist(resourceName),
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

func testAccInstanceAutomatedBackupsReplicationConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
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

  provider = "awsalternate"
}

resource "aws_db_instance_automated_backups_replication" "test" {
  source_db_instance_arn = aws_db_instance.test.arn
}
`, rName))
}

func testAccInstanceAutomatedBackupsReplicationConfig_retentionPeriod(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
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

  provider = "awsalternate"
}

resource "aws_db_instance_automated_backups_replication" "test" {
  source_db_instance_arn = aws_db_instance.test.arn
  retention_period       = 14
}
`, rName))
}

func testAccInstanceAutomatedBackupsReplicationConfig_kmsEncrypted(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
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

  provider = "awsalternate"
}

resource "aws_db_instance_automated_backups_replication" "test" {
  source_db_instance_arn = aws_db_instance.test.arn
  kms_key_id             = aws_kms_key.test.arn
}
`, rName))
}

func testAccCheckInstanceAutomatedBackupsReplicationExist(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS instance automated backups replication ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		_, err := tfrds.FindDBInstanceAutomatedBackupByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckInstanceAutomatedBackupsReplicationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance_automated_backups_replication" {
			continue
		}

		_, err := tfrds.FindDBInstanceAutomatedBackupByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("RDS instance automated backups replication %s still exists", rs.Primary.ID)
	}

	return nil
}
