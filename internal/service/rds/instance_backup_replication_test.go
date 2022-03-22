package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRDSInstanceBackupReplication_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_backup_replication.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInstanceRoleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceBackupReplicationConfig(rName),
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

func TestAccRDSInstanceBackupReplication_retentionPeriod(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_backup_replication.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInstanceBackupReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceBackupReplicationConfig_RetentionPeriod(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "retention_period", "30"),
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

func testAccInstanceBackupReplicationConfig(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

provider "aws" {
  region = "us-west-2"
  alias =  "replica"
}

resource "aws_kms_key" "test" {
  description             = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage    = 10
  identifier           = %[1]q
  engine               = "postgresql"
  engine_version       = "13.4"
  instance_class       = "db.t3.micro"
  name                 = "mydb"
  username             = "masterusername"
  password             = "mustbeeightcharacters"
  skip_final_snapshot  = true
}

resource "aws_db_instance_backup_replication" "default" {
  source_db_instance_arn = aws_db_instance.test.arn
  kms_key_id             = aws_kms_key.test.arn

  provider = "aws.replica"
}
`, rName)
}

func testAccInstanceBackupReplicationConfig_RetentionPeriod(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}
	
provider "aws" {
  region = "us-west-2"
  alias =  "replica"
}
	  
resource "aws_kms_key" "test" {
  description             = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage    = 10
  identifier           = %[1]q
  engine               = "postgresql"
  engine_version       = "13.4"
  instance_class       = "db.t3.micro"
  name                 = "mydb"
  username             = "masterusername"
  password             = "mustbeeightcharacters"
  skip_final_snapshot  = true
}

resource "aws_db_instance_backup_replication" "default" {
  source_db_instance_arn = aws_db_instance.test.arn
  kms_key_id             = aws_kms_key.test.arn
  retention_period       = 30

  provider = "aws.replica"
}
`, rName)
}

func testAccCheckInstanceBackupReplicationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance_backup_replication" {
			continue
		}

		input := &rds.DescribeDBInstanceAutomatedBackupsInput{
			DBInstanceAutomatedBackupsArn: &rs.Primary.ID,
		}

		output, err := conn.DescribeDBInstanceAutomatedBackups(input)

		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceAutomatedBackupNotFoundFault) {
			continue
		}

		if err != nil {
			return err
		}

		if output == nil {
			continue
		}

		return fmt.Errorf("RDS instance backup replication %q still exists", rs.Primary.ID)
	}

	return nil
}
