// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAuditManagerFrameworkDataSource_standard(t *testing.T) {
	// Standard frameworks are managed by AWS and will exist in the account automatically
	// once AuditManager is enabled.
	ctx := acctest.Context(t)
	name := "Essential Eight"
	dataSourceName := "data.aws_auditmanager_framework.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkDataSourceConfig_standard(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(dataSourceName, "control_sets.#", "8"),
				),
			},
		},
	})
}

func TestAccAuditManagerFrameworkDataSource_custom(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_auditmanager_framework.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkDataSourceConfig_custom(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(dataSourceName, "control_sets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "control_sets.0.name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "control_sets.0.controls.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccFrameworkDataSourceConfig_standard(rName string) string {
	return fmt.Sprintf(`
data "aws_auditmanager_framework" "test" {
  name           = %[1]q
  framework_type = "Standard"
}
`, rName)
}

func testAccFrameworkDataSourceConfig_custom(rName string) string {
	return fmt.Sprintf(`
resource "aws_auditmanager_control" "test" {
  name = %[1]q

  control_mapping_sources {
    source_name          = %[1]q
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }
}

resource "aws_auditmanager_framework" "test" {
  name = %[1]q

  control_sets {
    name = %[1]q
    controls {
      id = aws_auditmanager_control.test.id
    }
  }
}

data "aws_auditmanager_framework" "test" {
  name           = aws_auditmanager_framework.test.name
  framework_type = "Custom"
}
`, rName)
}
