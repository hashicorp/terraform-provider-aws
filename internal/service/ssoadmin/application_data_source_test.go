// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminApplicationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ssoadmin_application.test"
	applicationResourceName := "aws_ssoadmin_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig_basic(rName, testAccApplicationProviderARN),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "application_arn", applicationResourceName, "application_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "application_provider_arn", applicationResourceName, "application_provider_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_arn", applicationResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, applicationResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "portal_options.#", applicationResourceName, "portal_options.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, applicationResourceName, names.AttrStatus),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfig_basic(rName, applicationProviderARN string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_basic(rName, applicationProviderARN),
		`
data "aws_ssoadmin_application" "test" {
  application_arn = aws_ssoadmin_application.test.application_arn
}
`)
}
