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
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftimestreaminfluxdb "github.com/hashicorp/terraform-provider-aws/internal/service/timestreaminfluxdb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamInfluxDBDBInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					// Verification of read-only attributes and default values.
					// DB instance will not be publicly accessible and will not have an endpoint.
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "timestream-influxdb", regexache.MustCompile(`db-instance/+.`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, tftimestreaminfluxdb.DefaultBucketValue),
					resource.TestCheckResourceAttr(resourceName, "db_storage_type", string(awstypes.DbStorageTypeInfluxIoIncludedT1)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.DeploymentTypeSingleAz)),
					resource.TestCheckResourceAttrSet(resourceName, "influx_auth_parameters_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "organization", tftimestreaminfluxdb.DefaultOrganizationValue),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StatusAvailable)),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, tftimestreaminfluxdb.DefaultUsernameValue),
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

func TestAccTimestreamInfluxDBDBInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tftimestreaminfluxdb.ResourceDBInstance, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBInstance_logDeliveryConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1, dbInstance2 timestreaminfluxdb.GetDbInstanceOutput
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
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_logDeliveryConfigurationEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccDBInstanceConfig_logDeliveryConfigurationNotEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance2),
					testAccCheckDBInstanceNotRecreated(&dbInstance1, &dbInstance2),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.enabled", acctest.CtFalse),
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

func TestAccTimestreamInfluxDBDBInstance_publiclyAccessible(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_publiclyAccessible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
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

func TestAccTimestreamInfluxDBDBInstance_deploymentTypeMultiAzStandby(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_deploymentTypeMultiAzStandby(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					// DB instance will not be publicly accessible and will not have an endpoint.
					// DB instance will have a secondary availability zone.
					resource.TestCheckResourceAttrSet(resourceName, "secondary_availability_zone"),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.DeploymentTypeWithMultiazStandby)),
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

func TestAccTimestreamInfluxDBDBInstance_username(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance timestreaminfluxdb.GetDbInstanceOutput
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
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_username(rName, testUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, testUsername),
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

func TestAccTimestreamInfluxDBDBInstance_bucket(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance timestreaminfluxdb.GetDbInstanceOutput
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
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_bucket(rName, testBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, testBucketName),
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

func TestAccTimestreamInfluxDBDBInstance_organization(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance timestreaminfluxdb.GetDbInstanceOutput
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
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_organization(rName, testOrganizationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "organization", testOrganizationName),
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

func TestAccTimestreamInfluxDBDBInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1, dbInstance2, dbInstance3 timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				Config: testAccDBInstanceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance2),
					testAccCheckDBInstanceNotRecreated(&dbInstance2, &dbInstance2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
				),
			},
			{
				Config: testAccDBInstanceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance3),
					testAccCheckDBInstanceNotRecreated(&dbInstance2, &dbInstance3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
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

func TestAccTimestreamInfluxDBDBInstance_dbInstanceType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_dbInstanceTypeLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.large"),
				),
			},
			{
				Config: testAccDBInstanceConfig_dbInstanceTypeXLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.xlarge"),
				),
			},
			{
				Config: testAccDBInstanceConfig_dbInstanceType2XLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.2xlarge"),
				),
			},
			{
				Config: testAccDBInstanceConfig_dbInstanceType4XLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.4xlarge"),
				),
			},
			{
				Config: testAccDBInstanceConfig_dbInstanceType8XLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.8xlarge"),
				),
			},
			{
				Config: testAccDBInstanceConfig_dbInstanceType12XLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.12xlarge"),
				),
			},
			{
				Config: testAccDBInstanceConfig_dbInstanceType16XLarge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", "db.influx.16xlarge"),
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

func TestAccTimestreamInfluxDBDBInstance_dbStorageType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance timestreaminfluxdb.GetDbInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig_dbStorageTypeT2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "db_storage_type", string(awstypes.DbStorageTypeInfluxIoIncludedT2)),
				),
			},
			{
				Config: testAccDBInstanceConfig_dbStorageTypeT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "db_storage_type", string(awstypes.DbStorageTypeInfluxIoIncludedT3)),
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

func testAccCheckDBInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
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
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingDestroyed, tftimestreaminfluxdb.ResNameDBInstance, rs.Primary.ID, err)
			}

			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingDestroyed, tftimestreaminfluxdb.ResNameDBInstance, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDBInstanceExists(ctx context.Context, name string, dbInstance *timestreaminfluxdb.GetDbInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDBInstance, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDBInstance, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamInfluxDBClient(ctx)
		resp, err := conn.GetDbInstance(ctx, &timestreaminfluxdb.GetDbInstanceInput{
			Identifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDBInstance, rs.Primary.ID, err)
		}

		*dbInstance = *resp

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

func testAccCheckDBInstanceNotRecreated(before, after *timestreaminfluxdb.GetDbInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingNotRecreated, tftimestreaminfluxdb.ResNameDBInstance, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccDBInstanceConfig_base() string {
	return `
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
`
}

// Minimal configuration.
func testAccDBInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q
}
`, rName))
}

// Configuration with log_delivery_configuration set and enabled.
func testAccDBInstanceConfig_logDeliveryConfigurationEnabled(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_s3_bucket" "test_s3_bucket" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "allow_timestreaminfluxdb" {
  statement {
    actions = ["s3:PutObject"]
    principals {
      type        = "Service"
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
  allocated_storage      = 20
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  publicly_accessible    = false
  name                   = %[1]q

  log_delivery_configuration {
    s3_configuration {
      bucket_name = %[1]q
      enabled     = true
    }
  }
}
`, rName))
}

// Configuration with log_delivery_configuration set but not enabled.
func testAccDBInstanceConfig_logDeliveryConfigurationNotEnabled(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_s3_bucket" "test_s3_bucket" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "allow_timestreaminfluxdb" {
  statement {
    actions = ["s3:PutObject"]
    principals {
      type        = "Service"
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
  allocated_storage      = 20
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  publicly_accessible    = false
  name                   = %[1]q

  log_delivery_configuration {
    s3_configuration {
      bucket_name = %[1]q
      enabled     = false
    }
  }
}
`, rName))
}

// Configuration that is publicly accessible. An endpoint will be created
// for the DB instance but no inbound rules will be defined, preventing access.
func testAccDBInstanceConfig_publiclyAccessible(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_internet_gateway" "test_internet_gateway" {
  vpc_id = aws_vpc.test_vpc.id
}

resource "aws_route" "test_route" {
  route_table_id         = aws_vpc.test_vpc.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test_internet_gateway.id
}

resource "aws_route_table_association" "test_route_table_association" {
  subnet_id      = aws_subnet.test_subnet.id
  route_table_id = aws_vpc.test_vpc.main_route_table_id
}

resource "aws_vpc_security_group_ingress_rule" "test_vpc_security_group_ingress_rule_vpc" {
  security_group_id            = aws_security_group.test_security_group.id
  referenced_security_group_id = aws_security_group.test_security_group.id
  ip_protocol                  = -1
}

resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q

  publicly_accessible = true
}
`, rName))
}

func testAccDBInstanceConfig_deploymentTypeMultiAzStandby(rName string, regionName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test_vpc" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test_subnet_1" {
  vpc_id            = aws_vpc.test_vpc.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "%[2]sa"
}

resource "aws_subnet" "test_subnet_2" {
  vpc_id            = aws_vpc.test_vpc.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "%[2]sb"
}

resource "aws_security_group" "test_security_group" {
  vpc_id = aws_vpc.test_vpc.id
}

resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = [aws_subnet.test_subnet_1.id, aws_subnet.test_subnet_2.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q

  deployment_type = "WITH_MULTIAZ_STANDBY"
}
`, rName, regionName)
}

func testAccDBInstanceConfig_username(rName, username string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q

  username = %[2]q
}
`, rName, username))
}

func testAccDBInstanceConfig_bucket(rName, bucketName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q

  bucket = %[2]q
}
`, rName, bucketName))
}

func testAccDBInstanceConfig_organization(rName, organizationName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q

  organization = %[2]q
}
`, rName, organizationName))
}

func testAccDBInstanceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDBInstanceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDBInstanceConfig_dbStorageTypeT2(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q

  allocated_storage = 400
  db_storage_type   = "InfluxIOIncludedT2"
}
`, rName))
}

func testAccDBInstanceConfig_dbStorageTypeT3(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q

  allocated_storage = 400
  db_storage_type   = "InfluxIOIncludedT3"
}
`, rName))
}

func testAccDBInstanceConfig_dbInstanceTypeLarge(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  db_storage_type        = "InfluxIOIncludedT1"
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  name                   = %[1]q

  db_instance_type = "db.influx.large"
}
`, rName))
}

func testAccDBInstanceConfig_dbInstanceTypeXLarge(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  db_storage_type        = "InfluxIOIncludedT1"
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  name                   = %[1]q

  db_instance_type = "db.influx.xlarge"
}
`, rName))
}

func testAccDBInstanceConfig_dbInstanceType2XLarge(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  db_storage_type        = "InfluxIOIncludedT1"
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  name                   = %[1]q

  allocated_storage = 40
  db_instance_type  = "db.influx.2xlarge"
}
`, rName))
}

func testAccDBInstanceConfig_dbInstanceType4XLarge(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  db_storage_type        = "InfluxIOIncludedT1"
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  name                   = %[1]q

  db_instance_type = "db.influx.4xlarge"
}
`, rName))
}

func testAccDBInstanceConfig_dbInstanceType8XLarge(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  db_storage_type        = "InfluxIOIncludedT1"
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  name                   = %[1]q

  db_instance_type = "db.influx.8xlarge"
}
`, rName))
}

func testAccDBInstanceConfig_dbInstanceType12XLarge(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  db_storage_type        = "InfluxIOIncludedT1"
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  name                   = %[1]q

  db_instance_type = "db.influx.12xlarge"
}
`, rName))
}

func testAccDBInstanceConfig_dbInstanceType16XLarge(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  allocated_storage      = 20
  db_storage_type        = "InfluxIOIncludedT1"
  password               = "testpassword"
  vpc_subnet_ids         = [aws_subnet.test_subnet.id]
  vpc_security_group_ids = [aws_security_group.test_security_group.id]
  name                   = %[1]q

  db_instance_type = "db.influx.16xlarge"
}
`, rName))
}
