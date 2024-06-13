// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tftimestreaminfluxdb "github.com/hashicorp/terraform-provider-aws/internal/service/timestreaminfluxdb"
)

func TestAccTimestreamInfluxDBDbInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbinstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDbInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					// Verification of read-only attributes and default values.
					// DB instance will not be publicly accessible and will not have an endpoint.
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "timestream-influxdb", regexache.MustCompile(`db-instance/+.`)),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttr(resourceName, "bucket", tftimestreaminfluxdb.DefaultBucketValue),
					resource.TestCheckResourceAttr(resourceName, "db_storage_type", string(awstypes.DbStorageTypeInfluxIoIncludedT1)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.DeploymentTypeSingleAz)),
					resource.TestCheckResourceAttrSet(resourceName, "influx_auth_parameters_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "organization", tftimestreaminfluxdb.DefaultOrganizationValue),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StatusAvailable)),
					resource.TestCheckResourceAttr(resourceName, "username", tftimestreaminfluxdb.DefaultUsernameValue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDbInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbinstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDbInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tftimestreaminfluxdb.ResourceDbInstance, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTimestreamInfluxDBDbInstance_logDeliveryConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbinstance timestreaminfluxdb.GetDbInstanceOutput
	// The same random name will be used for both the DB instance and the log S3 bucket name.
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDbInstanceConfig_logDeliveryConfigurationEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.enabled", "true"),
				),
			},
			{
				Config: testAccDbInstanceConfig_logDeliveryConfigurationNotEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.enabled", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDbInstance_publiclyAccessible(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbinstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDbInstanceConfig_publiclyAccessible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDbInstance_deploymentTypeMultiAzStandby(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbinstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDbInstanceConfig_deploymentTypeMultiAzStandby(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					// DB instance will not be publicly accessible and will not have an endpoint.
					// DB instance will have a secondary availability zone.
					resource.TestCheckResourceAttrSet(resourceName, "secondary_availability_zone"),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.DeploymentTypeWithMultiazStandby)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDbInstance_username(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbinstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"
	testUsername := "testusername"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDbInstanceConfig_username(rName, testUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "username", testUsername),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDbInstance_bucket(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbinstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"
	testBucketName := "testbucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDbInstanceConfig_bucket(rName, testBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "bucket", testBucketName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDbInstance_organization(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbinstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"
	testOrganizationName := "testorganization"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDbInstanceConfig_organization(rName, testOrganizationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "organization", testOrganizationName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDbInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbinstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDbInstanceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				Config: testAccDbInstanceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
				),
			},
			{
				Config: testAccDbInstanceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDbInstance_dbInstanceType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbinstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDbInstanceConfig_dbInstanceTypeLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.large"),
				),
			},
			{
				Config: testAccDbInstanceConfig_dbInstanceTypeXLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.xlarge"),
				),
			},
			{
				Config: testAccDbInstanceConfig_dbInstanceType2XLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.2xlarge"),
				),
			},
			{
				Config: testAccDbInstanceConfig_dbInstanceType4XLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.4xlarge"),
				),
			},
			{
				Config: testAccDbInstanceConfig_dbInstanceType8XLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.8xlarge"),
				),
			},
			{
				Config: testAccDbInstanceConfig_dbInstanceType12XLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.12xlarge"),
				),
			},
			{
				Config: testAccDbInstanceConfig_dbInstanceType16XLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbInstanceExists(ctx, resourceName, &dbinstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.16xlarge"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func testAccCheckDbInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamInfluxDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_timestreaminfluxdb_db_instance" {
				continue
			}

			input := &timestreaminfluxdb.GetDbInstanceInput{
				Identifier: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetDbInstance(ctx, input)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingDestroyed, tftimestreaminfluxdb.ResNameDbInstance, rs.Primary.ID, err)
			}

			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingDestroyed, tftimestreaminfluxdb.ResNameDbInstance, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDbInstanceExists(ctx context.Context, name string, dbinstance *timestreaminfluxdb.GetDbInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDbInstance, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDbInstance, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamInfluxDBClient(ctx)
		resp, err := conn.GetDbInstance(ctx, &timestreaminfluxdb.GetDbInstanceInput{
			Identifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDbInstance, rs.Primary.ID, err)
		}

		*dbinstance = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamInfluxDBClient(ctx)

	input := &timestreaminfluxdb.ListDbInstancesInput{}
	_, err := conn.ListDbInstances(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckDbInstanceNotRecreated(before, after *timestreaminfluxdb.GetDbInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingNotRecreated, tftimestreaminfluxdb.ResNameDbInstance, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccDbInstanceConfig_base() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test_vpc" {
	cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test_subnet" {
  vpc_id     = aws_vpc.test_vpc.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_security_group" "test_security_group" {
	vpc_id = aws_vpc.test_vpc.id
}
`)
}

// Minimal configuration.
func testAccDbInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	name = %[1]q
}
`, rName))
}

// Configuration with log_delivery_configuration set and enabled.
func testAccDbInstanceConfig_logDeliveryConfigurationEnabled(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_s3_bucket" "test_s3_bucket" {
	bucket = %[1]q
	force_destroy = true
}

data "aws_iam_policy_document" "allow_timestreaminfluxdb" {
	statement {
		actions = ["s3:PutObject"]
		principals {
			type = "Service"
			identifiers = ["timestream-influxdb.amazonaws.com"]
		}
		resources = [
			"${aws_s3_bucket.test_s3_bucket.arn}/*"
		]
	}
}

resource "aws_s3_bucket_policy" "allow_timestreaminfluxdb" {
	bucket = aws_s3_bucket.test_s3_bucket.id
	policy = data.aws_iam_policy_document.allow_timestreaminfluxdb.json
}

resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	publicly_accessible = false
	name = %[1]q

	log_delivery_configuration {
		s3_configuration {
			bucket_name = %[1]q
			enabled = true
		}
	}
}
`, rName))
}

// Configuration with log_delivery_configuration set but not enabled.
func testAccDbInstanceConfig_logDeliveryConfigurationNotEnabled(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_s3_bucket" "test_s3_bucket" {
	bucket = %[1]q
	force_destroy = true
}

data "aws_iam_policy_document" "allow_timestreaminfluxdb" {
	statement {
		actions = ["s3:PutObject"]
		principals {
			type = "Service"
			identifiers = ["timestream-influxdb.amazonaws.com"]
		}
		resources = [
			"${aws_s3_bucket.test_s3_bucket.arn}/*"
		]
	}
}

resource "aws_s3_bucket_policy" "allow_timestreaminfluxdb" {
	bucket = aws_s3_bucket.test_s3_bucket.id
	policy = data.aws_iam_policy_document.allow_timestreaminfluxdb.json
}

resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	publicly_accessible = false
	name = %[1]q

	log_delivery_configuration {
		s3_configuration {
			bucket_name = %[1]q
			enabled = false
		}
	}
}
`, rName))
}

// Configuration that is publicly accessible. An endpoint will be created
// for the DB instance but no inbound rules will be defined, preventing access.
func testAccDbInstanceConfig_publiclyAccessible(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_internet_gateway" "test_internet_gateway" {
  vpc_id = aws_vpc.test_vpc.id
}

resource "aws_route" "test_route" {
	route_table_id = aws_vpc.test_vpc.main_route_table_id
	destination_cidr_block = "0.0.0.0/0"
	gateway_id = aws_internet_gateway.test_internet_gateway.id
}

resource "aws_route_table_association" "test_route_table_association" {
  subnet_id      = aws_subnet.test_subnet.id
  route_table_id = aws_vpc.test_vpc.main_route_table_id
}

resource "aws_vpc_security_group_ingress_rule" "test_vpc_security_group_ingress_rule_vpc" {
	security_group_id = aws_security_group.test_security_group.id
	referenced_security_group_id = aws_security_group.test_security_group.id
	ip_protocol       = -1
}

resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	db_storage_type = "InfluxIOIncludedT1"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	name = %[1]q

	publicly_accessible = true
}
`, rName))
}

func testAccDbInstanceConfig_deploymentTypeMultiAzStandby(rName string, regionName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test_vpc" {
	cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test_subnet_1" {
  vpc_id     = aws_vpc.test_vpc.id
  cidr_block = "10.0.1.0/24"
  availability_zone = "%[2]sa"
}

resource "aws_subnet" "test_subnet_2" {
  vpc_id     = aws_vpc.test_vpc.id
  cidr_block = "10.0.2.0/24"
  availability_zone = "%[2]sb"
}

resource "aws_security_group" "test_security_group" {
	vpc_id = aws_vpc.test_vpc.id
}

resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	db_storage_type = "InfluxIOIncludedT1"
	vpc_subnet_ids = [aws_subnet.test_subnet_1.id, aws_subnet.test_subnet_2.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	name = %[1]q

	deployment_type = "WITH_MULTIAZ_STANDBY"
}
`, rName, regionName)
}

func testAccDbInstanceConfig_username(rName, username string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	db_storage_type = "InfluxIOIncludedT1"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	name = %[1]q

	username = %[2]q
}
`, rName, username))
}

func testAccDbInstanceConfig_bucket(rName, bucketName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	db_storage_type = "InfluxIOIncludedT1"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	name = %[1]q

	bucket = %[2]q
}
`, rName, bucketName))
}

func testAccDbInstanceConfig_organization(rName, organizationName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	db_storage_type = "InfluxIOIncludedT1"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	name = %[1]q

	organization = %[2]q
}
`, rName, organizationName))
}

func testAccDbInstanceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	db_storage_type = "InfluxIOIncludedT1"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	name = %[1]q

	tags = {
		%[2]q = %[3]q
	}
}
`, rName, tagKey1, tagValue1))
}

func testAccDbInstanceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	db_storage_type = "InfluxIOIncludedT1"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	name = %[1]q

	tags = {
		%[2]q = %[3]q
		%[4]q = %[5]q
	}
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDbInstanceConfig_dbStorageTypeT2(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	name = %[1]q

	db_storage_type = "InfluxIOIncludedT2"
}
`, rName))
}

func testAccDbInstanceConfig_dbStorageTypeT3(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	db_instance_type = "db.influx.medium"
	name = %[1]q

	db_storage_type = "InfluxIOIncludedT3"
}
`, rName))
}

func testAccDbInstanceConfig_dbInstanceTypeLarge(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	db_storage_type = "InfluxIOIncludedT1"
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	name = %[1]q

	db_instance_type = "db.influx.large"
}
`, rName))
}

func testAccDbInstanceConfig_dbInstanceTypeXLarge(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	db_storage_type = "InfluxIOIncludedT1"
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	name = %[1]q

	db_instance_type = "db.influx.xlarge"
}
`, rName))
}

func testAccDbInstanceConfig_dbInstanceType2XLarge(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	db_storage_type = "InfluxIOIncludedT1"
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	name = %[1]q

	allocated_storage = 40
	db_instance_type = "db.influx.2xlarge"
}
`, rName))
}

func testAccDbInstanceConfig_dbInstanceType4XLarge(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	db_storage_type = "InfluxIOIncludedT1"
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	name = %[1]q

	db_instance_type = "db.influx.4xlarge"
}
`, rName))
}

func testAccDbInstanceConfig_dbInstanceType8XLarge(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	db_storage_type = "InfluxIOIncludedT1"
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	name = %[1]q

	db_instance_type = "db.influx.8xlarge"
}
`, rName))
}

func testAccDbInstanceConfig_dbInstanceType12XLarge(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	db_storage_type = "InfluxIOIncludedT1"
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	name = %[1]q

	db_instance_type = "db.influx.12xlarge"
}
`, rName))
}

func testAccDbInstanceConfig_dbInstanceType16XLarge(rName string) string {
	return acctest.ConfigCompose(testAccDbInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
	allocated_storage = 20
	db_storage_type = "InfluxIOIncludedT1"
	password = "testpassword"
	vpc_subnet_ids = [aws_subnet.test_subnet.id]
	vpc_security_group_ids = [aws_security_group.test_security_group.id]
	name = %[1]q

	db_instance_type = "db.influx.16xlarge"
}
`, rName))
}
