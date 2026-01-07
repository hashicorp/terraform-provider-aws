// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccODBDBSystemShapesListDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	dataSourceName := "data.aws_odb_db_system_shapes.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: basicConfigDBSystemShapeDataSource("use1-az6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "db_system_shapes.#", "2"),
				),
			},
		},
	})
}

func basicConfigDBSystemShapeDataSource(availabilityZoneId string) string {
	return fmt.Sprintf(`
data "aws_odb_db_system_shapes" "test"{
  availability_zone_id = %[1]q
}
`, availabilityZoneId)
}
