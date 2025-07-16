//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
	"testing"
)

func TestAccGiVersionsListDataSource_basicX9M(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	dataSourceName := "data.aws_odb_gi_versions_list.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGiVersionsListeConfig_basic("Exadata.X9M"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "gi_versions.#", "2"),
				),
			},
		},
	})

}

func TestAccGiVersionsListDataSource_basicX11M(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	dataSourceName := "data.aws_odb_gi_versions_list.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGiVersionsListeConfig_basic("Exadata.X11M"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "gi_versions.#", "2"),
				),
			},
		},
	})
}

func testAccGiVersionsListeConfig_basic(shape string) string {
	return fmt.Sprintf(`

data "aws_odb_gi_versions_list" "test" {
  shape             = %[1]q
}
`, shape)
}
