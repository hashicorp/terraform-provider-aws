// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountAccessApplicationDataSource_byInstance(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_accountaccess_application.test"
	resourceName := "aws_accountaccess_application.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationDataSource/by_instance/"),
				ConfigVariables: config.Variables{},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_center_application_arn", resourceName, "identity_center_application_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_center_instance_arn", resourceName, "identity_center_instance_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tenant_id", resourceName, "tenant_id"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.test", names.AttrValue),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccAccountAccessApplicationDataSource_byARN(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_accountaccess_application.test"
	resourceName := "aws_accountaccess_application.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationDataSource/by_arn/"),
				ConfigVariables: config.Variables{},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_center_application_arn", resourceName, "identity_center_application_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_center_instance_arn", resourceName, "identity_center_instance_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tenant_id", resourceName, "tenant_id"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.test", names.AttrValue),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated_at"),
				),
			},
		},
	})
}
