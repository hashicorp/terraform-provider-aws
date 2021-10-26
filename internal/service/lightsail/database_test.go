package lightsail_test

import (
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDatabase_basic(t *testing.T) {
	var db lightsail.RelationalDatabase
	resourceName := "aws_lightsail_database.test"
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "blueprint_id", "mysql_8_0"),
					resource.TestCheckResourceAttr(resourceName, "bundle_id", "micro_1_0"),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", "testdatabasename"),
					resource.TestCheckResourceAttr(resourceName, "master_username", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "cpu_count"),
					resource.TestCheckResourceAttrSet(resourceName, "ram_size"),
					resource.TestCheckResourceAttrSet(resourceName, "disk_size"),
					resource.TestCheckResourceAttrSet(resourceName, "master_endpoint_port"),
					resource.TestCheckResourceAttrSet(resourceName, "master_endpoint_address"),
					resource.TestCheckResourceAttrSet(resourceName, "support_code"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func TestAccDatabase_Name(t *testing.T) {
	var db lightsail.RelationalDatabase
	resourceName := "aws_lightsail_database.test"
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:    acctest.ErrorCheck(t, lightsail.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     acctest.Providers,
		CheckDestroy:  testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func TestAccDatabase_MasterDatabaseName(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigMasterDatabaseName(rName, "databasename1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", "databasename1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigMasterDatabaseName(rName, "databasename2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", "databasename2"),
				),
			},
		},
	})
}

func TestAccDatabase_MasterUsername(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigMasterUsername(rName, "username1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_username", "username1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigMasterUsername(rName, "username2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_username", "username2"),
				),
			},
		},
	})
}

func TestAccDatabase_PreferredBackupWindow(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigPreferredBackupWindow(rName, "09:30-10:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "09:30-10:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigPreferredBackupWindow(rName, "09:45-10:15"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "09:45-10:15"),
				),
			},
		},
	})
}

func TestAccDatabase_PreferredMaintenanceWindow(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigPreferredMaintenanceWindow(rName, "tue:04:30-tue:05:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "tue:04:30-tue:05:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigPreferredMaintenanceWindow(rName, "wed:06:00-wed:07:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "wed:06:00-wed:07:30"),
				),
			},
		},
	})
}

func TestAccDatabase_PubliclyAccessible(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigPubliclyAccessible(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigPubliclyAccessible(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
				),
			},
		},
	})
}

func TestAccDatabase_BackupRetentionEnabled(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigBackupRetentionEnabled(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigBackupRetentionEnabled(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_enabled", "false"),
				),
			},
		},
	})
}

func TestAccDatabase_FinalSnapshotName(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	resourceName := "aws_lightsail_database.test"
	sName := fmt.Sprintf("%s-snapshot", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDatabaseSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigFinalSnapshotName(rName, sName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func TestAccDatabase_Tags(t *testing.T) {
	var db1, db2, db3 lightsail.RelationalDatabase
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		IDRefreshName: resourceName,
		ErrorCheck:    acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:     acctest.Providers,
		CheckDestroy:  testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDatabaseConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDatabase_disappears(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	resourceName := "aws_lightsail_database.test"

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Database
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		_, err := conn.DeleteRelationalDatabase(&lightsail.DeleteRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rName),
			SkipFinalSnapshot:      aws.Bool(true),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Database in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		Providers:    acctest.Providers,
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(resourceName, &db),
					testDestroy),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSDatabaseExists(n string, res *lightsail.RelationalDatabase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Database ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetRelationalDatabase(&params)

		if err != nil {
			return err
		}

		if resp == nil || resp.RelationalDatabase == nil {
			return fmt.Errorf("Database (%s) not found", rs.Primary.Attributes["name"])
		}
		*res = *resp.RelationalDatabase
		return nil
	}
}

func testAccCheckAWSDatabaseDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_database" {
			continue
		}

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetRelationalDatabase(&params)

		if err == nil {
			return fmt.Errorf("Lightsail Database %q still exists", rs.Primary.ID)
		}

		// Verify the error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccCheckAWSDatabaseSnapshotDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_database" {
			continue
		}

		// Try and delete the snapshot before we check for the cluster not found
		snapshot_identifier := fmt.Sprintf("%s-snapshot", rs.Primary.ID)

		log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
		_, err := conn.DeleteRelationalDatabaseSnapshot(
			&lightsail.DeleteRelationalDatabaseSnapshotInput{
				RelationalDatabaseSnapshotName: aws.String(snapshot_identifier),
			})

		if err != nil {
			return err
		}

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		_, err = conn.GetRelationalDatabase(&params)

		if err == nil {
			return fmt.Errorf("Lightsail Database %q still exists", rs.Primary.ID)
		}

		// Verify the error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccDatabaseConfigBase() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`
}

func testAccDatabaseConfigBasic(rName string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
}
`, rName))
}

func testAccDatabaseConfigMasterDatabaseName(rName string, masterDatabaseName string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = %[2]q
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
}
`, rName, masterDatabaseName))
}

func testAccDatabaseConfigMasterUsername(rName string, masterUsername string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = %[2]q
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
}
`, rName, masterUsername))
}

func testAccDatabaseConfigPreferredBackupWindow(rName string, preferredBackupWindow string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                    = %[1]q
  availability_zone       = data.aws_availability_zones.available.names[0]
  master_database_name    = "testdatabasename"
  master_password         = "testdatabasepassword"
  master_username         = "test"
  blueprint_id            = "mysql_8_0"
  bundle_id               = "micro_1_0"
  preferred_backup_window = %[2]q
  apply_immediately       = true
  skip_final_snapshot     = true
}
`, rName, preferredBackupWindow))
}

func testAccDatabaseConfigPreferredMaintenanceWindow(rName string, preferredMaintenanceWindow string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                         = %[1]q
  availability_zone            = data.aws_availability_zones.available.names[0]
  master_database_name         = "testdatabasename"
  master_password              = "testdatabasepassword"
  master_username              = "test"
  blueprint_id                 = "mysql_8_0"
  bundle_id                    = "micro_1_0"
  preferred_maintenance_window = %[2]q
  apply_immediately            = true
  skip_final_snapshot          = true
}
`, rName, preferredMaintenanceWindow))
}

func testAccDatabaseConfigPubliclyAccessible(rName string, publiclyAccessible string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  publicly_accessible  = %[2]q
  apply_immediately    = true
  skip_final_snapshot  = true
}
`, rName, publiclyAccessible))
}

func testAccDatabaseConfigBackupRetentionEnabled(rName string, backupRetentionEnabled string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                     = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "test"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  backup_retention_enabled = %[2]q
  apply_immediately        = true
  skip_final_snapshot      = true
}
`, rName, backupRetentionEnabled))
}

func testAccDatabaseConfigFinalSnapshotName(rName string, sName string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "test"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  final_snapshot_name  = %[2]q
}
`, rName, sName))
}

func testAccDatabaseConfigTags1(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDatabaseConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
