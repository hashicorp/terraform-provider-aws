// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAuditManagerControlDataSource_standard(t *testing.T) {
	// Standard controls are managed by AWS and will exist in the account automatically
	// once AuditManager is enabled.
	ctx := acctest.Context(t)
	name := "1. Risk Management"
	dataSourceName := "data.aws_auditmanager_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccControlDataSourceConfig_standard(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(dataSourceName, "control_mapping_sources.#", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccAuditManagerControlDataSource_custom(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_auditmanager_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccControlDataSourceConfig_custom(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(dataSourceName, "control_mapping_sources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "control_mapping_sources.0.source_name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "control_mapping_sources.0.source_set_up_option", string(types.SourceSetUpOptionProceduralControlsMapping)),
					resource.TestCheckResourceAttr(dataSourceName, "control_mapping_sources.0.source_type", string(types.SourceTypeManual)),
				),
			},
		},
	})
}

func testAccControlDataSourceConfig_standard(rName string) string {
	return fmt.Sprintf(`
data "aws_auditmanager_control" "test" {
  name = %[1]q
  type = "Standard"
}
`, rName)
}

func testAccControlDataSourceConfig_custom(rName string) string {
	return fmt.Sprintf(`
resource "aws_auditmanager_control" "test" {
  name = %[1]q

  control_mapping_sources {
    source_name          = %[1]q
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }
}

data "aws_auditmanager_control" "test" {
  name = aws_auditmanager_control.test.name
  type = "Custom"
}
`, rName)
}
