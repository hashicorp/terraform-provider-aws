// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDatabase_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "relational_database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "blueprint_id", "mysql_8_0"),
					resource.TestCheckResourceAttr(resourceName, "bundle_id", "micro_2_0"),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", "testdatabasename"),
					resource.TestCheckResourceAttr(resourceName, "master_username", "test"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
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
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func testAccDatabase_relationalDatabaseName(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameTooShort := "s"
	rNameTooLong := fmt.Sprintf("%s-%s", rName, sdkacctest.RandString(255))
	rNameContainsUnderscore := fmt.Sprintf("%s-%s", rName, "_test")
	rNameStartingDash := fmt.Sprintf("%s-%s", "-", rName)
	rNameEndingDash := fmt.Sprintf("%s-%s", rName, "-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_basic(rNameTooShort),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`expected length of relational_database_name to be in the range \(2 - 255\), got %s`, rNameTooShort)),
			},
			{
				Config:      testAccDatabaseConfig_basic(rNameTooLong),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`expected length of relational_database_name to be in the range \(2 - 255\), got %s`, rNameTooLong)),
			},
			{
				Config:      testAccDatabaseConfig_basic(rNameContainsUnderscore),
				ExpectError: regexache.MustCompile(`Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number`),
			},
			{
				Config:      testAccDatabaseConfig_basic(rNameStartingDash),
				ExpectError: regexache.MustCompile(`Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number`),
			},
			{
				Config:      testAccDatabaseConfig_basic(rNameEndingDash),
				ExpectError: regexache.MustCompile(`Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number`),
			},
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "relational_database_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func testAccDatabase_masterDatabaseName(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
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
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_masterDatabaseName(rName, dbNameTooShort),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`expected length of master_database_name to be in the range \(1 - 64\), got %s`, dbNameTooShort)),
			},
			{
				Config:      testAccDatabaseConfig_masterDatabaseName(rName, dbNameTooLong),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`expected length of master_database_name to be in the range \(1 - 64\), got %s`, dbNameTooLong)),
			},
			{
				Config:      testAccDatabaseConfig_masterDatabaseName(rName, dbNameContainsSpaces),
				ExpectError: regexache.MustCompile(`Subsequent characters can be letters, underscores, or digits \(0- 9\)`),
			},
			{
				Config:      testAccDatabaseConfig_masterDatabaseName(rName, dbNameContainsStartingDigit),
				ExpectError: regexache.MustCompile(`Must begin with a letter`),
			},
			{
				Config: testAccDatabaseConfig_masterDatabaseName(rName, dbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", dbName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfig_masterDatabaseName(rName, dbNameContainsUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", dbNameContainsUnderscore),
				),
			},
		},
	})
}

func testAccDatabase_masterUsername(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
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
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_masterUsername(rName, usernameTooShort),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`expected length of master_username to be in the range \(1 - 63\), got %s`, usernameTooShort)),
			},
			{
				Config:      testAccDatabaseConfig_masterUsername(rName, usernameTooLong),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`expected length of master_username to be in the range \(1 - 63\), got %s`, usernameTooLong)),
			},
			{
				Config:      testAccDatabaseConfig_masterUsername(rName, usernameStartingDigit),
				ExpectError: regexache.MustCompile(`Must begin with a letter`),
			},
			{
				Config:      testAccDatabaseConfig_masterUsername(rName, usernameContainsDash),
				ExpectError: regexache.MustCompile(`Subsequent characters can be letters, underscores, or digits \(0- 9\)`),
			},
			{
				Config:      testAccDatabaseConfig_masterUsername(rName, usernameContainsSpecial),
				ExpectError: regexache.MustCompile(`Subsequent characters can be letters, underscores, or digits \(0- 9\)`),
			},
			{
				Config: testAccDatabaseConfig_masterUsername(rName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "master_username", username),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfig_masterUsername(rName, usernameContainsUndercore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "master_username", usernameContainsUndercore),
				),
			},
		},
	})
}

func testAccDatabase_masterPassword(t *testing.T, semaphore tfsync.Semaphore) {
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
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordTooShort),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`expected length of master_password to be in the range \(8 - 128\), got %s`, passwordTooShort)),
			},
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordTooLong),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`expected length of master_password to be in the range \(8 - 128\), got %s`, passwordTooLong)),
			},
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordContainsSlash),
				ExpectError: regexache.MustCompile(`The password can include any printable ASCII character except \"/\", \"\"\", or \"@\". It cannot contain spaces.`),
			},
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordContainsQuotes),
				ExpectError: regexache.MustCompile(`The password can include any printable ASCII character except \"/\", \"\"\", or \"@\". It cannot contain spaces.`),
			},
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordContainsAtSymbol),
				ExpectError: regexache.MustCompile(`The password can include any printable ASCII character except \"/\", \"\"\", or \"@\". It cannot contain spaces.`),
			},
			{
				Config:      testAccDatabaseConfig_masterPassword(rName, passwordContainsSpaces),
				ExpectError: regexache.MustCompile(`The password can include any printable ASCII character except \"/\", \"\"\", or \"@\". It cannot contain spaces.`),
			},
		},
	})
}

func testAccDatabase_preferredBackupWindow(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"
	backupWindowInvalidHour := "25:30-10:00"
	backupWindowInvalidMinute := "10:00-10:70"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_preferredBackupWindow(rName, backupWindowInvalidHour),
				ExpectError: regexache.MustCompile(`must satisfy the format of \"hh24:mi-hh24:mi\"`),
			},
			{
				Config:      testAccDatabaseConfig_preferredBackupWindow(rName, backupWindowInvalidMinute),
				ExpectError: regexache.MustCompile(`must satisfy the format of \"hh24:mi-hh24:mi\"`),
			},
			{
				Config: testAccDatabaseConfig_preferredBackupWindow(rName, "09:30-10:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "09:30-10:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfig_preferredBackupWindow(rName, "09:45-10:15"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "09:45-10:15"),
				),
			},
		},
	})
}

func testAccDatabase_preferredMaintenanceWindow(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"
	maintenanceWindowInvalidDay := "tuesday:04:30-tue:05:00"
	maintenanceWindowInvalidHour := "tue:04:30-tue:30:00"
	maintenanceWindowInvalidMinute := "tue:04:85-tue:05:00"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_preferredMaintenanceWindow(rName, maintenanceWindowInvalidDay),
				ExpectError: regexache.MustCompile(`must satisfy the format of \"ddd:hh24:mi-ddd:hh24:mi\"`),
			},
			{
				Config:      testAccDatabaseConfig_preferredMaintenanceWindow(rName, maintenanceWindowInvalidHour),
				ExpectError: regexache.MustCompile(`must satisfy the format of \"ddd:hh24:mi-ddd:hh24:mi\"`),
			},
			{
				Config:      testAccDatabaseConfig_preferredMaintenanceWindow(rName, maintenanceWindowInvalidMinute),
				ExpectError: regexache.MustCompile(`must satisfy the format of \"ddd:hh24:mi-ddd:hh24:mi\"`),
			},
			{
				Config: testAccDatabaseConfig_preferredMaintenanceWindow(rName, "tue:04:30-tue:05:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "tue:04:30-tue:05:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfig_preferredMaintenanceWindow(rName, "wed:06:00-wed:07:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPreferredMaintenanceWindow, "wed:06:00-wed:07:30"),
				),
			},
		},
	})
}

func testAccDatabase_publiclyAccessible(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_publiclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfig_publiclyAccessible(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccDatabase_backupRetentionEnabled(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_backupRetentionEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfig_backupRetentionEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccDatabase_finalSnapshotName(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"
	sName := fmt.Sprintf("%s-snapshot", rName)
	sNameTooShort := "s"
	sNameTooLong := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(255))
	sNameContainsSpaces := fmt.Sprint(sName, "string with spaces")
	sNameContainsUnderscore := fmt.Sprintf("%s_123456", sName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_finalSnapshotName(rName, sNameTooShort),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`expected length of final_snapshot_name to be in the range \(2 - 255\), got %s`, sNameTooShort)),
			},
			{
				Config:      testAccDatabaseConfig_finalSnapshotName(rName, sNameTooLong),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`expected length of final_snapshot_name to be in the range \(2 - 255\), got %s`, sNameTooLong)),
			},
			{
				Config:      testAccDatabaseConfig_finalSnapshotName(rName, sNameContainsSpaces),
				ExpectError: regexache.MustCompile(`Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number`),
			},
			{
				Config:      testAccDatabaseConfig_finalSnapshotName(rName, sNameContainsUnderscore),
				ExpectError: regexache.MustCompile(`Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number`),
			},
			{
				Config: testAccDatabaseConfig_finalSnapshotName(rName, sName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func testAccDatabase_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDatabaseConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccDatabase_keyOnlyTags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_tags1(rName, acctest.CtKey1, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
			{
				Config: testAccDatabaseConfig_tags1(rName, acctest.CtKey2, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
		},
	})
}

func testAccDatabase_ha(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_ha(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "relational_database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "bundle_id", "micro_ha_2_0"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func testAccDatabase_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_database.test"

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Database
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		_, err := conn.DeleteRelationalDatabase(ctx, &lightsail.DeleteRelationalDatabaseInput{
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
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					testDestroy),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDatabaseExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Database ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetRelationalDatabase(ctx, &params)

		if err != nil {
			return err
		}

		if resp == nil || resp.RelationalDatabase == nil {
			return fmt.Errorf("Database (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDatabaseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_database" {
				continue
			}

			params := lightsail.GetRelationalDatabaseInput{
				RelationalDatabaseName: aws.String(rs.Primary.ID),
			}

			respDatabase, err := conn.GetRelationalDatabase(ctx, &params)

			if tflightsail.IsANotFoundError(err) {
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_database" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			snapshot_identifier := rs.Primary.Attributes["final_snapshot_name"]

			log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
			_, err := conn.DeleteRelationalDatabaseSnapshot(ctx, &lightsail.DeleteRelationalDatabaseSnapshotInput{
				RelationalDatabaseSnapshotName: aws.String(snapshot_identifier),
			})

			if err != nil {
				return err
			}

			params := lightsail.GetRelationalDatabaseInput{
				RelationalDatabaseName: aws.String(rs.Primary.ID),
			}

			respDatabase, err := conn.GetRelationalDatabase(ctx, &params)

			if tflightsail.IsANotFoundError(err) {
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
  bundle_id                = "micro_2_0"
  skip_final_snapshot      = true
}
`, rName))
}

func testAccDatabaseConfig_masterDatabaseName(rName, masterDatabaseName string) string {
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
  bundle_id                = "micro_2_0"
  skip_final_snapshot      = true
}
`, rName, masterDatabaseName))
}

func testAccDatabaseConfig_masterUsername(rName, masterUsername string) string {
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
  bundle_id                = "micro_2_0"
  skip_final_snapshot      = true
}
`, rName, masterUsername))
}

func testAccDatabaseConfig_masterPassword(rName, masterPassword string) string {
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
  bundle_id                = "micro_2_0"
  skip_final_snapshot      = true
}
`, rName, masterPassword))
}

func testAccDatabaseConfig_preferredBackupWindow(rName, preferredBackupWindow string) string {
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
  bundle_id                = "micro_2_0"
  preferred_backup_window  = %[2]q
  apply_immediately        = true
  skip_final_snapshot      = true
}
`, rName, preferredBackupWindow))
}

func testAccDatabaseConfig_preferredMaintenanceWindow(rName, preferredMaintenanceWindow string) string {
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
  bundle_id                    = "micro_2_0"
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
  bundle_id                = "micro_2_0"
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
  bundle_id                = "micro_2_0"
  backup_retention_enabled = %[2]t
  apply_immediately        = true
  skip_final_snapshot      = true
}
`, rName, backupRetentionEnabled))
}

func testAccDatabaseConfig_finalSnapshotName(rName, sName string) string {
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
  bundle_id                = "micro_2_0"
  final_snapshot_name      = %[2]q
}
`, rName, sName))
}

func testAccDatabaseConfig_tags1(rName, tagKey1, tagValue1 string) string {
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
  bundle_id                = "micro_2_0"
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
  bundle_id                = "micro_2_0"
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
  bundle_id                = "micro_ha_2_0"
  skip_final_snapshot      = true
}
`, rName))
}
