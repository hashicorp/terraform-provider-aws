// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2ApplicationIntegrationURLDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_wafv2_application_integration_url.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: `data "aws_wafv2_application_integration_url" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrHasPrefix(dataSourceName, names.AttrURL, "https://"),
				),
			},
		},
	})
}
