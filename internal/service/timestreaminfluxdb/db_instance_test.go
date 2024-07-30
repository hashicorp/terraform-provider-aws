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
	tftimestreaminfluxdb "github.com/hashicorp/terraform-provider-aws/internal/service/timestreaminfluxdb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "timestream-influxdb", regexache.MustCompile(`db-instance/+.`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttr(resourceName, "db_storage_type", string(awstypes.DbStorageTypeInfluxIoIncludedT1)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.DeploymentTypeSingleAz)),
					resource.TestCheckResourceAttrSet(resourceName, "influx_auth_parameters_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bucket", "username", "password", "organization"},
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
				Config: testAccDBInstanceConfig_logDeliveryConfigurationEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bucket", "username", "password", "organization"},
			},
			{
				Config: testAccDBInstanceConfig_logDeliveryConfigurationEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance2),
					testAccCheckDBInstanceNotRecreated(&dbInstance1, &dbInstance2),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.enabled", acctest.CtFalse),
				),
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
				Config: testAccDBInstanceConfig_deploymentTypeMultiAzStandby(rName),
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

//func TestAccTimestreamInfluxDBDBInstance_tags(t *testing.T) {
//	ctx := acctest.Context(t)
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//	var dbInstance1, dbInstance2, dbInstance3 timestreaminfluxdb.GetDbInstanceOutput
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_timestreaminfluxdb_db_instance.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			testAccPreCheck(ctx, t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
//		Steps: []resource.TestStep{
//			{
//				Config: testAccDBInstanceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance1),
//					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
//					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
//					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
//					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
//				),
//			},
//			{
//				Config: testAccDBInstanceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance2),
//					testAccCheckDBInstanceNotRecreated(&dbInstance2, &dbInstance2),
//					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
//					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
//					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
//					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
//					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1Updated),
//					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
//				),
//			},
//			{
//				Config: testAccDBInstanceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance3),
//					testAccCheckDBInstanceNotRecreated(&dbInstance2, &dbInstance3),
//					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
//					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
//					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
//					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
//				),
//			},
//			{
//				ResourceName:      resourceName,
//				ImportState:       true,
//				ImportStateVerify: true,
//			},
//		},
//	})
//}

func testAccCheckDBInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamInfluxDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_timestreaminfluxdb_db_instance" {
				continue
			}

			_, err := tftimestreaminfluxdb.FindDBInstanceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
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
		resp, err := tftimestreaminfluxdb.FindDBInstanceByID(ctx, conn, rs.Primary.ID)

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

func testAccDBInstanceConfig_base(rName string, subnetCount int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, subnetCount), `
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}
`)
}

// Minimal configuration.
func testAccDBInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(rName, 1), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test.*.id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"
}
`, rName))
}

// Configuration with log_delivery_configuration set and enabled.
func testAccDBInstanceConfig_logDeliveryConfigurationEnabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(rName, 1), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
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
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}

resource "aws_s3_bucket_policy" "allow_timestreaminfluxdb" {
  bucket = aws_s3_bucket.test_s3_bucket.id
  policy = data.aws_iam_policy_document.allow_timestreaminfluxdb.json
}

resource "aws_timestreaminfluxdb_db_instance" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test.*.id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  publicly_accessible    = false
  bucket                 = "initial"
  organization           = "organization"

  log_delivery_configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.bucket
      enabled     = %[2]t
    }
  }
}
`, rName, enabled))
}

func testAccDBInstanceConfig_publiclyAccessible(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(rName, 1), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route" "test" {
  route_table_id         = aws_vpc.test.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_vpc.test.main_route_table_id
}

resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id            = aws_security_group.test.id
  referenced_security_group_id = aws_security_group.test.id
  ip_protocol                  = -1
}

resource "aws_timestreaminfluxdb_db_instance" "test" {
  name                   = %[1]q
  allocated_storage      = 20
username = "admin"
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = aws_subnet.test.*.id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"

  publicly_accessible = true
}
`, rName))
}

func testAccDBInstanceConfig_deploymentTypeMultiAzStandby(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  name                   = %[1]q
  allocated_storage      = 20
username = "admin"
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = aws_subnet.test.*.id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"

  deployment_type = "WITH_MULTIAZ_STANDBY"
}
`, rName))
}

func testAccDBInstanceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(rName, 1), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(rName, 1), fmt.Sprintf(`
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
