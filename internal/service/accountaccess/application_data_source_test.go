// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountAccessApplicationDataSource_byInstance(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_accountaccess_application.test"
	resourceName := "aws_accountaccess_application.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

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
				Config: testAccApplicationDataSourceConfig_byInstance(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_center_application_arn", resourceName, "identity_center_application_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_center_instance_arn", resourceName, "identity_center_instance_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

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
				Config: testAccApplicationDataSourceConfig_byARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "tenant_id", resourceName, "tenant_id"),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfig_byInstance(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_basic(rName), `
data "aws_accountaccess_application" "test" {
  identity_center_instance_arn = aws_accountaccess_application.test.identity_center_instance_arn
}
`)
}

func testAccApplicationDataSourceConfig_byARN(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_basic(rName), `
data "aws_accountaccess_application" "test" {
  arn = aws_accountaccess_application.test.arn
}
`)
}
