// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectLocationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dsResourceName := "data.aws_dx_location.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsResourceName, "available_macsec_port_speeds.#"),
					resource.TestCheckResourceAttrSet(dsResourceName, "available_port_speeds.#"),
					resource.TestCheckResourceAttrSet(dsResourceName, "available_providers.#"),
					resource.TestCheckResourceAttrSet(dsResourceName, "location_code"),
					resource.TestCheckResourceAttrSet(dsResourceName, "location_name"),
				),
			},
		},
	})
}

const testAccLocationDataSourceConfig_basic = `
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
}

data "aws_dx_location" "test" {
  location_code = local.location_codes[length(local.location_codes) - 1]
}
`
