// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccElasticBeanstalkApplicationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_elastic_beanstalk_application.test"
	resourceName := "aws_elastic_beanstalk_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "appversion_lifecycle.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "appversion_lifecycle.0.delete_source_from_s3", dataSourceResourceName, "appversion_lifecycle.0.delete_source_from_s3"),
					resource.TestCheckResourceAttrPair(resourceName, "appversion_lifecycle.0.max_age_in_days", dataSourceResourceName, "appversion_lifecycle.0.max_age_in_days"),
					resource.TestCheckResourceAttrPair(resourceName, "appversion_lifecycle.0.service_role", dataSourceResourceName, "appversion_lifecycle.0.service_role"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceResourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceResourceName, "name"),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_maxAge(rName), `
data "aws_elastic_beanstalk_application" "test" {
  name = aws_elastic_beanstalk_application.test.name
}
`)
}
