// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminInstancesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssoadmin_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "identity_store_ids.#", acctest.Ct1),
					acctest.MatchResourceAttrGlobalARNNoAccount(dataSourceName, "arns.0", "sso", regexache.MustCompile("instance/(sso)?ins-[0-9A-Za-z.-]{16}")),
					resource.TestMatchResourceAttr(dataSourceName, "identity_store_ids.0", regexache.MustCompile("^[0-9A-Za-z-]*")),
				),
			},
		},
	})
}

const testAccInstancesDataSourceConfig_basic = `data "aws_ssoadmin_instances" "test" {}`
