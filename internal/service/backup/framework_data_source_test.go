// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccFrameworkDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_backup_framework.test"
	resourceName := "aws_backup_framework.test"
	rName := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccFrameworkPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "control.#", resourceName, "control.#"),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "control.*", map[string]string{
						names.AttrName:            "BACKUP_RECOVERY_POINT_MINIMUM_RETENTION_CHECK",
						"input_parameter.#":       acctest.Ct1,
						"input_parameter.0.name":  "requiredRetentionDays",
						"input_parameter.0.value": "35",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "control.*", map[string]string{
						names.AttrName:      "BACKUP_PLAN_MIN_FREQUENCY_AND_MIN_RETENTION_CHECK",
						"input_parameter.#": acctest.Ct3,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "control.*", map[string]string{
						names.AttrName: "BACKUP_RECOVERY_POINT_ENCRYPTED",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "control.*", map[string]string{
						names.AttrName:                        "BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN",
						"scope.#":                             acctest.Ct1,
						"scope.0.compliance_resource_ids.#":   acctest.Ct1,
						"scope.0.compliance_resource_types.#": acctest.Ct1,
						"scope.0.compliance_resource_types.0": "EBS",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "control.*", map[string]string{
						names.AttrName: "BACKUP_RECOVERY_POINT_MANUAL_DELETION_DISABLED",
					}),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "control.*.scope.0.compliance_resource_ids.0", "aws_ebs_volume.test", names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCreationTime, resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrPair(datasourceName, "deployment_status", resourceName, "deployment_status"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccFrameworkDataSource_controlScopeTag(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_backup_framework.test"
	resourceName := "aws_backup_framework.test"
	rName := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccFrameworkPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkDataSourceConfig_controlScopeTag(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "control.#", resourceName, "control.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "control.0.name", resourceName, "control.0.name"),
					resource.TestCheckResourceAttrPair(datasourceName, "control.0.scope.#", resourceName, "control.0.scope.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "control.0.scope.0.tags.%", resourceName, "control.0.scope.0.tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "control.0.scope.0.tags.Name", resourceName, "control.0.scope.0.tags.Name"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCreationTime, resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrPair(datasourceName, "deployment_status", resourceName, "deployment_status"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccFrameworkDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 1
}

resource "aws_backup_framework" "test" {
  name        = %[1]q
  description = "Example framework"

  control {
    name = "BACKUP_RECOVERY_POINT_MINIMUM_RETENTION_CHECK"

    input_parameter {
      name  = "requiredRetentionDays"
      value = "35"
    }
  }

  control {
    name = "BACKUP_PLAN_MIN_FREQUENCY_AND_MIN_RETENTION_CHECK"

    input_parameter {
      name  = "requiredFrequencyUnit"
      value = "hours"
    }

    input_parameter {
      name  = "requiredRetentionDays"
      value = "35"
    }

    input_parameter {
      name  = "requiredFrequencyValue"
      value = "1"
    }
  }

  control {
    name = "BACKUP_RECOVERY_POINT_ENCRYPTED"
  }

  control {
    name = "BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN"

    scope {
      compliance_resource_ids = [
        aws_ebs_volume.test.id
      ]

      compliance_resource_types = [
        "EBS"
      ]
    }
  }

  control {
    name = "BACKUP_RECOVERY_POINT_MANUAL_DELETION_DISABLED"
  }

  tags = {
    "Name" = "Test Framework"
  }
}

data "aws_backup_framework" "test" {
  name = aws_backup_framework.test.name
}
`, rName)
}

func testAccFrameworkDataSourceConfig_controlScopeTag(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_framework" "test" {
  name        = %[1]q
  description = "Example framework"

  control {
    name = "BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN"

    scope {
      tags = {
        "Name" = "Example"
      }
    }
  }

  tags = {
    "Name" = "Test Framework"
  }
}

data "aws_backup_framework" "test" {
  name = aws_backup_framework.test.name
}
`, rName)
}
