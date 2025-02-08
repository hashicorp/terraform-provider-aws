// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSActivityStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbInstance types.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_activity_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, names.StandardPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivityStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivityStreamConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityStreamExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "engine_native_audit_fields_included", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "kinesis_stream_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccActivityStreamConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityStreamExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "engine_native_audit_fields_included", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "kinesis_stream_name"),
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

func TestAccRDSActivityStream_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dbInstance types.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_activity_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, names.StandardPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivityStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivityStreamConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityStreamExists(ctx, resourceName, &dbInstance),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceActivityStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckActivityStreamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_activity_stream" {
				continue
			}

			_, err := tfrds.FindDBInstanceWithActivityStream(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DB Activity Stream %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckActivityStreamExists(ctx context.Context, n string, v *types.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBInstanceWithActivityStream(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccActivityStreamConfig_baseVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName))
}

func testAccActivityStreamConfig_DBInstanceMSQQLBase(rName string) string {
	return acctest.ConfigCompose(
		testAccActivityStreamConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
resource "aws_db_instance" "test" {
  allocated_storage       = 20
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = "sqlserver-se"
  engine_version          = "15.00"
  identifier              = %[1]q
  instance_class          = "db.m6i.large"
  license_model           = "license-included"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  storage_encrypted       = true
  username                = "tfacctest"
}
`, rName))
}

func testAccActivityStreamConfig_basic(rName string, engineNativeAuditFieldsIncluded bool) string {
	return acctest.ConfigCompose(testAccActivityStreamConfig_DBInstanceMSQQLBase(rName),
		fmt.Sprintf(`
resource "aws_db_activity_stream" "test" {
  resource_arn                        = aws_db_instance.test.arn
  kms_key_id                          = aws_kms_key.test.key_id
  mode                                = "async"
  engine_native_audit_fields_included = %[1]t
}
`, engineNativeAuditFieldsIncluded))
}
