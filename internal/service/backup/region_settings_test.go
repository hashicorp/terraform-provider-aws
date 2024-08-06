// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupRegionSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var settings backup.DescribeRegionSettingsOutput
	resourceName := "aws_backup_region_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionSettingsConfig_1(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegionSettingsExists(ctx, &settings),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.%", "16"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Aurora", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.DocumentDB", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.DynamoDB", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EBS", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EC2", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EFS", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.FSx", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Neptune", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.RDS", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.S3", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Storage Gateway", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.VirtualMachine", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_management_preference.%", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "resource_type_management_preference.DynamoDB"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_type_management_preference.EFS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegionSettingsConfig_2(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegionSettingsExists(ctx, &settings),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.%", "16"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Aurora", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.DocumentDB", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.DynamoDB", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EBS", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EC2", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EFS", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.FSx", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Neptune", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.RDS", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.S3", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Storage Gateway", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.VirtualMachine", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_management_preference.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "resource_type_management_preference.DynamoDB", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_management_preference.EFS", acctest.CtTrue),
				),
			},
			{
				Config: testAccRegionSettingsConfig_3(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegionSettingsExists(ctx, &settings),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.%", "16"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Aurora", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.DocumentDB", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.DynamoDB", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EBS", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EC2", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EFS", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.FSx", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Neptune", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.RDS", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.S3", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Storage Gateway", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.VirtualMachine", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "resource_type_management_preference.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "resource_type_management_preference.DynamoDB", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "resource_type_management_preference.EFS", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckRegionSettingsExists(ctx context.Context, v *backup.DescribeRegionSettingsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		output, err := conn.DescribeRegionSettings(ctx, &backup.DescribeRegionSettingsInput{})

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRegionSettingsConfig_1() string {
	return `
resource "aws_backup_region_settings" "test" {
  resource_type_opt_in_preference = {
    "Aurora"                 = true
    "CloudFormation"         = true
    "DocumentDB"             = true
    "DynamoDB"               = true
    "EBS"                    = true
    "EC2"                    = true
    "EFS"                    = true
    "FSx"                    = true
    "Neptune"                = true
    "RDS"                    = true
    "Redshift"               = true
    "S3"                     = true
    "SAP HANA on Amazon EC2" = true
    "Storage Gateway"        = true
    "Timestream"             = true
    "VirtualMachine"         = true
  }
}
`
}

func testAccRegionSettingsConfig_2() string {
	return `
resource "aws_backup_region_settings" "test" {
  resource_type_opt_in_preference = {
    "Aurora"                 = false
    "CloudFormation"         = true
    "DocumentDB"             = true
    "DynamoDB"               = true
    "EBS"                    = true
    "EC2"                    = true
    "EFS"                    = true
    "FSx"                    = true
    "Neptune"                = true
    "RDS"                    = true
    "Redshift"               = true
    "S3"                     = true
    "SAP HANA on Amazon EC2" = true
    "Storage Gateway"        = true
    "Timestream"             = true
    "VirtualMachine"         = true
  }

  resource_type_management_preference = {
    "DynamoDB" = true
    "EFS"      = true
  }
}
`
}

func testAccRegionSettingsConfig_3() string {
	return `
resource "aws_backup_region_settings" "test" {
  resource_type_opt_in_preference = {
    "Aurora"                 = false
    "CloudFormation"         = true
    "DocumentDB"             = true
    "DynamoDB"               = true
    "EBS"                    = true
    "EC2"                    = true
    "EFS"                    = true
    "FSx"                    = true
    "Neptune"                = true
    "RDS"                    = true
    "Redshift"               = true
    "S3"                     = true
    "SAP HANA on Amazon EC2" = true
    "Storage Gateway"        = true
    "Timestream"             = true
    "VirtualMachine"         = false
  }

  resource_type_management_preference = {
    "DynamoDB" = false
    "EFS"      = true
  }
}
`
}
