// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticBeanstalkSolutionStackDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_elastic_beanstalk_solution_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSolutionStackDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrName, regexache.MustCompile("^64bit Amazon Linux (.*) running Python (.*)$")),
				),
			},
		},
	})
}

const testAccSolutionStackDataSourceConfig_basic = `
data "aws_elastic_beanstalk_solution_stack" "test" {
  most_recent = true

  # e.g. "64bit Amazon Linux 2018.03 v2.10.14 running Python 3.6"
  name_regex = "^64bit Amazon Linux (.*) running Python (.*)$"
}
`
