// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccImageBuilderInfrastructureConfigurationsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_infrastructure_configurations.test"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "arns.0", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccInfrastructureConfigurationsDataSourceConfig_filter(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  name                  = %[1]q
  instance_profile_name = aws_iam_instance_profile.test.name
}

data "aws_imagebuilder_infrastructure_configurations" "test" {
  filter {
    name   = "name"
    values = [aws_imagebuilder_infrastructure_configuration.test.name]
  }
}
`, rName)
}
