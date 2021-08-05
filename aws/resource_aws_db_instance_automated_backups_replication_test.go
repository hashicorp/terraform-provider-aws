package aws

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/finder"
)

func TestAccAWSDbInstanceAutomatedBackupsReplication_basic(t *testing.T) {
	var providers []*schema.Provider
	var dbInstanceAutomatedBackup rds.DBInstanceAutomatedBackup

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance_automated_backups_replication.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheck(t, rds.EndpointsID),
		ProviderFactories: testAccProviderFactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckAWSDbInstanceAutomatedBackupsReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDbInstanceAutomatedBackupsReplicationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDbInstanceAutomatedBackupsReplicationExists(resourceName, &dbInstanceAutomatedBackup),
					testAccCheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "source_db_instance_arn", "rds", regexp.MustCompile(`db:.+`).String()),
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

func TestAccAWSDbInstanceAutomatedBackupsReplication_disappears(t *testing.T) {
	var providers []*schema.Provider
	var dbInstanceAutomatedBackup rds.DBInstanceAutomatedBackup

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance_automated_backups_replication.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
			testAccAlternateAccountPreCheck(t)
		},
		ErrorCheck:        testAccErrorCheck(t, rds.EndpointsID),
		ProviderFactories: testAccProviderFactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckAWSDbInstanceAutomatedBackupsReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDbInstanceAutomatedBackupsReplicationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDbInstanceAutomatedBackupsReplicationExists(resourceName, &dbInstanceAutomatedBackup),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDbInstanceAutomatedBackupsReplication(), resourceName),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckAWSDbInstanceAutomatedBackupsReplicationExists(resourceName string, dbInstanceAutomatedBackup *rds.DBInstanceAutomatedBackup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		backup, err := finder.DBInstanceAutomatedBackup(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if backup == nil {
			return fmt.Errorf("RDS DB Instance Automated Backup not found (%s)", rs.Primary.ID)
		}

		*dbInstanceAutomatedBackup = *backup

		return nil
	}
}

func testAccCheckAWSDbInstanceAutomatedBackupsReplicationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance_automated_backups_replication" {
			continue
		}

		backup, err := finder.DBInstanceAutomatedBackup(context.Background(), conn, rs.Primary.ID)
		if isAWSErr(err, rds.ErrCodeDBInstanceAutomatedBackupNotFoundFault, "") {
			continue
		}

		if isAWSErr(err, rds.ErrCodeInvalidDBInstanceAutomatedBackupStateFault, "") {
			continue
		}

		if isAWSErr(err, rds.ErrCodeInvalidDBInstanceStateFault, "") {
			return nil
		}

		if err != nil {
			return err
		}

		if backup == nil {
			continue
		}

		return fmt.Errorf("RDS DB Instance Automated Backups Replication still exists %s", rs.Primary.ID)
	}

	return nil
}

func testAccAWSDbInstanceAutomatedBackupsReplicationConfig(rName string) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(2), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage        = 10
  engine                   = "postgres"
  identifier               = %[1]q
  instance_class           = "db.t3.micro"
  password                 = "avoid-plaintext-passwords"
  username                 = "tfacctest"
  skip_final_snapshot      = true
  delete_automated_backups = true
  backup_retention_period  = 1
  provider                 = awsalternate
}

resource "aws_db_instance_automated_backups_replication" "test" {
  source_db_instance_arn = aws_db_instance.test.arn
}
`, rName))
}
