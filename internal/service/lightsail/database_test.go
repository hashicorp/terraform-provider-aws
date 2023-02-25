package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailDatabase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	resourceName := "aws_lightsail_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "relational_database_name", rName),
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

func TestAccLightsailDatabase_relationalDatabaseName(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	resourceName := "aws_lightsail_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameTooShort := "s"
	rNameTooLong := fmt.Sprintf("%s-%s", rName, sdkacctest.RandString(255))
	rNameContainsUnderscore := fmt.Sprintf("%s-%s", rName, "_test")
	rNameStartingDash := fmt.Sprintf("%s-%s", "-", rName)
	rNameEndingDash := fmt.Sprintf("%s-%s", rName, "-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_basic(rNameTooShort),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected length of relational_database_name to be in the range \(2 - 255\), got %s`, rNameTooShort)),
			},
			{
				Config:      testAccDatabaseConfig_basic(rNameTooLong),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected length of relational_database_name to be in the range \(2 - 255\), got %s`, rNameTooLong)),
			},
			{
				Config:      testAccDatabaseConfig_basic(rNameContainsUnderscore),
				ExpectError: regexp.MustCompile(`Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number`),
			},
			{
				Config:      testAccDatabaseConfig_basic(rNameStartingDash),
				ExpectError: regexp.MustCompile(`Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number`),
			},
			{
				Config:      testAccDatabaseConfig_basic(rNameEndingDash),
				ExpectError: regexp.MustCompile(`Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number`),
			},
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "relational_database_name", rName),
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

func TestAccLightsailDatabase_masterDatabaseName(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"
	dbName := "randomdatabasename"
	dbNameTooShort := ""
	dbNameTooLong := fmt.Sprintf("%s-%s", dbName, sdkacctest.RandString(64))
	dbNameContainsSpaces := fmt.Sprint(dbName, "string with spaces")
	dbNameContainsStartingDigit := fmt.Sprintf("01_%s", dbName)
	dbNameContainsUnderscore := fmt.Sprintf("%s_123456", dbName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_masterDatabaseName(rName, dbNameTooShort),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected length of master_database_name to be in the range \(1 - 64\), got %s`, dbNameTooShort)),
			},
			{
				Config:      testAccDatabaseConfig_masterDatabaseName(rName, dbNameTooLong),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected length of master_database_name to be in the range \(1 - 64\), got %s`, dbNameTooLong)),
			},
			{
				Config:      testAccDatabaseConfig_masterDatabaseName(rName, dbNameContainsSpaces),
				ExpectError: regexp.MustCompile(`Subsequent characters can be letters, underscores, or digits \(0- 9\)`),
			},
			{
				Config:      testAccDatabaseConfig_masterDatabaseName(rName, dbNameContainsStartingDigit),
				ExpectError: regexp.MustCompile(`Must begin with a letter`),
			},
			{
				Config: testAccDatabaseConfig_masterDatabaseName(rName, dbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", dbName),
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
				Config: testAccDatabaseConfig_masterDatabaseName(rName, dbNameContainsUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", dbNameContainsUnderscore),
				),
			},
		},
	})
}

func TestAccLightsailDatabase_masterUsername(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"
	username := "username1"
	usernameTooShort := ""
	usernameTooLong := fmt.Sprintf("%s-%s", username, sdkacctest.RandString(63))
	usernameStartingDigit := fmt.Sprintf("01%s", username)
	usernameContainsDash := fmt.Sprintf("%s-test", username)
	usernameContainsSpecial := fmt.Sprintf("%s@", username)
	usernameContainsUndercore := fmt.Sprintf("%s_test", username)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_masterUsername(rName, usernameTooShort),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected length of master_username to be in the range \(1 - 63\), got %s`, usernameTooShort)),
			},
			{
				Config:      testAccDatabaseConfig_masterUsername(rName, usernameTooLong),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected length of master_username to be in the range \(1 - 63\), got %s`, usernameTooLong)),
			},
			{
				Config:      testAccDatabaseConfig_masterUsername(rName, usernameStartingDigit),
				ExpectError: regexp.MustCompile(`Must begin with a letter`),
			},
			{
				Config:      testAccDatabaseConfig_masterUsername(rName, usernameContainsDash),
				ExpectError: regexp.MustCompile(`Subsequent characters can be letters, underscores, or digits \(0- 9\)`),
			},
			{
				Config:      testAccDatabaseConfig_masterUsername(rName, usernameContainsSpecial),
				ExpectError: regexp.MustCompile(`Subsequent characters can be letters, underscores, or digits \(0- 9\)`),
			},
			{
				Config: testAccDatabaseConfig_masterUsername(rName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_username", username),
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
				Config: testAccDatabaseConfig_masterUsername(rName, usernameContainsUndercore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_username", usernameContainsUndercore),
				),
			},
		},
	})
}

func TestAccLightsailDatabase_masterPassword(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	password := "testpassword"
	passwordTooShort := "short"
	passwordTooLong := fmt.Sprintf("%s-%s", password, sdkacctest.RandString(128))
	passwordContainsSlash := fmt.Sprintf("%s/", password)
	passwordContainsQuotes := fmt.Sprintf("%s\"", password)
	passwordContainsAtSymbol := fmt.Sprintf("%s@", password)
	passwordContainsSpaces := fmt.Sprintf("%s spaces here", password)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordTooShort),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected length of master_password to be in the range \(8 - 128\), got %s`, passwordTooShort)),
			},
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordTooLong),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected length of master_password to be in the range \(8 - 128\), got %s`, passwordTooLong)),
			},
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordContainsSlash),
				ExpectError: regexp.MustCompile(`The password can include any printable ASCII character except \"/\", \"\"\", or \"@\". It cannot contain spaces.`),
			},
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordContainsQuotes),
				ExpectError: regexp.MustCompile(`The password can include any printable ASCII character except \"/\", \"\"\", or \"@\". It cannot contain spaces.`),
			},
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordContainsAtSymbol),
				ExpectError: regexp.MustCompile(`The password can include any printable ASCII character except \"/\", \"\"\", or \"@\". It cannot contain spaces.`),
			},
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordContainsSpaces),
				ExpectError: regexp.MustCompile(`The password can include any printable ASCII character except \"/\", \"\"\", or \"@\". It cannot contain spaces.`),
			},
		},
	})
}

func TestAccLightsailDatabase_preferredBackupWindow(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"
	backupWindowInvalidHour := "25:30-10:00"
	backupWindowInvalidMinute := "10:00-10:70"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_preferredBackupWindow(rName, backupWindowInvalidHour),
				ExpectError: regexp.MustCompile(`must satisfy the format of \"hh24:mi-hh24:mi\".`),
			},
			{
				Config:      testAccDatabaseConfig_preferredBackupWindow(rName, backupWindowInvalidMinute),
				ExpectError: regexp.MustCompile(`must satisfy the format of \"hh24:mi-hh24:mi\".`),
			},
			{
				Config: testAccDatabaseConfig_preferredBackupWindow(rName, "09:30-10:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
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
				Config: testAccDatabaseConfig_preferredBackupWindow(rName, "09:45-10:15"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "09:45-10:15"),
				),
			},
		},
	})
}

func TestAccLightsailDatabase_preferredMaintenanceWindow(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"
	maintenanceWindowInvalidDay := "tuesday:04:30-tue:05:00"
	maintenanceWindowInvalidHour := "tue:04:30-tue:30:00"
	maintenanceWindowInvalidMinute := "tue:04:85-tue:05:00"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_preferredMaintenanceWindow(rName, maintenanceWindowInvalidDay),
				ExpectError: regexp.MustCompile(`must satisfy the format of \"ddd:hh24:mi-ddd:hh24:mi\".`),
			},
			{
				Config:      testAccDatabaseConfig_preferredMaintenanceWindow(rName, maintenanceWindowInvalidHour),
				ExpectError: regexp.MustCompile(`must satisfy the format of \"ddd:hh24:mi-ddd:hh24:mi\".`),
			},
			{
				Config:      testAccDatabaseConfig_preferredMaintenanceWindow(rName, maintenanceWindowInvalidMinute),
				ExpectError: regexp.MustCompile(`must satisfy the format of \"ddd:hh24:mi-ddd:hh24:mi\".`),
			},
			{
				Config: testAccDatabaseConfig_preferredMaintenanceWindow(rName, "tue:04:30-tue:05:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
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
				Config: testAccDatabaseConfig_preferredMaintenanceWindow(rName, "wed:06:00-wed:07:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "wed:06:00-wed:07:30"),
				),
			},
		},
	})
}

func TestAccLightsailDatabase_publiclyAccessible(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_publiclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
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
				Config: testAccDatabaseConfig_publiclyAccessible(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
				),
			},
		},
	})
}

func TestAccLightsailDatabase_backupRetentionEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_backupRetentionEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
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
				Config: testAccDatabaseConfig_backupRetentionEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_enabled", "false"),
				),
			},
		},
	})
}

func TestAccLightsailDatabase_finalSnapshotName(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"
	sName := fmt.Sprintf("%s-snapshot", rName)
	sNameTooShort := "s"
	sNameTooLong := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(255))
	sNameContainsSpaces := fmt.Sprint(sName, "string with spaces")
	sNameContainsUnderscore := fmt.Sprintf("%s_123456", sName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_finalSnapshotName(rName, sNameTooShort),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected length of final_snapshot_name to be in the range \(2 - 255\), got %s`, sNameTooShort)),
			},
			{
				Config:      testAccDatabaseConfig_finalSnapshotName(rName, sNameTooLong),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected length of final_snapshot_name to be in the range \(2 - 255\), got %s`, sNameTooLong)),
			},
			{
				Config:      testAccDatabaseConfig_finalSnapshotName(rName, sNameContainsSpaces),
				ExpectError: regexp.MustCompile(`Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number`),
			},
			{
				Config:      testAccDatabaseConfig_finalSnapshotName(rName, sNameContainsUnderscore),
				ExpectError: regexp.MustCompile(`Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number`),
			},
			{
				Config: testAccDatabaseConfig_finalSnapshotName(rName, sName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
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

func TestAccLightsailDatabase_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var db1, db2, db3 lightsail.RelationalDatabase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db1),
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
				Config: testAccDatabaseConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDatabaseConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccLightsailDatabase_ha(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	resourceName := "aws_lightsail_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_ha(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "relational_database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "bundle_id", "micro_ha_1_0"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
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

func TestAccLightsailDatabase_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var db lightsail.RelationalDatabase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Database
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		_, err := conn.DeleteRelationalDatabaseWithContext(ctx, &lightsail.DeleteRelationalDatabaseInput{
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
			testAccPreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName, &db),
					testDestroy),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDatabaseExists(ctx context.Context, n string, res *lightsail.RelationalDatabase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Database ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetRelationalDatabaseWithContext(ctx, &params)

		if err != nil {
			return err
		}

		if resp == nil || resp.RelationalDatabase == nil {
			return fmt.Errorf("Database (%s) not found", rs.Primary.ID)
		}
		*res = *resp.RelationalDatabase
		return nil
	}
}

func testAccCheckDatabaseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_database" {
				continue
			}

			params := lightsail.GetRelationalDatabaseInput{
				RelationalDatabaseName: aws.String(rs.Primary.ID),
			}

			respDatabase, err := conn.GetRelationalDatabaseWithContext(ctx, &params)

			if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
				continue
			}

			if err == nil {
				if respDatabase.RelationalDatabase != nil {
					return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResNameDatabase, rs.Primary.ID, errors.New("still exists"))
				}
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResNameDatabase, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccCheckDatabaseSnapshotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_database" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			snapshot_identifier := rs.Primary.Attributes["final_snapshot_name"]

			log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
			_, err := conn.DeleteRelationalDatabaseSnapshotWithContext(ctx, &lightsail.DeleteRelationalDatabaseSnapshotInput{
				RelationalDatabaseSnapshotName: aws.String(snapshot_identifier),
			})

			if err != nil {
				return err
			}

			params := lightsail.GetRelationalDatabaseInput{
				RelationalDatabaseName: aws.String(rs.Primary.ID),
			}

			respDatabase, err := conn.GetRelationalDatabaseWithContext(ctx, &params)

			if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
				continue
			}

			if err == nil {
				if respDatabase.RelationalDatabase != nil {
					return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResNameDatabase, rs.Primary.ID, errors.New("still exists"))
				}
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResNameDatabase, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccDatabaseConfig_base() string {
	return acctest.ConfigAvailableAZsNoOptIn()
}

func testAccDatabaseConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "testdatabasename"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  skip_final_snapshot      = true
}
`, rName))
}

func testAccDatabaseConfig_masterDatabaseName(rName string, masterDatabaseName string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = %[2]q
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  skip_final_snapshot      = true
}
`, rName, masterDatabaseName))
}

func testAccDatabaseConfig_masterUsername(rName string, masterUsername string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "testdatabasename"
  master_password          = "testdatabasepassword"
  master_username          = %[2]q
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  skip_final_snapshot      = true
}
`, rName, masterUsername))
}

func testAccDatabaseConfig_masterPassword(rName string, masterPassword string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "testdatabasename"
  master_password          = %[2]q
  master_username          = "testusername"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  skip_final_snapshot      = true
}
`, rName, masterPassword))
}

func testAccDatabaseConfig_preferredBackupWindow(rName string, preferredBackupWindow string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "testdatabasename"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  preferred_backup_window  = %[2]q
  apply_immediately        = true
  skip_final_snapshot      = true
}
`, rName, preferredBackupWindow))
}

func testAccDatabaseConfig_preferredMaintenanceWindow(rName string, preferredMaintenanceWindow string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name     = %[1]q
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

func testAccDatabaseConfig_publiclyAccessible(rName string, publiclyAccessible bool) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "testdatabasename"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  publicly_accessible      = %[2]t
  apply_immediately        = true
  skip_final_snapshot      = true
}
`, rName, publiclyAccessible))
}

func testAccDatabaseConfig_backupRetentionEnabled(rName string, backupRetentionEnabled bool) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "test"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  backup_retention_enabled = %[2]t
  apply_immediately        = true
  skip_final_snapshot      = true
}
`, rName, backupRetentionEnabled))
}

func testAccDatabaseConfig_finalSnapshotName(rName string, sName string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "test"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  final_snapshot_name      = %[2]q
}
`, rName, sName))
}

func testAccDatabaseConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "testdatabasename"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  skip_final_snapshot      = true
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDatabaseConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "testdatabasename"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  skip_final_snapshot      = true
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDatabaseConfig_ha(rName string) string {
	return acctest.ConfigCompose(
		testAccDatabaseConfig_base(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  relational_database_name = %[1]q
  master_database_name     = "testdatabasename"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_ha_1_0"
  skip_final_snapshot      = true
}
`, rName))
}
